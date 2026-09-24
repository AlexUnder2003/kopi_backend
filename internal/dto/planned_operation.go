package dto

import (
	"time"

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
	Type        TransactionType           `json:"type"`
	Frequency   PlannedOperationFrequency `json:"frequency"`
	CategoryID  uuid.UUID                 `json:"category_id"`
	PlannedAt   time.Time                 `json:"planned_date"`
	IsRecurring bool                      `json:"is_recurring"`
}

type PlannedOperationUpdate struct {
	Name        string                    `json:"name,omitempty"`
	AccountID   uuid.UUID                 `json:"account_id,omitempty"`
	Amount      decimal.Decimal           `json:"amount,omitempty"`
	Type        TransactionType           `json:"type,omitempty"`
	Frequency   PlannedOperationFrequency `json:"frequency,omitempty"`
	CategoryID  uuid.UUID                 `json:"category_id,omitempty"`
	PlannedAt   time.Time                 `json:"planned_date,omitempty"`
	IsRecurring bool                      `json:"is_recurring,omitempty"`
}

type PlannedOperationResponse struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	Name        string                    `json:"name" db:"name"`
	Account     AccountResponseShort      `json:"account" db:"account"`
	Amount      decimal.Decimal           `json:"amount" db:"amount"`
	Type        TransactionType           `json:"type" db:"type"`
	Frequency   PlannedOperationFrequency `json:"frequency" db:"frequency"`
	Category    CategoryResponse          `json:"category" db:"category"`
	PlannedAt   time.Time                 `json:"planned_date" db:"planned_at"`
	IsRecurring bool                      `json:"is_recurring" db:"is_recurring"`
}
