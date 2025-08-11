package repository

import (
	"context"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/tareqpi/transfer-system/internal/config"
	"github.com/tareqpi/transfer-system/internal/domain"
	"github.com/tareqpi/transfer-system/internal/logger"
	"go.uber.org/zap"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

type Repository interface {
	CreateAccount(ctx context.Context, account domain.Account) (*domain.Account, error)
	GetAccount(ctx context.Context, id string) (*domain.Account, error)
	TransferMoney(ctx context.Context, transaction domain.Transaction) (*domain.Transaction, error)
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(databasePool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: databasePool}
}

func Setup(ctx context.Context) (*pgxpool.Pool, error) {
	databaseURL := config.Get().DatabaseURL

	if databaseURL == "" {
		return nil, errors.New("databaseURL is empty")
	}

	if err := migrateDB(databaseURL); err != nil {
		return nil, err
	}

	pool, err := initPool(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func migrateDB(databaseURL string) error {
	migration, err := migrate.New("file://../../migrations", databaseURL)
	if err != nil {
		logger.L().Error("migration setup failed", zap.Error(err))
		return err
	}
	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		logger.L().Error("migration up failed", zap.Error(err))
		return err
	}
	logger.L().Info("database migrations applied successfully")
	return nil
}

func initPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func (r *PGRepository) CreateAccount(ctx context.Context, account domain.Account) (*domain.Account, error) {
	var (
		id      int64
		balance decimal.Decimal
	)

	const insertSQL = `
        INSERT INTO accounts.accounts (id, balance)
        VALUES ($1, $2)
        RETURNING id, balance
    `

	if err := r.pool.QueryRow(ctx, insertSQL, account.ID, account.Balance).Scan(&id, &balance); err != nil {
		return nil, err
	}

	created := &domain.Account{ID: id, Balance: balance}
	return created, nil
}

func (r *PGRepository) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	var account domain.Account

	const selectSQL = `
        SELECT id, balance
        FROM accounts.accounts
        WHERE id = $1
    `

	if err := r.pool.QueryRow(ctx, selectSQL, id).Scan(&account.ID, &account.Balance); err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *PGRepository) TransferMoney(ctx context.Context, transaction domain.Transaction) (*domain.Transaction, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var sourceBalance decimal.Decimal
	var tmpID int64

	if transaction.SourceAccountID <= transaction.DestinationAccountID {
		if err = tx.QueryRow(ctx, `SELECT id, balance FROM accounts.accounts WHERE id = $1 FOR UPDATE`, transaction.SourceAccountID).Scan(&tmpID, &sourceBalance); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New("source account not found")
			}
			return nil, err
		}
		if err = tx.QueryRow(ctx, `SELECT id FROM accounts.accounts WHERE id = $1 FOR UPDATE`, transaction.DestinationAccountID).Scan(&tmpID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New("destination account not found")
			}
			return nil, err
		}
	} else {
		if err = tx.QueryRow(ctx, `SELECT id FROM accounts.accounts WHERE id = $1 FOR UPDATE`, transaction.DestinationAccountID).Scan(&tmpID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New("destination account not found")
			}
			return nil, err
		}
		if err = tx.QueryRow(ctx, `SELECT id, balance FROM accounts.accounts WHERE id = $1 FOR UPDATE`, transaction.SourceAccountID).Scan(&tmpID, &sourceBalance); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New("source account not found")
			}
			return nil, err
		}
	}

	if sourceBalance.LessThan(transaction.Amount) {
		return nil, ErrInsufficientBalance
	}

	if _, err = tx.Exec(ctx, `UPDATE accounts.accounts SET balance = balance - $1 WHERE id = $2`, transaction.Amount, transaction.SourceAccountID); err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, `UPDATE accounts.accounts SET balance = balance + $1 WHERE id = $2`, transaction.Amount, transaction.DestinationAccountID); err != nil {
		return nil, err
	}

	var id int64
	if err = tx.QueryRow(ctx, `
        INSERT INTO accounts.transactions (source_account_id, destination_account_id, amount)
        VALUES ($1, $2, $3)
        RETURNING id
    `, transaction.SourceAccountID, transaction.DestinationAccountID, transaction.Amount).Scan(&id); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &domain.Transaction{ID: id, SourceAccountID: transaction.SourceAccountID, DestinationAccountID: transaction.DestinationAccountID, Amount: transaction.Amount}, nil
}
