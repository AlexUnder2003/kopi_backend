package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BudgetFrequency string

const (
	BudgetFrequencyDaily   BudgetFrequency = "daily"
	BudgetFrequencyWeekly  BudgetFrequency = "weekly"
	BudgetFrequencyMonthly BudgetFrequency = "monthly"
	BudgetFrequencyYearly  BudgetFrequency = "yearly"
)

type Budget struct {
	ID         uuid.UUID       `db:"id"`
	Amount     decimal.Decimal `db:"amount"`
	UserID     uuid.UUID       `db:"user_id"`
	Currency   string          `db:"currency"`
	Frequency  BudgetFrequency `db:"frequency"`
	CategoryID uuid.UUID       `db:"category_id"`
}
