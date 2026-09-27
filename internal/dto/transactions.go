package dto

import (
	"errors"
	"time"

	"KopiBackend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (r *TransactionPost) Validate() error {
	if r.Name == "" {
		return errors.New("name")
	}
	if !r.Amount.IsPositive() {
		return errors.New("amount")
	}
	if !validTransactionType(r.Type) {
		return errors.New("type")
	}
	if r.OccurrenceDate.IsZero() {
		return errors.New("occurrence_date")
	}
	return nil
}

func (r *TransactionUpdate) Validate() error {
	if r.Type != "" && !validTransactionType(r.Type) {
		return errors.New("type")
	}
	if r.Amount.IsNegative() {
		return errors.New("amount")
	}
	return nil
}

func validTransactionType(txType models.TransactionType) bool {
	switch txType {
	case models.TransactionTypeIncome, models.TransactionTypeExpense, models.TransactionTypeTransferIn, models.TransactionTypeTransferOut:
		return true
	default:
		return false
	}
}

type TransactionPost struct {
	Name           string                 `json:"name"`
	Amount         decimal.Decimal        `json:"amount"`
	Type           models.TransactionType `json:"type"`
	AccountID      uuid.UUID              `json:"account_id,omitempty"`
	CategoryID     uuid.UUID              `json:"category_id"`
	OccurrenceDate time.Time              `json:"occurrence_date"`
}

type TransactionTransferPost struct {
	Name           string          `json:"name"`
	Amount         decimal.Decimal `json:"amount"`
	FromAccountID  uuid.UUID       `json:"from_account_id"`
	ToAccountID    uuid.UUID       `json:"to_account_id"`
	OccurrenceDate time.Time       `json:"occurrence_date"`
}

type TransactionUpdate struct {
	Name           string                 `json:"name,omitempty"`
	Amount         decimal.Decimal        `json:"amount,omitempty"`
	Type           models.TransactionType `json:"type,omitempty"`
	AccountID      uuid.UUID              `json:"account_id,omitempty"`
	CategoryID     uuid.UUID              `json:"category_id,omitempty"`
	OccurrenceDate time.Time              `json:"occurrence_date,omitempty"`
}

type TransactionTransferUpdate struct {
	Name           string          `json:"name,omitempty"`
	Amount         decimal.Decimal `json:"amount,omitempty"`
	FromAccountID  uuid.UUID       `json:"from_account_id"`
	ToAccountID    uuid.UUID       `json:"to_account_id"`
	OccurrenceDate time.Time       `json:"occurrence_date,omitempty"`
}

type TransactionResponse struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	Name           string                 `json:"name" db:"name"`
	Amount         decimal.Decimal        `json:"amount" db:"amount"`
	Type           models.TransactionType `json:"type" db:"type"`
	Account        AccountResponseShort   `json:"account,omitempty" db:"account"`
	FromAccount    AccountResponseShort   `json:"from_account,omitempty" db:"-"`
	ToAccount      AccountResponseShort   `json:"to_account,omitempty" db:"-"`
	Category       CategoryResponse       `json:"category" db:"category"`
	OccurrenceDate time.Time              `json:"occurrence_date" db:"occurrence_date"`
}
