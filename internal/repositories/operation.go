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

const operationSelectColumns = `
		t.id,
		t.name,
		t.amount,
		t.type,
		t.planned_operation_id,
		t.goal_id,
		t.occurrence_date,
	a.id AS "account.id",
	a.name AS "account.name",
	a.currency AS "account.currency",
	a.icon AS "account.icon",
	fa.id AS "from_account.id",
	COALESCE(fa.name, '') AS "from_account.name",
	COALESCE(fa.currency::text, '') AS "from_account.currency",
	COALESCE(fa.icon, '') AS "from_account.icon",
	po.id AS "planned_operation.id",
	COALESCE(po.name, '') AS "planned_operation.name",
	g.id AS "goal.id",
	g.user_id AS "goal.user_id",
	COALESCE(g.name, '') AS "goal.name",
	g.deadline AS "goal.deadline",
	COALESCE(g.target_amount, 0) AS "goal.target_amount",
	COALESCE(g.current_amount, 0) AS "goal.current_amount",
	c.id AS "category.id",
	c.name AS "category.name",
	c.icon AS "category.icon",
	CASE WHEN c.user_id IS NULL THEN TRUE ELSE FALSE END AS "category.is_system"`

const operationJoins = `
	JOIN accounts a ON a.id = t.account_id
	JOIN categories c ON c.id = t.category_id
	LEFT JOIN accounts fa ON fa.id = t.from_account_id
	LEFT JOIN planned_operations po ON po.id = t.planned_operation_id
	LEFT JOIN goals g ON g.id = t.goal_id`

type OperationListParams struct {
	ListParams
	AccountID          uuid.UUID
	CategoryID         uuid.UUID
	PlannedOperationID uuid.UUID
	GoalID             uuid.UUID
	StartDate          time.Time
	EndDate            time.Time
}
type OperationRepository struct {
	db *sql.DB
}

func NewOperationRepository(db *sql.DB) *OperationRepository {
	return &OperationRepository{db: db}
}

func (r *OperationRepository) Create(ctx context.Context, op *models.Operation, tx *sql.Tx) (*dto.OperationResponseTransfer, error) {
	const q = `
		WITH inserted AS (
				INSERT INTO operations (name, type, account_id, from_account_id, category_id, amount, occurrence_date, planned_operation_id, goal_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING id, name, type, account_id, from_account_id, category_id, amount, occurrence_date, planned_operation_id, goal_id
		)
		SELECT ` + operationSelectColumns + `
		FROM inserted t
		` + operationJoins

	var resp dto.OperationResponseTransfer
	if err := sqlscan.Get(
		ctx, DBorTx(r.db, tx), &resp, q,
		op.Name, op.Type, op.AccountID, nullUUID(op.FromAccountID), op.CategoryID, op.Amount, op.OccurrenceDate, nullUUID(op.PlannedOperationID), nullUUID(op.GoalID),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *OperationRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.OperationResponseTransfer, error) {
	const q = `
		SELECT ` + operationSelectColumns + `
		FROM operations t
		` + operationJoins + `
		WHERE t.id = $1 AND a.user_id = $2`

	var resp dto.OperationResponseTransfer
	if err := sqlscan.Get(ctx, r.db, &resp, q, id, userID); err != nil {
		return nil, err
	}
	return &resp, nil
}

type operationListRow struct {
	dto.OperationResponseTransfer
	TotalCount int `db:"total_count"`
}

func (r *OperationRepository) List(ctx context.Context, params OperationListParams, userID uuid.UUID) ([]dto.OperationResponseTransfer, int, error) {
	const q = `
		SELECT ` + operationSelectColumns + `,
			COUNT(*) OVER() AS total_count
		FROM operations t
		` + operationJoins + `
			WHERE a.user_id = $1
				AND (t.account_id = $2 OR t.from_account_id = $2 OR $2 IS NULL)
				AND (t.category_id = $3 OR $3 IS NULL)
				AND (t.occurrence_date >= $4 OR $4 IS NULL)
				AND (t.occurrence_date <= $5 OR $5 IS NULL)
				AND (t.name ILIKE '%' || $6 || '%' OR $6 IS NULL)
				AND (t.planned_operation_id = $7 OR $7 IS NULL)
				AND (t.goal_id = $8 OR $8 IS NULL)
			ORDER BY t.occurrence_date DESC
			LIMIT $9 OFFSET $10`

	var rows []operationListRow
	if err := sqlscan.Select(
		ctx, r.db, &rows, q,
		userID,
		nullUUID(params.AccountID),
		nullUUID(params.CategoryID),
		nullTime(params.StartDate),
		nullTime(params.EndDate),
		nullString(params.Filter),
		nullUUID(params.PlannedOperationID),
		nullUUID(params.GoalID),
		nullInt(params.Limit),
		nullInt(params.Offset),
	); err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return nil, 0, nil
	}

	resp := make([]dto.OperationResponseTransfer, len(rows))
	for i, row := range rows {
		resp[i] = row.OperationResponseTransfer
	}
	return resp, rows[0].TotalCount, nil
}

