package domain

import "github.com/shopspring/decimal"

type Transaction struct {
	ID                   int64           `db:"id" json:"transaction_id"`
	SourceAccountID      int64           `db:"source_account_id" json:"source_account_id"`
	DestinationAccountID int64           `db:"destination_account_id" json:"destination_account_id"`
	Amount               decimal.Decimal `db:"amount" json:"amount"`
}
