package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PlannedOperation struct {
	ID           uuid.UUID       `db:"id"`
	Name         string          `db:"name"`
	AccountID    uuid.UUID       `db:"account_id"`
	Amount       decimal.Decimal `db:"amount"`
	Type         OperationType   `db:"type"`
	IntervalType IntervalType    `db:"interval_type"`
	Interval     int             `db:"interval"`
	CategoryID   uuid.UUID       `db:"category_id"`
	PlannedAt    time.Time       `db:"planned_at"`
	NextRunAt    time.Time       `db:"next_run_at"`
	IsRecurring  bool            `db:"is_recurring"`
}
