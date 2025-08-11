package domain

import "github.com/shopspring/decimal"

type Account struct {
	ID      int64           `db:"id" json:"account_id"`
	Balance decimal.Decimal `db:"balance" json:"balance"`
}
