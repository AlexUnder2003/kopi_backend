package dto

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type Dashboard struct {
	TotalBalance     decimal.Decimal `json:"total_balance"`
	AvailableBalance decimal.Decimal `json:"available_balance"`
	TotalIncome      decimal.Decimal `json:"total_income"`
	TotalExpenses    decimal.Decimal `json:"total_expenses"`
}

type CurrencyResponse struct {
	Date  string
	Base  string
	Rates map[string]float64
}

func (r *CurrencyResponse) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if err := json.Unmarshal(raw["date"], &r.Date); err != nil {
		return err
	}
	delete(raw, "date")

	for base, rates := range raw {
		r.Base = base
		return json.Unmarshal(rates, &r.Rates)
	}
	return nil
}
