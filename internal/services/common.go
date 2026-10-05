package services

import (
	"time"

	"KopiBackend/internal/models"
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
