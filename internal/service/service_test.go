package service

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/tareqpi/transfer-system/internal/domain"
	"github.com/tareqpi/transfer-system/internal/repository"
)

type mockRepository struct {
	createAccountFn func(ctx context.Context, account domain.Account) (*domain.Account, error)
	getAccountFn    func(ctx context.Context, id string) (*domain.Account, error)
	transferMoneyFn func(ctx context.Context, tx domain.Transaction) (*domain.Transaction, error)

	createAccountCalls int
	getAccountCalls    int
	transferMoneyCalls int

	lastTransferTx domain.Transaction
}

func (m *mockRepository) CreateAccount(ctx context.Context, account domain.Account) (*domain.Account, error) {
	m.createAccountCalls++
	if m.createAccountFn != nil {
		return m.createAccountFn(ctx, account)
	}
	return nil, nil
}

func (m *mockRepository) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	m.getAccountCalls++
	if m.getAccountFn != nil {
		return m.getAccountFn(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) TransferMoney(ctx context.Context, tx domain.Transaction) (*domain.Transaction, error) {
	m.transferMoneyCalls++
	m.lastTransferTx = tx
	if m.transferMoneyFn != nil {
		return m.transferMoneyFn(ctx, tx)
	}
	return &tx, nil
}

func TestDefaultService_CreateAccount_Success(t *testing.T) {
	t.Parallel()

	mockRepo := &mockRepository{
		createAccountFn: func(ctx context.Context, account domain.Account) (*domain.Account, error) {
			created := account
			return &created, nil
		},
	}
	svc := NewService(mockRepo)

	account := domain.Account{ID: 1, Balance: decimal.NewFromInt(100)}
	got, err := svc.CreateAccount(context.Background(), account)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != account.ID || !got.Balance.Equal(account.Balance) {
		t.Fatalf("unexpected account: got=%+v want=%+v", got, account)
	}
	if mockRepo.createAccountCalls != 1 {
		t.Fatalf("CreateAccount calls: got=%d want=1", mockRepo.createAccountCalls)
	}
}

func TestDefaultService_CreateAccount_RepositoryError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("db error")
	mockRepo := &mockRepository{
		createAccountFn: func(ctx context.Context, account domain.Account) (*domain.Account, error) {
			return nil, wantErr
		},
	}
	svc := NewService(mockRepo)

	_, err := svc.CreateAccount(context.Background(), domain.Account{ID: 2, Balance: decimal.NewFromInt(50)})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error mismatch: got=%v want=%v", err, wantErr)
	}
}

func TestDefaultService_GetAccount_Success(t *testing.T) {
	t.Parallel()

	expected := &domain.Account{ID: 42, Balance: decimal.NewFromInt(250)}
	mockRepo := &mockRepository{
		getAccountFn: func(ctx context.Context, id string) (*domain.Account, error) {
			return expected, nil
		},
	}
	svc := NewService(mockRepo)

	got, err := svc.GetAccount(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != expected.ID || !got.Balance.Equal(expected.Balance) {
		t.Fatalf("unexpected account: got=%+v want=%+v", got, expected)
	}
	if mockRepo.getAccountCalls != 1 {
		t.Fatalf("GetAccount calls: got=%d want=1", mockRepo.getAccountCalls)
	}
}

func TestDefaultService_GetAccount_RepositoryError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("not found")
	mockRepo := &mockRepository{
		getAccountFn: func(ctx context.Context, id string) (*domain.Account, error) {
			return nil, wantErr
		},
	}
	svc := NewService(mockRepo)

	_, err := svc.GetAccount(context.Background(), "404")
	if !errors.Is(err, wantErr) {
		t.Fatalf("error mismatch: got=%v want=%v", err, wantErr)
	}
}

