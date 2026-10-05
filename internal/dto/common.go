package dto

import (
	"KopiBackend/internal/models"
	"slices"
)

func ValidCurrency(currency string) bool {
	return slices.Contains(models.CurrencyCodes, models.CurrencyCode(currency))
}

func validIntervalType(intervalType models.IntervalType) bool {
	switch intervalType {
	case models.IntervalTypeDaily, models.IntervalTypeWeekly, models.IntervalTypeBiweekly, models.IntervalTypeMonthly, models.IntervalTypeCustom:
		return true
	default:
		return false
	}
}

func validOperationType(opType models.OperationType) bool {
	switch opType {
	case models.OperationTypeIncome, models.OperationTypeExpense, models.OperationTypeTransfer:
		return true
	default:
		return false
	}
}
