package repositories

import (
	"context"
	"database/sql"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
)

const plannedOperationSelectColumns = `
	p.id,
	p.name,
	p.amount,
	p.type,
	p.frequency,
	p.planned_at,
	p.is_recurring,
	a.id AS "account.id",
	a.name AS "account.name",
	a.currency AS "account.currency",
	a.icon AS "account.icon",
	c.id AS "category.id",
	c.name AS "category.name",
	c.icon AS "category.icon"`

type PlannedOperationRepository struct {
	db *sql.DB
}

type PlannedOperationListParams struct {
	AccountID uuid.UUID
	UserID    uuid.UUID
}

func NewPlannedOperationRepository(db *sql.DB) *PlannedOperationRepository {
	return &PlannedOperationRepository{db: db}
}

func (r *PlannedOperationRepository) Create(ctx context.Context, op *models.PlannedOperation) (*dto.PlannedOperationResponse, error) {
	const q = `
		WITH inserted AS (
			INSERT INTO planned_operations (
				name, account_id, amount, type, frequency, category_id, planned_at, next_run_at, is_recurring
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, name, account_id, amount, type, frequency, category_id, planned_at, is_recurring
		)
		SELECT ` + plannedOperationSelectColumns + `
		FROM inserted p
		JOIN accounts a ON a.id = p.account_id
		JOIN categories c ON c.id = p.category_id`

	var resp dto.PlannedOperationResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		op.Name, op.AccountID, op.Amount, op.Type, op.Frequency,
		op.CategoryID, op.PlannedAt, op.NextRunAt, op.IsRecurring,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *PlannedOperationRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.PlannedOperationResponse, error) {
	const q = `
		SELECT ` + plannedOperationSelectColumns + `
		FROM planned_operations p
		JOIN accounts a ON a.id = p.account_id
		JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1`

	var resp dto.PlannedOperationResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *PlannedOperationRepository) List(ctx context.Context, params PlannedOperationListParams) ([]dto.PlannedOperationResponse, error) {
	const q = `
		SELECT ` + plannedOperationSelectColumns + `
		FROM planned_operations p
		JOIN accounts a ON a.id = p.account_id
		JOIN categories c ON c.id = p.category_id
		WHERE (a.user_id = $1 OR $1 IS NULL) AND (p.account_id = $2 OR $2 IS NULL)
		ORDER BY p.planned_at`

	var resp []dto.PlannedOperationResponse
	if err := sqlscan.Select(ctx, r.db, &resp, q,
		nullUUID(params.UserID),
		nullUUID(params.AccountID),
	); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *PlannedOperationRepository) Update(ctx context.Context, op *models.PlannedOperation) (*dto.PlannedOperationResponse, error) {
	const q = `
		WITH updated AS (
			UPDATE planned_operations
			SET name = COALESCE($2, name),
			    account_id = COALESCE($3, account_id),
			    amount = COALESCE($4, amount),
			    type = COALESCE($5::transaction_type, type),
			    frequency = COALESCE($6::planned_operation_frequency, frequency),
			    category_id = COALESCE($7, category_id),
			    planned_at = COALESCE($8, planned_at),
			    next_run_at = COALESCE($9, next_run_at),
			    is_recurring = COALESCE($10::boolean, is_recurring)
			WHERE id = $1
			RETURNING id, name, account_id, amount, type, frequency, category_id, planned_at, is_recurring
		)
		SELECT ` + plannedOperationSelectColumns + `
		FROM updated p
		JOIN accounts a ON a.id = p.account_id
		JOIN categories c ON c.id = p.category_id`

	var resp dto.PlannedOperationResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		op.ID,
		nullString(op.Name),
		nullUUID(op.AccountID),
		nullDecimal(op.Amount),
		nullString(string(op.Type)),
		nullString(string(op.Frequency)),
		nullUUID(op.CategoryID),
		nullTime(op.PlannedAt),
		nullTimePtr(op.NextRunAt),
		nullBoolPtr(op.IsRecurring),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *PlannedOperationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM planned_operations WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
