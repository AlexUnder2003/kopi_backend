package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type IntervalType string

const (
	IntervalTypeDaily    IntervalType = "daily"
	IntervalTypeWeekly   IntervalType = "weekly"
	IntervalTypeBiweekly IntervalType = "biweekly"
	IntervalTypeMonthly  IntervalType = "monthly"
	IntervalTypeCustom   IntervalType = "custom"
)

type Budget struct {
	ID           uuid.UUID       `db:"id"`
	Name         string          `db:"name"`
	Amount       decimal.Decimal `db:"amount"`
	Balance      decimal.Decimal `db:"balance"`
	UserID       uuid.UUID       `db:"user_id"`
	Currency     string          `db:"currency"`
	IntervalType IntervalType    `db:"interval_type"`
	Interval     int             `db:"interval"`
	StartDate    time.Time       `db:"start_date"`
	ResetDate    time.Time       `db:"reset_date"`
	CategoryID   uuid.UUID       `db:"category_id"`
	IsActive     *bool           `db:"is_active"`
}
