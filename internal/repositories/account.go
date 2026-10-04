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

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *models.Account) (*dto.AccountResponse, error) {
	const q = `
		INSERT INTO accounts (name, currency, icon, balance, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, currency, icon, balance, user_id`

	var resp dto.AccountResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		account.Name, account.Currency, account.Icon, account.Balance, account.UserID,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.AccountResponse, error) {
	q := `SELECT id, name, currency, icon, balance, user_id FROM accounts WHERE id = $1`

	var resp dto.AccountResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *AccountRepository) GetByIDs(ctx context.Context, operations []dto.PlannedOperationDueResponse) ([]dto.AccountResponse, error) {
	if len(operations) == 0 {
		return []dto.AccountResponse{}, nil
	}

	ids := make([]uuid.UUID, len(operations))
	for i, operation := range operations {
		ids[i] = operation.Account.ID
	}

	const q = `
		SELECT a.id, a.name, a.currency, a.icon, a.balance, a.user_id
		FROM accounts AS a
		JOIN unnest($1::text[]) AS ids(id) ON a.id = ids.id::uuid
		GROUP BY a.id`

	var resp []dto.AccountResponse
	if err := sqlscan.Select(ctx, r.db, &resp, q, ids); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *AccountRepository) List(ctx context.Context, userID uuid.UUID) ([]dto.AccountResponse, error) {
	const q = `
		SELECT id, name, currency, icon, balance, user_id
		FROM accounts
		WHERE user_id = $1
		ORDER BY name`

	var resp []dto.AccountResponse
	if err := sqlscan.Select(ctx, r.db, &resp, q, userID); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *AccountRepository) Update(ctx context.Context, account *models.Account) (*dto.AccountResponse, error) {
	const q = `
		UPDATE accounts
		SET name = COALESCE($2, name),
		    icon = COALESCE($3, icon)
		WHERE id = $1
		RETURNING id, name, currency, icon, user_id`

	var resp dto.AccountResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		account.ID, nullString(account.Name), nullString(account.Icon),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *AccountRepository) BulkUpdateBalance(ctx context.Context, updates []dto.AccountBalanceUpdate, tx *sql.Tx) error {
	if len(updates) == 0 {
		return nil
	}

	ids := make([]string, len(updates))
	deltas := make([]string, len(updates))
	for i, update := range updates {
		ids[i] = update.ID.String()
		deltas[i] = update.Delta.String()
	}

	const q = `
		UPDATE accounts AS a
		SET balance = a.balance + v.delta
		FROM (
			SELECT id::uuid, SUM(delta::numeric) AS delta
			FROM unnest($1::text[], $2::text[]) AS u(id, delta)
			GROUP BY id
		) AS v
		WHERE a.id = v.id`

	_, err := DBorTx(r.db, tx).ExecContext(ctx, q, ids, deltas)
	return err
}

func (r *AccountRepository) UpdateBalance(ctx context.Context, accountID uuid.UUID, delta decimal.Decimal, tx *sql.Tx) error {
	const q = `
		UPDATE accounts
		SET balance = balance + $2
		WHERE id = $1
		RETURNING balance`

	var balance decimal.Decimal
	if err := sqlscan.Get(
		ctx, DBorTx(r.db, tx), &balance, q, accountID, delta,
	); err != nil {
		return err
	}
	return nil
}

func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM accounts WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
