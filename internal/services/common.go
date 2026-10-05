package services

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"KopiBackend/internal/config"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/shopspring/decimal"
)

const (
	errInternalServerError = "internal_server_error"
)

func NextDate(date time.Time, intervalType models.IntervalType, interval int) time.Time {
	var next time.Time

	now := time.Now().UTC()
	date = date.UTC()

	for {
		switch intervalType {
		case models.IntervalTypeDaily:
			next = date.AddDate(0, 0, 1)
		case models.IntervalTypeWeekly:
			next = date.AddDate(0, 0, 7)
		case models.IntervalTypeBiweekly:
			next = date.AddDate(0, 0, 14)
		case models.IntervalTypeMonthly:
			next = date.AddDate(0, 1, 0)
		case models.IntervalTypeCustom:
			next = date.AddDate(0, 0, interval)
		}
		if next.After(now) {
			return next
		}
		date = next
	}
}

func effectiveTime(updated, existing time.Time) time.Time {
	if updated.IsZero() {
		return existing
	}
	return updated
}

func effectiveIntervalType(updated, existing models.IntervalType) models.IntervalType {
	if updated == "" {
		return existing
	}
	return updated
}

func effectiveInterval(updated, existing int) int {
	if updated == 0 {
		return existing
	}
	return updated
}

func GetCurrencyRates(config *config.AppConfig, base string) (map[string]decimal.Decimal, error) {
	url := config.CurrencyAPIURL + "/currencies/" + strings.ToLower(base) + ".json"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var currencyResponse dto.CurrencyResponse
	if err := json.Unmarshal(body, &currencyResponse); err != nil {
		return nil, err
	}

	rates := make(map[string]decimal.Decimal, len(currencyResponse.Rates))
	for code, rate := range currencyResponse.Rates {
		rates[code] = decimal.NewFromFloat(rate)
	}
	return rates, nil
}
