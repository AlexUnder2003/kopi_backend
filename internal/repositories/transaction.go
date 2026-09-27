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

const transactionSelectColumns = `
	t.id,
	t.name,
	t.amount,
	t.type,
	t.occurrence_date,
	a.id AS "account.id",
	a.name AS "account.name",
	a.currency AS "account.currency",
	a.icon AS "account.icon",
	c.id AS "category.id",
	c.name AS "category.name",
	c.icon AS "category.icon"`

type TransactionListParams struct {
	AccountID  uuid.UUID
	CategoryID uuid.UUID
	StartDate  time.Time
	EndDate    time.Time
	Limit      int
	Offset     int
}

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, txModel *models.Transaction, tx *sql.Tx) (*dto.TransactionResponse, error) {
	const q = `
		WITH inserted AS (
			INSERT INTO transactions (name, type, account_id, category_id, amount, occurrence_date, transfer_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, name, type, account_id, category_id, amount, occurrence_date
		)
		SELECT ` + transactionSelectColumns + `
		FROM inserted t
		JOIN accounts a ON a.id = t.account_id
		JOIN categories c ON c.id = t.category_id`

	var resp dto.TransactionResponse
	if err := sqlscan.Get(
		ctx, DBorTx(r.db, tx), &resp, q,
		txModel.Name, txModel.Type, txModel.AccountID, txModel.CategoryID, txModel.Amount, txModel.OccurrenceDate, txModel.TransferID,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.TransactionResponse, error) {
	const q = `
		SELECT ` + transactionSelectColumns + `
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		JOIN categories c ON c.id = t.category_id
		WHERE t.id = $1 AND a.user_id = $2`

	var resp dto.TransactionResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id, userID); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *TransactionRepository) List(ctx context.Context, params TransactionListParams, userID uuid.UUID) ([]dto.TransactionResponse, error) {
	const q = `
		SELECT ` + transactionSelectColumns + `
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		JOIN categories c ON c.id = t.category_id
		WHERE a.user_id = $1
			AND (t.account_id = $2 OR $2 IS NULL)
			AND (t.category_id = $3 OR $3 IS NULL)
			AND (t.occurrence_date >= $4 OR $4 IS NULL)
			AND (t.occurrence_date <= $5 OR $5 IS NULL)
		ORDER BY t.occurrence_date DESC
		LIMIT $6 OFFSET $7`

	var resp []dto.TransactionResponse
	if err := sqlscan.Select(
		ctx, r.db, &resp, q,
		userID,
		nullUUID(params.AccountID),
		nullUUID(params.CategoryID),
		nullTime(params.StartDate),
		nullTime(params.EndDate),
		params.Limit,
		params.Offset,
	); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *TransactionRepository) GetByTransferID(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) ([]dto.TransactionResponse, error) {
	const q = `
		SELECT ` + transactionSelectColumns + `
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		JOIN categories c ON c.id = t.category_id
		WHERE t.transfer_id = $1
			AND a.user_id = $2
		ORDER BY t.occurrence_date DESC`

	var resp []dto.TransactionResponse
	if err := sqlscan.Select(ctx, r.db, &resp, q, transferID, userID); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *TransactionRepository) Update(ctx context.Context, txModel *models.Transaction, tx *sql.Tx) (*dto.TransactionResponse, error) {
	const q = `
		WITH updated AS (
			UPDATE transactions
			SET name = COALESCE($2, name),
			    type = COALESCE($3::transaction_type, type),
			    account_id = COALESCE($4, account_id),
			    category_id = COALESCE($5, category_id),
			    amount = COALESCE($6, amount),
			    occurrence_date = COALESCE($7, occurrence_date)
			WHERE id = $1
			RETURNING id, name, type, account_id, category_id, amount, occurrence_date
		)
		SELECT ` + transactionSelectColumns + `
		FROM updated t
		JOIN accounts a ON a.id = t.account_id
		JOIN categories c ON c.id = t.category_id`

	var resp dto.TransactionResponse
	if err := sqlscan.Get(
		ctx, DBorTx(r.db, tx), &resp, q,
		txModel.ID,
		nullString(txModel.Name),
		nullString(string(txModel.Type)),
		nullUUID(txModel.AccountID),
		nullUUID(txModel.CategoryID),
		nullDecimal(txModel.Amount),
		nullTime(txModel.OccurrenceDate),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *TransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM transactions WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
