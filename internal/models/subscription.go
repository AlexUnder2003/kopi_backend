package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionPlan string

const (
	SubscriptionPlanFree    SubscriptionPlan = "free"
	SubscriptionPlanPro     SubscriptionPlan = "pro"
	SubscriptionPlanProPlus SubscriptionPlan = "pro_plus"
)

type Subscription struct {
	ID        uuid.UUID        `db:"id"`
	UserID    uuid.UUID        `db:"user_id"`
	Plan      SubscriptionPlan `db:"plan"`
	StartDate time.Time        `db:"start_date"`
	EndDate   time.Time        `db:"end_date"`
}
