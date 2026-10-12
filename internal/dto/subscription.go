package dto

import (
	"time"

	"KopiBackend/internal/models"

	"github.com/google/uuid"
)

type SubscriptionResponse struct {
	ID        uuid.UUID               `json:"id" db:"id"`
	UserID    uuid.UUID               `json:"-" db:"user_id"`
	Plan      models.SubscriptionPlan `json:"plan" db:"plan"`
	StartDate time.Time               `json:"start_date" db:"start_date"`
	EndDate   time.Time               `json:"end_date,omitempty" db:"end_date"`
}
