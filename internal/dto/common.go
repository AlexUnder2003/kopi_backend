package dto

import (
	"KopiBackend/internal/models"
	"slices"
)

func validCurrency(currency string) bool {
	return slices.Contains(models.CurrencyCodes, models.CurrencyCode(currency))
}
