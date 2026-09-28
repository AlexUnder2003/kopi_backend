package dto

import (
	"errors"
	"time"

	"KopiBackend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PlannedOperationFrequency string

const (
	PlannedOperationFrequencyDaily   PlannedOperationFrequency = "daily"
	PlannedOperationFrequencyWeekly  PlannedOperationFrequency = "weekly"
	PlannedOperationFrequencyMonthly PlannedOperationFrequency = "monthly"
	PlannedOperationFrequencyYearly  PlannedOperationFrequency = "yearly"
)

type PlannedOperationPost struct {
	Name        string                    `json:"name"`
	AccountID   uuid.UUID                 `json:"account_id"`
	Amount      decimal.Decimal           `json:"amount"`
	Type        models.TransactionType    `json:"type"`
	Frequency   PlannedOperationFrequency `json:"frequency"`
	CategoryID  uuid.UUID                 `json:"category_id"`
	PlannedAt   time.Time                 `json:"planned_date"`
	IsRecurring bool                      `json:"is_recurring"`
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
	if !validPlannedOperationFrequency(r.Frequency) {
		return errors.New("frequency")
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
	Name        string                    `json:"name,omitempty"`
	AccountID   uuid.UUID                 `json:"account_id,omitempty"`
	Amount      decimal.Decimal           `json:"amount,omitempty"`
	Type        models.TransactionType    `json:"type,omitempty"`
	Frequency   PlannedOperationFrequency `json:"frequency,omitempty"`
	CategoryID  uuid.UUID                 `json:"category_id,omitempty"`
	PlannedAt   time.Time                 `json:"planned_date,omitempty"`
	IsRecurring *bool                     `json:"is_recurring,omitempty"`
}

func (r *PlannedOperationUpdate) Validate() error {
	if r.Type != "" && !validTransactionType(r.Type) {
		return errors.New("type")
	}
	if r.Frequency != "" && !validPlannedOperationFrequency(r.Frequency) {
		return errors.New("frequency")
	}
	if r.Amount.IsNegative() {
		return errors.New("amount")
	}
	return nil
}

func validPlannedOperationFrequency(frequency PlannedOperationFrequency) bool {
	switch frequency {
	case PlannedOperationFrequencyDaily, PlannedOperationFrequencyWeekly, PlannedOperationFrequencyMonthly, PlannedOperationFrequencyYearly:
		return true
	default:
		return false
	}
}

type PlannedOperationResponse struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	Name        string                    `json:"name" db:"name"`
	Account     AccountResponseShort      `json:"account" db:"account"`
	Amount      decimal.Decimal           `json:"amount" db:"amount"`
	Type        models.TransactionType    `json:"type" db:"type"`
	Frequency   PlannedOperationFrequency `json:"frequency" db:"frequency"`
	Category    CategoryResponse          `json:"category" db:"category"`
	PlannedAt   time.Time                 `json:"planned_date" db:"planned_at"`
	IsRecurring bool                      `json:"is_recurring" db:"is_recurring"`
}
