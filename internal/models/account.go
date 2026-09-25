package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CurrencyCode string

const (
	CurrencyCodeUSD CurrencyCode = "USD"
	CurrencyCodeEUR CurrencyCode = "EUR"
	CurrencyCodeGBP CurrencyCode = "RUB"
	CurrencyCodeJPY CurrencyCode = "KZT"
	CurrencyCodeBYN CurrencyCode = "BYN"
)

var CurrencyCodes = []CurrencyCode{
	CurrencyCodeUSD,
	CurrencyCodeEUR,
	CurrencyCodeGBP,
	CurrencyCodeJPY,
	CurrencyCodeBYN,
}

type Account struct {
	ID       uuid.UUID       `db:"id"`
	Name     string          `db:"name"`
	Currency CurrencyCode    `db:"currency"`
	Icon     string          `db:"icon"`
	Balance  decimal.Decimal `db:"balance"`
	UserID   uuid.UUID       `db:"user_id"`
}
