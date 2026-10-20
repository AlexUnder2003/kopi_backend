package repositories

import (
	"context"
	"database/sql"
	"time"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const plannedOperationSelectColumns = `
	p.id,
	p.name,
	p.amount,
	p.balance,
	p.type,
	p.interval_type,
	p."interval",
	p.planned_at,
	p.next_run_at,
	p.is_recurring,
	a.id AS "account.id",
	a.name AS "account.name",
	a.currency AS "account.currency",
	a.icon AS "account.icon",
	c.id AS "category.id",
	c.name AS "category.name",
	c.icon AS "category.icon",
	CASE WHEN c.user_id IS NULL THEN TRUE ELSE FALSE END AS "category.is_system"`

type PlannedOperationRepository struct {
	db *sql.DB
}

type PlannedOperationListParams struct {
	ListParams
	AccountID uuid.UUID
	UserID    uuid.UUID
	StartDate time.Time
	EndDate   time.Time
}

func NewPlannedOperationRepository(db *sql.DB) *PlannedOperationRepository {
	return &PlannedOperationRepository{db: db}
}

func (r *PlannedOperationRepository) Create(ctx context.Context, op *models.PlannedOperation) (*dto.PlannedOperationResponse, error) {
	const q = `
		WITH inserted AS (
			INSERT INTO planned_operations (
				name, account_id, amount, balance, type, interval_type, "interval", category_id, planned_at, next_run_at, is_recurring
			)
			VALUES ($1, $2, $3, $4, $5, $6::interval_type, $7, $8, $9, $10, $11)
			RETURNING id, name, account_id, amount, balance, type, interval_type, "interval", category_id, planned_at, next_run_at, is_recurring
		)
		SELECT ` + plannedOperationSelectColumns + `
		FROM inserted p
		JOIN accounts a ON a.id = p.account_id
		JOIN categories c ON c.id = p.category_id`

	var resp dto.PlannedOperationResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		op.Name, op.AccountID, op.Amount, op.Balance, op.Type, nullString(string(op.IntervalType)), nullInt(op.Interval),
		op.CategoryID, op.PlannedAt, nullTime(op.NextRunAt), op.IsRecurring,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *PlannedOperationRepository) NearestExpense(ctx context.Context, userID uuid.UUID, params ListParams) (*dto.PlannedOperationResponse, error) {
	const q = `
		SELECT ` + plannedOperationSelectColumns + `
		FROM planned_operations p
		JOIN accounts a ON a.id = p.account_id
		JOIN categories c ON c.id = p.category_id
		WHERE a.user_id = $1 AND p.type = 'expense'
		ORDER BY p.next_run_at ASC
		LIMIT 1`

	var resp dto.PlannedOperationResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, userID); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *PlannedOperationRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.PlannedOperationResponse, error) {
	const q = `
		SELECT ` + plannedOperationSelectColumns + `
		FROM planned_operations p
		JOIN accounts a ON a.id = p.account_id
		JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1 AND a.user_id = $2`

	var resp dto.PlannedOperationResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id, userID); err != nil {
		return nil, err
	}
	return &resp, nil
}

type plannedOperationListRow struct {
	dto.PlannedOperationResponse
	TotalCount int `db:"total_count"`
}

func (r *PlannedOperationRepository) List(ctx context.Context, params PlannedOperationListParams) ([]dto.PlannedOperationResponse, int, error) {
	const q = `
		SELECT ` + plannedOperationSelectColumns + `,
			COUNT(*) OVER() AS total_count
		FROM planned_operations p
		JOIN accounts a ON a.id = p.account_id
		JOIN categories c ON c.id = p.category_id
			WHERE (a.user_id = $1 OR $1 IS NULL) AND (p.account_id = $2 OR $2 IS NULL)
				AND (p.next_run_at >= $3 OR $3 IS NULL) AND (p.next_run_at <= $4 OR $4 IS NULL)
				AND (p.name ILIKE '%' || $5 || '%' OR $5 IS NULL)
			ORDER BY p.planned_at
			LIMIT $6 OFFSET $7`

	var rows []plannedOperationListRow
	if err := sqlscan.Select(ctx, r.db, &rows, q,
		nullUUID(params.UserID),
		nullUUID(params.AccountID),
		nullTime(params.StartDate),
		nullTime(params.EndDate),
		nullString(params.Filter),
		nullInt(params.Limit),
		nullInt(params.Offset),
	); err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return nil, 0, nil
	}

	resp := make([]dto.PlannedOperationResponse, len(rows))
	for i, row := range rows {
		resp[i] = row.PlannedOperationResponse
	}
	return resp, rows[0].TotalCount, nil
}