func TestDefaultService_TransferMoney_InputValidation(t *testing.T) {
	t.Parallel()

	zero := decimal.Zero
	negative := decimal.NewFromInt(-10)
	positive := decimal.NewFromInt(10)

	cases := []struct {
		name    string
		tx      domain.Transaction
		wantErr error
	}{
		{
			name:    "same source and destination",
			tx:      domain.Transaction{SourceAccountID: 1, DestinationAccountID: 1, Amount: positive},
			wantErr: ErrSameSourceAndDestination,
		},
		{
			name:    "zero amount",
			tx:      domain.Transaction{SourceAccountID: 1, DestinationAccountID: 2, Amount: zero},
			wantErr: ErrNonPositiveAmount,
		},
		{
			name:    "negative amount",
			tx:      domain.Transaction{SourceAccountID: 1, DestinationAccountID: 2, Amount: negative},
			wantErr: ErrNonPositiveAmount,
		},
		{
			name:    "non-positive source account id",
			tx:      domain.Transaction{SourceAccountID: 0, DestinationAccountID: 2, Amount: positive},
			wantErr: ErrInvalidAccountIDs,
		},
		{
			name:    "non-positive destination account id",
			tx:      domain.Transaction{SourceAccountID: 1, DestinationAccountID: -2, Amount: positive},
			wantErr: ErrInvalidAccountIDs,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := &mockRepository{}
			svc := NewService(mockRepo)

			err := svc.TransferMoney(context.Background(), tc.tx)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error mismatch: got=%v want=%v", err, tc.wantErr)
			}
			if mockRepo.transferMoneyCalls != 0 {
				t.Fatalf("TransferMoney should not hit repository on validation error; calls=%d", mockRepo.transferMoneyCalls)
			}
		})
	}
}

func TestDefaultService_TransferMoney_InsufficientBalance(t *testing.T) {
	t.Parallel()

	mockRepo := &mockRepository{
		transferMoneyFn: func(ctx context.Context, tx domain.Transaction) (*domain.Transaction, error) {
			return nil, repository.ErrInsufficientBalance
		},
	}
	svc := NewService(mockRepo)

	tx := domain.Transaction{SourceAccountID: 1, DestinationAccountID: 2, Amount: decimal.NewFromInt(10)}
	err := svc.TransferMoney(context.Background(), tx)
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("error mismatch: got=%v want=%v", err, ErrInsufficientBalance)
	}
	if mockRepo.transferMoneyCalls != 1 {
		t.Fatalf("TransferMoney calls: got=%d want=1", mockRepo.transferMoneyCalls)
	}
}

func TestDefaultService_TransferMoney_RepositoryError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("db failure")
	mockRepo := &mockRepository{
		transferMoneyFn: func(ctx context.Context, tx domain.Transaction) (*domain.Transaction, error) {
			return nil, wantErr
		},
	}
	svc := NewService(mockRepo)

	tx := domain.Transaction{SourceAccountID: 1, DestinationAccountID: 2, Amount: decimal.NewFromInt(10)}
	err := svc.TransferMoney(context.Background(), tx)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error mismatch: got=%v want=%v", err, wantErr)
	}
}

func TestDefaultService_TransferMoney_Success(t *testing.T) {
	t.Parallel()

	mockRepo := &mockRepository{
		transferMoneyFn: func(ctx context.Context, tx domain.Transaction) (*domain.Transaction, error) {
			return &tx, nil
		},
	}
	svc := NewService(mockRepo)

	tx := domain.Transaction{SourceAccountID: 10, DestinationAccountID: 20, Amount: decimal.NewFromInt(99)}
	err := svc.TransferMoney(context.Background(), tx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mockRepo.transferMoneyCalls != 1 {
		t.Fatalf("TransferMoney calls: got=%d want=1", mockRepo.transferMoneyCalls)
	}
	if mockRepo.lastTransferTx.SourceAccountID != tx.SourceAccountID ||
		mockRepo.lastTransferTx.DestinationAccountID != tx.DestinationAccountID ||
		!mockRepo.lastTransferTx.Amount.Equal(tx.Amount) {
		t.Fatalf("transaction mismatch: got=%+v want=%+v", mockRepo.lastTransferTx, tx)
	}
}
