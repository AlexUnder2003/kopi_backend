package models

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

type PlannedOperation struct {
	ID          uuid.UUID                 `db:"id"`
	Name        string                    `db:"name"`
	AccountID   uuid.UUID                 `db:"account_id"`
	Amount      decimal.Decimal           `db:"amount"`
	Type        TransactionType           `db:"type"`
	Frequency   PlannedOperationFrequency `db:"frequency"`
	CategoryID  uuid.UUID                 `db:"category_id"`
	PlannedAt   time.Time                 `db:"planned_at"`
	NextRunAt   *time.Time                `db:"next_run_at"`
	IsRecurring bool                      `db:"is_recurring"`
}
