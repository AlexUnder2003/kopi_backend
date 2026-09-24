package dto

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

type BudgetPost struct {
	Amount     decimal.Decimal `json:"amount"`
	Frequency  BudgetFrequency `json:"frequency"`
	CategoryID uuid.UUID       `json:"category_id"`
	Currency   string          `json:"currency"`
}

type BudgetUpdate struct {
	Amount    decimal.Decimal `json:"amount,omitempty"`
	Frequency BudgetFrequency `json:"frequency,omitempty"`
	Currency  string          `json:"currency,omitempty"`
}

type BudgetResponse struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	TotalAmount decimal.Decimal  `json:"total_amount" db:"total_amount"`
	AmountSpent decimal.Decimal  `json:"amount_spent" db:"amount_spent"`
	AmountLeft  decimal.Decimal  `json:"amount_left" db:"amount_left"`
	Frequency   BudgetFrequency  `json:"frequency" db:"frequency"`
	Category    CategoryResponse `json:"category" db:"category"`
	Currency    string           `json:"currency" db:"currency"`
}
