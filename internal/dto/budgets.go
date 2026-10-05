package dto

import (
	"errors"
	"time"

	"KopiBackend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BudgetPost struct {
	Name         string              `json:"name"`
	Amount       decimal.Decimal     `json:"amount"`
	IntervalType models.IntervalType `json:"interval_type"`
	Interval     int                 `json:"interval,omitempty"`
	CategoryID   uuid.UUID           `json:"category_id"`
	Currency     string              `json:"currency"`
	StartDate    time.Time           `json:"start_date,omitempty"`
}

func (r *BudgetPost) Validate() error {
	if r.Name == "" {
		return errors.New("name")
	}
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
	if !ValidCurrency(r.Currency) {
		return errors.New("currency")
	}
	if r.IntervalType == models.IntervalTypeCustom && r.StartDate.IsZero() {
		return errors.New("start_date")
	}
	return nil
}

type BudgetUpdate struct {
	Name         string              `json:"name,omitempty"`
	Amount       decimal.Decimal     `json:"amount,omitempty"`
	StartDate    time.Time           `json:"start_date,omitempty"`
	IntervalType models.IntervalType `json:"interval_type,omitempty"`
	Interval     int                 `json:"interval,omitempty"`
	IsActive     *bool               `json:"is_active,omitempty"`
}

func (r *BudgetUpdate) Validate() error {
	if r.IntervalType != "" && !validIntervalType(r.IntervalType) {
		return errors.New("interval_type")
	}
	if r.IntervalType == models.IntervalTypeCustom && r.Interval < 1 {
		return errors.New("interval")
	}
	if r.IntervalType == models.IntervalTypeCustom && r.StartDate.IsZero() {
		return errors.New("start_date")
	}
	return nil
}

type BudgetBalanceUpdate struct {
	UserID     uuid.UUID
	CategoryID uuid.UUID
	Currency   string
	Delta      decimal.Decimal
}

type BudgetResponse struct {
	ID           uuid.UUID           `json:"id" db:"id"`
	Name         string              `json:"name" db:"name"`
	Amount       decimal.Decimal     `json:"amount" db:"amount"`
	Balance      decimal.Decimal     `json:"balance" db:"balance"`
	IntervalType models.IntervalType `json:"interval_type" db:"interval_type"`
	Interval     *int                `json:"interval" db:"interval"`
	Currency     string              `json:"currency" db:"currency"`
	StartDate    time.Time           `json:"start_date" db:"start_date"`
	ResetDate    time.Time           `json:"reset_date" db:"reset_date"`
	IsActive     bool                `json:"is_active" db:"is_active"`
	Category     CategoryResponse    `json:"category" db:"category"`
}