func (r *PlannedOperationRepository) Update(ctx context.Context, op *models.PlannedOperation) (*dto.PlannedOperationResponse, error) {
	const q = `
		WITH updated AS (
			UPDATE planned_operations
			SET name = COALESCE($2, name),
			    account_id = COALESCE($3, account_id),
			    amount = COALESCE($4::numeric, amount),
			    balance = COALESCE($5::numeric, balance),
			    type = COALESCE($6::operation_type, type),
			    interval_type = COALESCE($7::interval_type, interval_type),
			    "interval" = COALESCE($8::int, "interval"),
			    category_id = COALESCE($9, category_id),
			    planned_at = COALESCE($10::timestamptz, planned_at),
			    next_run_at = COALESCE($11::timestamptz, next_run_at),
			    is_recurring = COALESCE($12::boolean, is_recurring)
			WHERE id = $1
			RETURNING id, name, account_id, amount, balance, type, interval_type, "interval", category_id, planned_at, next_run_at, is_recurring
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
		nullDecimal(op.Balance),
		nullString(string(op.Type)),
		nullString(string(op.IntervalType)),
		nullInt(op.Interval),
		nullUUID(op.CategoryID),
		nullTime(op.PlannedAt),
		nullTime(op.NextRunAt),
		op.IsRecurring,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *PlannedOperationRepository) GetBalance(ctx context.Context, nextRunAt time.Time, userID uuid.UUID) (decimal.Decimal, error) {
	const q = `
		SELECT COALESCE(SUM(p.balance), 0)
		FROM planned_operations p
		JOIN accounts a ON a.id = p.account_id
		WHERE p.balance > 0
			AND p.next_run_at > $1
			AND a.user_id = $2
	`
	var balance decimal.Decimal
	if err := r.db.QueryRowContext(ctx, q, nextRunAt, userID).Scan(&balance); err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

func (r *PlannedOperationRepository) GetByNextRunAt(ctx context.Context, nextRunAt time.Time, tx *sql.Tx) ([]dto.PlannedOperationDueResponse, error) {
	const q = `
		SELECT ` + plannedOperationSelectColumns + `,
			a.user_id
		FROM planned_operations p
		JOIN accounts a ON a.id = p.account_id
		JOIN categories c ON c.id = p.category_id
		WHERE p.next_run_at <= $1
		ORDER BY p.next_run_at
		FOR UPDATE SKIP LOCKED`
	var resp []dto.PlannedOperationDueResponse
	if err := sqlscan.Select(ctx, DBorTx(r.db, tx), &resp, q, nextRunAt); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *PlannedOperationRepository) BulkUpdate(ctx context.Context, ops []models.PlannedOperation, tx *sql.Tx) error {
	if len(ops) == 0 {
		return nil
	}

	ids := make([]string, len(ops))
	nextRunAts := make([]string, len(ops))

	for i, op := range ops {
		ids[i] = op.ID.String()
		if !op.NextRunAt.IsZero() {
			nextRunAts[i] = op.NextRunAt.UTC().Format(time.RFC3339)
		}
	}

	const q = `
		UPDATE planned_operations AS p
		SET next_run_at = NULLIF(v.next_run_at, '')::timestamptz
		FROM unnest($1::text[], $2::text[]) AS v(id, next_run_at)
		WHERE p.id = v.id::uuid`

	_, err := DBorTx(r.db, tx).ExecContext(ctx, q, ids, nextRunAts)
	return err
}

func (r *PlannedOperationRepository) GetNearestIncomeDate(
	ctx context.Context,
	userID uuid.UUID,
	from time.Time,
) (*time.Time, error) {
	const q = `
		SELECT p.next_run_at
		FROM planned_operations p
		JOIN accounts a ON a.id = p.account_id
		WHERE a.user_id = $1
			AND p.type = 'income'
			AND p.next_run_at >= $2
		ORDER BY p.next_run_at ASC
		LIMIT 1`

	var nextRunAt time.Time
	if err := r.db.QueryRowContext(ctx, q, userID, from).Scan(&nextRunAt); err != nil {
		return nil, err
	}
	return &nextRunAt, nil
}

func (r *PlannedOperationRepository) UpdateBalance(ctx context.Context, id uuid.UUID, delta decimal.Decimal, tx *sql.Tx) error {
	const q = `
		UPDATE planned_operations
		SET balance = balance + $2
		WHERE id = $1`

	_, err := DBorTx(r.db, tx).ExecContext(ctx, q, id, delta)
	if err != nil {
		return err
	}
	return nil
}

func (r *PlannedOperationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM planned_operations WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
