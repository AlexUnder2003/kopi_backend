package dto

import (
	"errors"
	"time"

	"KopiBackend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PlannedOperationPost struct {
	Name         string                 `json:"name"`
	AccountID    uuid.UUID              `json:"account_id"`
	Amount       decimal.Decimal        `json:"amount"`
	Type         models.TransactionType `json:"type"`
	IntervalType models.IntervalType    `json:"interval_type"`
	Interval     int                    `json:"interval"`
	CategoryID   uuid.UUID              `json:"category_id"`
	PlannedAt    time.Time              `json:"planned_date"`
	IsRecurring  bool                   `json:"is_recurring"`
}

func (r *PlannedOperationPost) Validate() error {
	if r.Name == "" {
		return errors.New("name")
	}
	if !r.Amount.IsPositive() {
		return errors.New("amount")
	}
	if !validTransactionType(r.Type) {
		return errors.New("type")
	}
	if !validIntervalType(r.IntervalType) {
		return errors.New("interval_type")
	}
	if r.AccountID == uuid.Nil {
		return errors.New("account_id")
	}
	if r.CategoryID == uuid.Nil {
		return errors.New("category_id")
	}
	if r.PlannedAt.IsZero() {
		return errors.New("planned_date")
	}
	return nil
}

type PlannedOperationUpdate struct {
	Name         string                 `json:"name,omitempty"`
	AccountID    uuid.UUID              `json:"account_id,omitempty"`
	Amount       decimal.Decimal        `json:"amount,omitempty"`
	Type         models.TransactionType `json:"type,omitempty"`
	IntervalType models.IntervalType    `json:"interval_type,omitempty"`
	Interval     int                    `json:"interval,omitempty"`
	CategoryID   uuid.UUID              `json:"category_id,omitempty"`
	PlannedAt    time.Time              `json:"planned_date,omitempty"`
	IsRecurring  *bool                  `json:"is_recurring,omitempty"`
}

func (r *PlannedOperationUpdate) Validate() error {
	if r.Type != "" && !validTransactionType(r.Type) {
		return errors.New("type")
	}
	if r.IntervalType != "" && !validIntervalType(r.IntervalType) {
		return errors.New("interval_type")
	}
	if r.Amount.IsNegative() {
		return errors.New("amount")
	}
	return nil
}

type PlannedOperationResponse struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	Account      AccountResponseShort   `json:"account" db:"account"`
	Amount       decimal.Decimal        `json:"amount" db:"amount"`
	Type         models.TransactionType `json:"type" db:"type"`
	IntervalType models.IntervalType    `json:"interval_type" db:"interval_type"`
	Interval     int                    `json:"interval" db:"interval"`
	Category     CategoryResponse       `json:"category" db:"category"`
	PlannedAt    time.Time              `json:"planned_date" db:"planned_at"`
	NextRunAt    time.Time              `json:"next_run_date" db:"next_run_at"`
	IsRecurring  bool                   `json:"is_recurring" db:"is_recurring"`
}