func (r *OperationRepository) Update(ctx context.Context, op *models.Operation, tx *sql.Tx) (*dto.OperationResponseTransfer, error) {
	const q = `
		WITH updated AS (
			UPDATE operations
			SET name = COALESCE($2, name),
			    type = COALESCE($3::operation_type, type),
			    account_id = COALESCE($4, account_id),
			    from_account_id = COALESCE($5, from_account_id),
			    category_id = COALESCE($6, category_id),
			    amount = COALESCE($7, amount),
			    occurrence_date = COALESCE($8, occurrence_date)
			WHERE id = $1
			RETURNING id, name, type, account_id, from_account_id, category_id, amount, occurrence_date, planned_operation_id, goal_id
		)
		SELECT ` + operationSelectColumns + `
		FROM updated t
		` + operationJoins

	var resp dto.OperationResponseTransfer
	if err := sqlscan.Get(
		ctx, DBorTx(r.db, tx), &resp, q,
		op.ID,
		nullString(op.Name),
		nullString(string(op.Type)),
		nullUUID(op.AccountID),
		nullUUID(op.FromAccountID),
		nullUUID(op.CategoryID),
		nullDecimal(op.Amount),
		nullTime(op.OccurrenceDate),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *OperationRepository) BulkCreate(ctx context.Context, operations []models.Operation, tx *sql.Tx) error {
	if len(operations) == 0 {
		return nil
	}

	names := make([]string, len(operations))
	types := make([]string, len(operations))
	accountIDs := make([]string, len(operations))
	fromAccountIDs := make([]string, len(operations))
	categoryIDs := make([]string, len(operations))
	amounts := make([]string, len(operations))
	occurrenceDates := make([]string, len(operations))
	plannedOperationIDs := make([]string, len(operations))

	for i, operation := range operations {
		names[i] = operation.Name
		types[i] = string(operation.Type)
		accountIDs[i] = operation.AccountID.String()
		if operation.FromAccountID != uuid.Nil {
			fromAccountIDs[i] = operation.FromAccountID.String()
		}
		categoryIDs[i] = operation.CategoryID.String()
		amounts[i] = operation.Amount.String()
		occurrenceDates[i] = operation.OccurrenceDate.UTC().Format(time.RFC3339)
		if operation.PlannedOperationID != uuid.Nil {
			plannedOperationIDs[i] = operation.PlannedOperationID.String()
		}
	}

	const q = `
			INSERT INTO operations (name, type, account_id, from_account_id, category_id, amount, occurrence_date, planned_operation_id)
			SELECT
				v.name,
				v.type::operation_type,
				v.account_id::uuid,
				NULLIF(v.from_account_id, '')::uuid,
				v.category_id::uuid,
				v.amount::numeric,
				v.occurrence_date::timestamptz,
				NULLIF(v.planned_operation_id, '')::uuid
			FROM unnest($1::text[], $2::text[], $3::text[], $4::text[], $5::text[], $6::text[], $7::text[], $8::text[])
				AS v(name, type, account_id, from_account_id, category_id, amount, occurrence_date, planned_operation_id)`

	_, err := DBorTx(r.db, tx).ExecContext(
		ctx, q,
		names, types, accountIDs, fromAccountIDs, categoryIDs, amounts, occurrenceDates, plannedOperationIDs,
	)
	return err
}

func (r *OperationRepository) Delete(ctx context.Context, id uuid.UUID, tx *sql.Tx) error {
	const q = `DELETE FROM operations WHERE id = $1`

	_, err := DBorTx(r.db, tx).ExecContext(ctx, q, id)
	return err
}
