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
	fa.id AS "from_account.id",
	COALESCE(fa.name, '') AS "from_account.name",
	COALESCE(fa.currency::text, '') AS "from_account.currency",
	COALESCE(fa.icon, '') AS "from_account.icon",
	c.id AS "category.id",
	c.name AS "category.name",
	c.icon AS "category.icon"`

const transactionJoins = `
	JOIN accounts a ON a.id = t.account_id
	JOIN categories c ON c.id = t.category_id
	LEFT JOIN accounts fa ON fa.id = t.from_account_id`

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

func (r *TransactionRepository) Create(ctx context.Context, txModel *models.Transaction, tx *sql.Tx) (*dto.TransactionResponseTransfer, error) {
	const q = `
		WITH inserted AS (
			INSERT INTO transactions (name, type, account_id, from_account_id, category_id, amount, occurrence_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, name, type, account_id, from_account_id, category_id, amount, occurrence_date
		)
		SELECT ` + transactionSelectColumns + `
		FROM inserted t
		` + transactionJoins

	var resp dto.TransactionResponseTransfer
	if err := sqlscan.Get(
		ctx, DBorTx(r.db, tx), &resp, q,
		txModel.Name, txModel.Type, txModel.AccountID, nullUUID(txModel.FromAccountID), txModel.CategoryID, txModel.Amount, txModel.OccurrenceDate,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.TransactionResponseTransfer, error) {
	const q = `
		SELECT ` + transactionSelectColumns + `
		FROM transactions t
		` + transactionJoins + `
		WHERE t.id = $1 AND a.user_id = $2`

	var resp dto.TransactionResponseTransfer
	if err := sqlscan.Get(ctx, r.db, &resp, q, id, userID); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *TransactionRepository) List(ctx context.Context, params TransactionListParams, userID uuid.UUID) ([]dto.TransactionResponseTransfer, error) {
	const q = `
		SELECT ` + transactionSelectColumns + `
		FROM transactions t
		` + transactionJoins + `
		WHERE a.user_id = $1
			AND (t.account_id = $2 OR t.from_account_id = $2 OR $2 IS NULL)
			AND (t.category_id = $3 OR $3 IS NULL)
			AND (t.occurrence_date >= $4 OR $4 IS NULL)
			AND (t.occurrence_date <= $5 OR $5 IS NULL)
		ORDER BY t.occurrence_date DESC
		LIMIT $6 OFFSET $7`

	var resp []dto.TransactionResponseTransfer
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

func (r *TransactionRepository) Update(ctx context.Context, txModel *models.Transaction, tx *sql.Tx) (*dto.TransactionResponseTransfer, error) {
	const q = `
		WITH updated AS (
			UPDATE transactions
			SET name = COALESCE($2, name),
			    type = COALESCE($3::transaction_type, type),
			    account_id = COALESCE($4, account_id),
			    from_account_id = COALESCE($5, from_account_id),
			    category_id = COALESCE($6, category_id),
			    amount = COALESCE($7, amount),
			    occurrence_date = COALESCE($8, occurrence_date)
			WHERE id = $1
			RETURNING id, name, type, account_id, from_account_id, category_id, amount, occurrence_date
		)
		SELECT ` + transactionSelectColumns + `
		FROM updated t
		` + transactionJoins

	var resp dto.TransactionResponseTransfer
	if err := sqlscan.Get(
		ctx, DBorTx(r.db, tx), &resp, q,
		txModel.ID,
		nullString(txModel.Name),
		nullString(string(txModel.Type)),
		nullUUID(txModel.AccountID),
		nullUUID(txModel.FromAccountID),
		nullUUID(txModel.CategoryID),
		nullDecimal(txModel.Amount),
		nullTime(txModel.OccurrenceDate),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *TransactionRepository) BulkCreate(ctx context.Context, transactions []models.Transaction, tx *sql.Tx) error {
	if len(transactions) == 0 {
		return nil
	}

	names := make([]string, len(transactions))
	types := make([]string, len(transactions))
	accountIDs := make([]string, len(transactions))
	fromAccountIDs := make([]string, len(transactions))
	categoryIDs := make([]string, len(transactions))
	amounts := make([]string, len(transactions))
	occurrenceDates := make([]string, len(transactions))

	for i, transaction := range transactions {
		names[i] = transaction.Name
		types[i] = string(transaction.Type)
		accountIDs[i] = transaction.AccountID.String()
		if transaction.FromAccountID != uuid.Nil {
			fromAccountIDs[i] = transaction.FromAccountID.String()
		}
		categoryIDs[i] = transaction.CategoryID.String()
		amounts[i] = transaction.Amount.String()
		occurrenceDates[i] = transaction.OccurrenceDate.Format(time.DateOnly)
	}

	const q = `
		INSERT INTO transactions (name, type, account_id, from_account_id, category_id, amount, occurrence_date)
		SELECT
			v.name,
			v.type::transaction_type,
			v.account_id::uuid,
			NULLIF(v.from_account_id, '')::uuid,
			v.category_id::uuid,
			v.amount::numeric,
			v.occurrence_date::date
		FROM unnest($1::text[], $2::text[], $3::text[], $4::text[], $5::text[], $6::text[], $7::text[])
			AS v(name, type, account_id, from_account_id, category_id, amount, occurrence_date)`

	_, err := DBorTx(r.db, tx).ExecContext(
		ctx, q,
		names, types, accountIDs, fromAccountIDs, categoryIDs, amounts, occurrenceDates,
	)
	return err
}

func (r *TransactionRepository) Delete(ctx context.Context, id uuid.UUID, tx *sql.Tx) error {
	const q = `DELETE FROM transactions WHERE id = $1`

	_, err := DBorTx(r.db, tx).ExecContext(ctx, q, id)
	return err
}
