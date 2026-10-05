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
	t.occurrence_date,
	a.id AS "account.id",
	a.name AS "account.name",
	a.currency AS "account.currency",
	a.icon AS "account.icon",
	fa.id AS "from_account.id",
	COALESCE(fa.name, '') AS "from_account.name",
	COALESCE(fa.currency::text, '') AS "from_account.currency",
	COALESCE(fa.icon, '') AS "from_account.icon",
	c.id AS "category.id",
	c.name AS "category.name",
	c.icon AS "category.icon"`

const operationJoins = `
	JOIN accounts a ON a.id = t.account_id
	JOIN categories c ON c.id = t.category_id
	LEFT JOIN accounts fa ON fa.id = t.from_account_id`

type OperationListParams struct {
	AccountID  uuid.UUID
	CategoryID uuid.UUID
	StartDate  time.Time
	EndDate    time.Time
	Limit      int
	Offset     int
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
			INSERT INTO operations (name, type, account_id, from_account_id, category_id, amount, occurrence_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, name, type, account_id, from_account_id, category_id, amount, occurrence_date
		)
		SELECT ` + operationSelectColumns + `
		FROM inserted t
		` + operationJoins

	var resp dto.OperationResponseTransfer
	if err := sqlscan.Get(
		ctx, DBorTx(r.db, tx), &resp, q,
		op.Name, op.Type, op.AccountID, nullUUID(op.FromAccountID), op.CategoryID, op.Amount, op.OccurrenceDate,
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

func (r *OperationRepository) List(ctx context.Context, params OperationListParams, userID uuid.UUID) ([]dto.OperationResponseTransfer, error) {
	const q = `
		SELECT ` + operationSelectColumns + `
		FROM operations t
		` + operationJoins + `
		WHERE a.user_id = $1
			AND (t.account_id = $2 OR t.from_account_id = $2 OR $2 IS NULL)
			AND (t.category_id = $3 OR $3 IS NULL)
			AND (t.occurrence_date >= $4 OR $4 IS NULL)
			AND (t.occurrence_date <= $5 OR $5 IS NULL)
		ORDER BY t.occurrence_date DESC
		LIMIT $6 OFFSET $7`

	var resp []dto.OperationResponseTransfer
	if err := sqlscan.Select(
		ctx, r.db, &resp, q,
		userID,
		nullUUID(params.AccountID),
		nullUUID(params.CategoryID),
		nullTime(params.StartDate),
		nullTime(params.EndDate),
		nullInt(params.Limit),
		nullInt(params.Offset),
	); err != nil {
		return nil, err
	}
	return resp, nil
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
			RETURNING id, name, type, account_id, from_account_id, category_id, amount, occurrence_date
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
	}

	const q = `
		INSERT INTO operations (name, type, account_id, from_account_id, category_id, amount, occurrence_date)
		SELECT
			v.name,
			v.type::operation_type,
			v.account_id::uuid,
			NULLIF(v.from_account_id, '')::uuid,
			v.category_id::uuid,
			v.amount::numeric,
			v.occurrence_date::timestamptz
		FROM unnest($1::text[], $2::text[], $3::text[], $4::text[], $5::text[], $6::text[], $7::text[])
			AS v(name, type, account_id, from_account_id, category_id, amount, occurrence_date)`

	_, err := DBorTx(r.db, tx).ExecContext(
		ctx, q,
		names, types, accountIDs, fromAccountIDs, categoryIDs, amounts, occurrenceDates,
	)
	return err
}

func (r *OperationRepository) Delete(ctx context.Context, id uuid.UUID, tx *sql.Tx) error {
	const q = `DELETE FROM operations WHERE id = $1`

	_, err := DBorTx(r.db, tx).ExecContext(ctx, q, id)
	return err
}
