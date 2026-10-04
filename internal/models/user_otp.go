package models

import (
	"time"

	"github.com/google/uuid"
)

type UserOTP struct {
	UserID    uuid.UUID `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
}
