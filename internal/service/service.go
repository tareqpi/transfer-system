package service

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tareqpi/transfer-system/internal/domain"
	"github.com/tareqpi/transfer-system/internal/repository"
)

var (
	ErrSameSourceAndDestination = errors.New("source and destination account IDs cannot be the same")
	ErrNonPositiveAmount        = errors.New("amount should be greater than zero")
	ErrInvalidAccountIDs        = errors.New("invalid account IDs")
	ErrInsufficientBalance      = errors.New("insufficient balance")
)

type Service interface {
	CreateAccount(ctx context.Context, newAccount domain.Account) (*domain.Account, error)
	GetAccount(ctx context.Context, accountID string) (*domain.Account, error)
	TransferMoney(ctx context.Context, transaction domain.Transaction) error
}

type DefaultService struct {
	repository repository.Repository
}

func NewService(dataRepository repository.Repository) Service {
	return &DefaultService{repository: dataRepository}
}

func (s DefaultService) CreateAccount(ctx context.Context, newAccount domain.Account) (*domain.Account, error) {
	account, err := s.repository.CreateAccount(ctx, newAccount)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s DefaultService) GetAccount(ctx context.Context, accountID string) (*domain.Account, error) {
	account, err := s.repository.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s DefaultService) TransferMoney(ctx context.Context, transaction domain.Transaction) error {
	if transaction.SourceAccountID == transaction.DestinationAccountID {
		return ErrSameSourceAndDestination
	}
	if transaction.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrNonPositiveAmount
	}
	if transaction.SourceAccountID <= 0 || transaction.DestinationAccountID <= 0 {
		return ErrInvalidAccountIDs
	}

	_, err := s.repository.TransferMoney(ctx, transaction)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInsufficientBalance):
			return ErrInsufficientBalance
		}
	}
	return err
}
