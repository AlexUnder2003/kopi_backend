package models

import (
	"time"

	"github.com/google/uuid"
)

type AIExecution struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	LeftQuota   int       `db:"left_quota"`
	NextResetAt time.Time `db:"next_reset_at"`
}
