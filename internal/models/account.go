package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Account struct {
	ID       uuid.UUID       `db:"id"`
	Name     string          `db:"name"`
	Currency string          `db:"currency"`
	Icon     string          `db:"icon"`
	Balance  decimal.Decimal `db:"balance"`
	UserID   uuid.UUID       `db:"user_id"`
}
