package repositories

import (
	"context"
	"database/sql"
	"time"

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

func (r *SubscriptionRepository) GetTodayExpiredSubscriptions(ctx context.Context, tx *sql.Tx) ([]*models.Subscription, error) {
	const q = `
		SELECT id, user_id, plan, start_date, end_date
		FROM subscription
		WHERE end_date <= CURRENT_TIMESTAMP
		FOR UPDATE SKIP LOCKED`

	var resp []*models.Subscription
	if err := sqlscan.Select(ctx, tx, &resp, q); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *SubscriptionRepository) BulkUpdate(ctx context.Context, subscriptions []models.Subscription, tx *sql.Tx) error {
	if len(subscriptions) == 0 {
		return nil
	}

	ids := make([]string, len(subscriptions))
	userIDs := make([]string, len(subscriptions))
	plans := make([]string, len(subscriptions))
	startDates := make([]string, len(subscriptions))
	endDates := make([]string, len(subscriptions))

	for i, subscription := range subscriptions {
		ids[i] = subscription.ID.String()
		userIDs[i] = subscription.UserID.String()
		plans[i] = string(subscription.Plan)
		if !subscription.StartDate.IsZero() {
			startDates[i] = subscription.StartDate.Format(time.DateOnly)
		}
		if !subscription.EndDate.IsZero() {
			endDates[i] = subscription.EndDate.Format(time.DateOnly)
		}
	}

	const q = `
		UPDATE subscription AS s
		SET plan = COALESCE(NULLIF(v.plan, '')::subscription_plan, s.plan),
		    start_date = COALESCE(NULLIF(v.start_date, '')::date, s.start_date),
		    end_date = COALESCE(NULLIF(v.end_date, '')::date, s.end_date),
		    updated_at = CURRENT_TIMESTAMP
		FROM unnest($1::text[], $2::text[], $3::text[], $4::text[], $5::text[]) AS v(id, user_id, plan, start_date, end_date)
		WHERE s.id = v.id::uuid AND s.user_id = v.user_id::uuid`

	_, err := tx.ExecContext(ctx, q, ids, userIDs, plans, startDates, endDates)
	return err
}
