package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CurrencyCode string

const (
	CurrencyCodeUSD CurrencyCode = "USD"
	CurrencyCodeEUR CurrencyCode = "EUR"
	CurrencyCodeRUB CurrencyCode = "RUB"
	CurrencyCodeKZT CurrencyCode = "KZT"
	CurrencyCodeBYN CurrencyCode = "BYN"
)

var CurrencyCodes = []CurrencyCode{
	CurrencyCodeUSD,
	CurrencyCodeEUR,
	CurrencyCodeRUB,
	CurrencyCodeKZT,
	CurrencyCodeBYN,
}

type Account struct {
	ID                   uuid.UUID       `db:"id"`
	Name                 string          `db:"name"`
	Currency             CurrencyCode    `db:"currency"`
	Icon                 string          `db:"icon"`
	Balance              decimal.Decimal `db:"balance"`
	UserID               uuid.UUID       `db:"user_id"`
	IncludeInFreeBalance *bool           `db:"include_in_free_balance"`
}
