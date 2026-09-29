package dto

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AccountPost struct {
	Name     string          `json:"name"`
	Currency string          `json:"currency"`
	Icon     string          `json:"icon"`
	Balance  decimal.Decimal `json:"balance"`
}

func (r *AccountPost) Validate() error {
	if r.Name == "" {
		return errors.New("name")
	}
	if r.Icon == "" {
		return errors.New("icon")
	}
	if r.Currency == "" {
		return errors.New("currency")
	}
	if !ValidCurrency(r.Currency) {
		return errors.New("currency")
	}
	if !r.Balance.IsPositive() {
		return errors.New("balance")
	}
	return nil
}

type AccountUpdate struct {
	Name string `json:"name,omitempty"`
	Icon string `json:"icon,omitempty"`
}

type AccountResponseShort struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	Currency string    `json:"currency" db:"currency"`
	Icon     string    `json:"icon" db:"icon"`
}

type AccountBalanceUpdate struct {
	ID    uuid.UUID
	Delta decimal.Decimal
}

type AccountResponse struct {
	ID       uuid.UUID       `json:"id" db:"id"`
	Name     string          `json:"name" db:"name"`
	Currency string          `json:"currency" db:"currency"`
	Icon     string          `json:"icon" db:"icon"`
	Balance  decimal.Decimal `json:"balance" db:"balance"`
	UserID   uuid.UUID       `json:"-" db:"user_id"`
}
