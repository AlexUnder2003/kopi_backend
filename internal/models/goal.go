package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Goal struct {
	ID            uuid.UUID        `db:"id"`
	UserID        uuid.UUID        `db:"user_id"`
	Name          string           `db:"name"`
	Currency      CurrencyCode     `db:"currency"`
	Deadline      time.Time        `db:"deadline"`
	TargetAmount  decimal.Decimal  `db:"target_amount"`
	CurrentAmount *decimal.Decimal `db:"current_amount"`
}
