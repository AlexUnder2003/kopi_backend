package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	appdto "KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type parsed struct {
	err error
}

func (p *parsed) string(raw string) string {
	return strings.TrimSpace(raw)
}

func (p *parsed) requiredString(raw, field string) string {
	value := p.string(raw)
	if value == "" && p.err == nil {
		p.err = fmt.Errorf("%s is required", field)
	}
	return value
}

func (p *parsed) requiredUUID(raw, field string) uuid.UUID {
	if p.err != nil {
		return uuid.Nil
	}
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		p.err = fmt.Errorf("%s must be a UUID", field)
		return uuid.Nil
	}
	return id
}

func (p *parsed) optionalUUID(raw, field string) uuid.UUID {
	if strings.TrimSpace(raw) == "" {
		return uuid.Nil
	}
	return p.requiredUUID(raw, field)
}

func (p *parsed) optionalBalance(raw string) decimal.Decimal {
	raw = strings.TrimSpace(raw)
	if raw == "" || p.err != nil {
		return decimal.Decimal{}
	}
	amount, err := decimal.NewFromString(raw)
	if err != nil || amount.IsNegative() {
		p.err = errors.New("balance must be a non-negative decimal")
		return decimal.Decimal{}
	}
	return amount
}

func (p *parsed) optionalAmount(raw string) decimal.Decimal {
	raw = strings.TrimSpace(raw)
	if raw == "" || p.err != nil {
		return decimal.Decimal{}
	}
	amount, err := decimal.NewFromString(raw)
	if err != nil || !amount.IsPositive() {
		p.err = errors.New("amount must be a positive decimal")
		return decimal.Decimal{}
	}
	return amount
}

func (p *parsed) requiredAmount(raw string) decimal.Decimal {
	if strings.TrimSpace(raw) == "" && p.err == nil {
		p.err = errors.New("amount is required")
		return decimal.Decimal{}
	}
	return p.optionalAmount(raw)
}

func (p *parsed) optionalNonNegative(raw, field string) decimal.Decimal {
	raw = strings.TrimSpace(raw)
	if raw == "" || p.err != nil {
		return decimal.Decimal{}
	}
	amount, err := decimal.NewFromString(raw)
	if err != nil || amount.IsNegative() {
		p.err = fmt.Errorf("%s must be a non-negative decimal", field)
		return decimal.Decimal{}
	}
	return amount
}

func (p *parsed) requiredCurrency(raw string) string {
	value := strings.ToUpper(p.string(raw))
	if value == "" && p.err == nil {
		p.err = errors.New("currency is required")
		return ""
	}
	if p.err != nil {
		return ""
	}
	if !appdto.ValidCurrency(value) {
		p.err = errors.New("currency must be USD, EUR, RUB, KZT, or BYN")
		return ""
	}
	return value
}

func (p *parsed) optionalTime(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" || p.err != nil {
		return time.Time{}
	}
	if parsedAt, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsedAt
	}
	parsedAt, err := time.Parse("2006-01-02", raw)
	if err != nil {
		p.err = errors.New("use RFC3339 or YYYY-MM-DD")
		return time.Time{}
	}
	return parsedAt
}

func (p *parsed) optionalOperationType(raw string, allowTransfer bool) models.OperationType {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" || p.err != nil {
		return ""
	}
	opType := models.OperationType(raw)
	switch opType {
	case models.OperationTypeIncome, models.OperationTypeExpense, models.OperationTypeReservation:
		return opType
	case models.OperationTypeTransfer:
		if allowTransfer {
			return opType
		}
	}
	if allowTransfer {
		p.err = errors.New("type must be income, expense, transfer, or reservation")
	} else {
		p.err = errors.New("type must be income, expense, or reservation")
	}
	return ""
}

func (p *parsed) requiredOperationType(raw string, allowTransfer bool) models.OperationType {
	opType := p.optionalOperationType(raw, allowTransfer)
	if opType == "" && p.err == nil {
		p.err = errors.New("type is required")
	}
	return opType
}

func (p *parsed) optionalIntervalType(raw string) models.IntervalType {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" || p.err != nil {
		return ""
	}
	intervalType := models.IntervalType(raw)
	switch intervalType {
	case models.IntervalTypeDaily, models.IntervalTypeWeekly, models.IntervalTypeBiweekly, models.IntervalTypeMonthly, models.IntervalTypeCustom:
		return intervalType
	default:
		p.err = errors.New("interval_type must be daily, weekly, biweekly, monthly, or custom")
		return ""
	}
}
