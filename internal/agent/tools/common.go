package tools

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"KopiBackend/internal/dto"

	"github.com/google/uuid"
	adkagent "google.golang.org/adk/v2/agent"
)

const (
	defaultListLimit = 50
	maxListLimit     = 100
)

func boundedLimit(limit int) int {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	return limit
}

func userIDFrom(ctx adkagent.Context) (uuid.UUID, error) {
	raw := strings.TrimSpace(ctx.UserID())
	if raw == "" {
		return uuid.Nil, errors.New("user is not authenticated")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user id: %w", err)
	}
	return id, nil
}

func optionalUUID(raw, field string) (uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return uuid.Nil, nil
	}
	return requiredUUID(raw, field)
}

func requiredUUID(raw, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s must be a UUID", field)
	}
	return id, nil
}

func parseOptionalTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed, nil
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, errors.New("use RFC3339 or YYYY-MM-DD")
	}
	return parsed, nil
}

func requiredCurrency(currency string) (string, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return "", errors.New("currency is required")
	}
	if !dto.ValidCurrency(currency) {
		return "", fmt.Errorf("unknown currency %q", currency)
	}
	return currency, nil
}
