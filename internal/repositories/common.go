package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}

func nullDecimal(d decimal.Decimal) any {
	if d.IsZero() {
		return nil
	}
	return d
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func nullTimePtr(t *time.Time) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return *t
}
