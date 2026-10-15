package repositories

import (
	"context"
	"database/sql"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
)

const subscriptionSelectColumns = `
	s.id,
	s.user_id,
	s.plan,
	s.start_date,
	s.end_date`

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	const q = `
		SELECT id, user_id, plan, start_date, end_date
		FROM subscription
		WHERE user_id = $1`

	var resp models.Subscription
	if err := sqlscan.Get(ctx, r.db, &resp, q, userID); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription, tx *sql.Tx) (*dto.SubscriptionResponse, error) {
	const q = `
		INSERT INTO subscription AS s (user_id, plan, start_date, end_date)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + subscriptionSelectColumns

	var resp dto.SubscriptionResponse
	if err := sqlscan.Get(
		ctx, DBorTx(r.db, tx), &resp, q,
		subscription.UserID, subscription.Plan, subscription.StartDate, subscription.EndDate,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, userID uuid.UUID, subscription *models.Subscription) (*dto.SubscriptionResponse, error) {
	const q = `
		UPDATE subscription AS s
		SET plan = COALESCE($3::subscription_plan, s.plan),
		    start_date = COALESCE($4::date, s.start_date),
		    end_date = COALESCE($5::date, s.end_date)
		WHERE s.id = $1 AND s.user_id = $2
		RETURNING ` + subscriptionSelectColumns

	var resp dto.SubscriptionResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		subscription.ID,
		userID,
		nullString(string(subscription.Plan)),
		nullTime(subscription.StartDate),
		nullTime(subscription.EndDate),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}
