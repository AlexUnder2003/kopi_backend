package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TransactionTypeIncome      TransactionType = "income"
	TransactionTypeExpense     TransactionType = "expense"
	TransactionTypeTransferIn  TransactionType = "transfer_in"
	TransactionTypeTransferOut TransactionType = "transfer_out"
)

type TransactionPost struct {
	Name           string          `json:"name"`
	Amount         decimal.Decimal `json:"amount"`
	Type           TransactionType `json:"type"`
	AccountID      uuid.UUID       `json:"account_id,omitempty"`
	CategoryID     uuid.UUID       `json:"category_id"`
	OccurrenceDate time.Time       `json:"occurrence_date"`
}

type TransactionTransferPost struct {
	Name           string          `json:"name"`
	Amount         decimal.Decimal `json:"amount"`
	FromAccountID  uuid.UUID       `json:"from_account_id"`
	ToAccountID    uuid.UUID       `json:"to_account_id"`
	OccurrenceDate time.Time       `json:"occurrence_date"`
}

type TransactionUpdate struct {
	Name           string          `json:"name,omitempty"`
	Amount         decimal.Decimal `json:"amount,omitempty"`
	Type           TransactionType `json:"type,omitempty"`
	AccountID      uuid.UUID       `json:"account_id,omitempty"`
	CategoryID     uuid.UUID       `json:"category_id,omitempty"`
	OccurrenceDate time.Time       `json:"occurrence_date,omitempty"`
}

type TransactionTransferUpdate struct {
	Name           string          `json:"name,omitempty"`
	Amount         decimal.Decimal `json:"amount,omitempty"`
	FromAccountID  uuid.UUID       `json:"from_account_id,omitempty"`
	ToAccountID    uuid.UUID       `json:"to_account_id,omitempty"`
	OccurrenceDate time.Time       `json:"occurrence_date,omitempty"`
}

type TransactionResponse struct {
	ID             uuid.UUID            `json:"id" db:"id"`
	Name           string               `json:"name" db:"name"`
	Amount         decimal.Decimal      `json:"amount" db:"amount"`
	Type           TransactionType      `json:"type" db:"type"`
	Account        AccountResponseShort `json:"account,omitempty" db:"account"`
	FromAccount    AccountResponseShort `json:"from_account,omitempty" db:"-"`
	ToAccount      AccountResponseShort `json:"to_account,omitempty" db:"-"`
	Category       CategoryResponse     `json:"category" db:"category"`
	OccurrenceDate time.Time            `json:"occurrence_date" db:"occurrence_date"`
}
