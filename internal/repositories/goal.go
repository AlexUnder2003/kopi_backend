package repositories

import (
	"context"
	"database/sql"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const goalSelectColumns = `
	g.id,
	g.user_id,
	g.name,
	g.currency,
	g.deadline,
	g.target_amount,
	g.current_amount`

type GoalRepository struct {
	db *sql.DB
}

func NewGoalRepository(db *sql.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

func (r *GoalRepository) Create(ctx context.Context, goal *models.Goal) (*dto.GoalResponse, error) {
	const q = `
		INSERT INTO goals AS g (user_id, name, currency, deadline, target_amount, current_amount)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6::numeric, 0))
		RETURNING ` + goalSelectColumns

	var resp dto.GoalResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		goal.UserID,
		goal.Name,
		goal.Currency,
		nullTime(goal.Deadline),
		goal.TargetAmount,
		nullDecimalPtr(goal.CurrentAmount),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *GoalRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.GoalResponse, error) {
	const q = `
		SELECT ` + goalSelectColumns + `
		FROM goals AS g
		WHERE g.id = $1 AND g.user_id = $2`

	var resp dto.GoalResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id, userID); err != nil {
		return nil, err
	}
	return &resp, nil
}

type goalListRow struct {
	dto.GoalResponse
	TotalCount int `db:"total_count"`
}

func (r *GoalRepository) List(ctx context.Context, userID uuid.UUID, params ListParams) ([]dto.GoalResponse, int, error) {
	const q = `
		SELECT ` + goalSelectColumns + `,
			COUNT(*) OVER() AS total_count
		FROM goals AS g
		WHERE g.user_id = $1
			AND (g.name ILIKE '%' || $2 || '%' OR $2 IS NULL)
		ORDER BY g.deadline ASC NULLS LAST, g.created_at DESC
		LIMIT $3 OFFSET $4`

	var rows []goalListRow
	if err := sqlscan.Select(
		ctx, r.db, &rows, q,
		userID, nullString(params.Filter), nullInt(params.Limit), nullInt(params.Offset),
	); err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return nil, 0, nil
	}

	resp := make([]dto.GoalResponse, len(rows))
	for i, row := range rows {
		resp[i] = row.GoalResponse
	}
	return resp, rows[0].TotalCount, nil
}

func (r *GoalRepository) Update(ctx context.Context, goal *models.Goal) (*dto.GoalResponse, error) {
	const q = `
		UPDATE goals AS g
		SET name = COALESCE($3, g.name),
		    deadline = COALESCE($4::timestamp, g.deadline),
		    target_amount = COALESCE($5::numeric, g.target_amount),
		    current_amount = COALESCE($6::numeric, g.current_amount),
		    updated_at = CURRENT_TIMESTAMP
		WHERE g.id = $1 AND g.user_id = $2
		RETURNING ` + goalSelectColumns

	var resp dto.GoalResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		goal.ID,
		goal.UserID,
		nullString(goal.Name),
		nullTime(goal.Deadline),
		nullDecimal(goal.TargetAmount),
		nullDecimalPtr(goal.CurrentAmount),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *GoalRepository) UpdateBalance(ctx context.Context, id uuid.UUID, delta decimal.Decimal, tx *sql.Tx) error {
	const q = `
		UPDATE goals
		SET current_amount = current_amount + $2,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := DBorTx(r.db, tx).ExecContext(ctx, q, id, delta)
	return err
}

func (r *GoalRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	const q = `DELETE FROM goals WHERE id = $1 AND user_id = $2`

	_, err := r.db.ExecContext(ctx, q, id, userID)
	return err
}
