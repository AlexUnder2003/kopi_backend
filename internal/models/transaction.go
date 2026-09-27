package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeTransfer TransactionType = "transfer"
)

type Transaction struct {
	ID             uuid.UUID       `db:"id"`
	Name           string          `db:"name"`
	Type           TransactionType `db:"type"`
	AccountID      uuid.UUID       `db:"account_id"`
	FromAccountID  uuid.UUID       `db:"from_account_id"`
	CategoryID     uuid.UUID       `db:"category_id"`
	Amount         decimal.Decimal `db:"amount"`
	OccurrenceDate time.Time       `db:"occurrence_date"`
}
