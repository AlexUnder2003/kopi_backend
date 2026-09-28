package dto

import (
	"errors"
	"slices"
	"time"

	"KopiBackend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BudgetPost struct {
	Amount       decimal.Decimal     `json:"amount"`
	IntervalType models.IntervalType `json:"interval_type"`
	Interval     int                 `json:"interval,omitempty"`
	CategoryID   uuid.UUID           `json:"category_id"`
	Currency     string              `json:"currency"`
	StartDate    time.Time           `json:"start_date"`
}

func (r *BudgetPost) Validate() error {
	if !r.Amount.IsPositive() {
		return errors.New("amount")
	}
	if !validIntervalType(r.IntervalType) {
		return errors.New("interval_type")
	}
	if r.IntervalType == models.IntervalTypeCustom && r.Interval < 1 {
		return errors.New("interval")
	}
	if r.CategoryID == uuid.Nil {
		return errors.New("category_id")
	}
	if !validBudgetCurrency(r.Currency) {
		return errors.New("currency")
	}
	if r.StartDate.IsZero() {
		return errors.New("start_date")
	}
	return nil
}

type BudgetUpdate struct {
	Amount     decimal.Decimal `json:"amount,omitempty"`
	CategoryID uuid.UUID       `json:"category_id,omitempty"`
	StartDate  time.Time       `json:"start_date,omitempty"`
}

func (r *BudgetUpdate) Validate() error {
	if r.Amount.IsNegative() {
		return errors.New("amount")
	}
	return nil
}

func validIntervalType(intervalType models.IntervalType) bool {
	switch intervalType {
	case models.IntervalTypeDaily, models.IntervalTypeWeekly, models.IntervalTypeBiweekly, models.IntervalTypeMonthly, models.IntervalTypeCustom:
		return true
	default:
		return false
	}
}

func validBudgetCurrency(currency string) bool {
	return slices.Contains(models.CurrencyCodes, models.CurrencyCode(currency))
}

type BudgetResponse struct {
	ID           uuid.UUID           `json:"id" db:"id"`
	Amount       decimal.Decimal     `json:"amount" db:"amount"`
	Balance      decimal.Decimal     `json:"balance" db:"balance"`
	IntervalType models.IntervalType `json:"interval_type" db:"interval_type"`
	Interval     int                 `json:"interval" db:"interval"`
	Currency     string              `json:"currency" db:"currency"`
	ResetDate    time.Time           `json:"reset_date" db:"reset_date"`
	Category     CategoryResponse    `json:"category" db:"category"`
}
