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

const budgetSelectColumns = `
	b.id,
	b.name,
	b.amount,
	b.balance,
	b.interval_type,
	b."interval",
	b.currency,
	b.start_date,
	b.reset_date,
	c.id AS "category.id",
	c.name AS "category.name",
	c.icon AS "category.icon"`

type BudgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) Create(ctx context.Context, budget *models.Budget) (*dto.BudgetResponse, error) {
	const q = `
		WITH inserted AS (
			INSERT INTO budgets (
				name, amount, balance, user_id, currency, interval_type, "interval", start_date, reset_date, category_id
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, name, amount, balance, currency, interval_type, "interval", start_date, reset_date, category_id
		)
		SELECT ` + budgetSelectColumns + `
		FROM inserted b
		JOIN categories c ON c.id = b.category_id`

	var resp dto.BudgetResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		budget.Name,
		budget.Amount,
		budget.Amount,
		budget.UserID,
		budget.Currency,
		budget.IntervalType,
		nullInt(budget.Interval),
		budget.StartDate,
		budget.ResetDate,
		budget.CategoryID,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *BudgetRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.BudgetResponse, error) {
	const q = `
		SELECT ` + budgetSelectColumns + `
		FROM budgets b
		JOIN categories c ON c.id = b.category_id
		WHERE b.id = $1 AND b.user_id = $2`

	var resp dto.BudgetResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id, userID); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *BudgetRepository) List(ctx context.Context, userID uuid.UUID) ([]dto.BudgetResponse, error) {
	const q = `
		SELECT ` + budgetSelectColumns + `
		FROM budgets b
		JOIN categories c ON c.id = b.category_id
		WHERE b.user_id = $1
		ORDER BY c.name`

	var resp []dto.BudgetResponse
	if err := sqlscan.Select(ctx, r.db, &resp, q, userID); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *BudgetRepository) Update(ctx context.Context, budget *models.Budget) (*dto.BudgetResponse, error) {
	const q = `
		WITH updated AS (
			UPDATE budgets
			SET name = COALESCE($2, name),
			    amount = COALESCE($3::numeric, amount),
			    interval_type = COALESCE($4::interval_type, interval_type),
			    "interval" = COALESCE($5::int, "interval"),
			    start_date = COALESCE($6::date, start_date),
			    reset_date = COALESCE($7::date, reset_date)
			WHERE id = $1
			RETURNING id, name, amount, balance, currency, interval_type, "interval", start_date, reset_date, category_id
		)
		SELECT ` + budgetSelectColumns + `
		FROM updated b
		JOIN categories c ON c.id = b.category_id`

	var resp dto.BudgetResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		budget.ID,
		nullString(budget.Name),
		nullDecimal(budget.Amount),
		nullString(string(budget.IntervalType)),
		nullInt(budget.Interval),
		nullTime(budget.StartDate),
		nullTime(budget.ResetDate),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *BudgetRepository) GetByCategoryID(ctx context.Context, userID, categoryID uuid.UUID, tx *sql.Tx) (uuid.UUID, error) {
	const q = `
		SELECT id
		FROM budgets
		WHERE user_id = $1 AND category_id = $2
		FOR UPDATE`

	var id uuid.UUID
	if err := tx.QueryRowContext(ctx, q, userID, categoryID).Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *BudgetRepository) UpdateBalance(ctx context.Context, id uuid.UUID, delta decimal.Decimal, tx *sql.Tx) error {
	const q = `UPDATE budgets SET balance = balance + $2 WHERE id = $1`
	_, err := tx.ExecContext(ctx, q, id, delta)
	return err
}

func (r *BudgetRepository) BulkUpdate(ctx context.Context, budgets []models.Budget) error {
	if len(budgets) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ids := make([]string, len(budgets))
	balances := make([]string, len(budgets))
	resetDates := make([]string, len(budgets))

	for i, budget := range budgets {
		ids[i] = budget.ID.String()
		balances[i] = budget.Balance.String()
		resetDates[i] = budget.ResetDate.Format(time.DateOnly)
	}

	const q = `
		UPDATE budgets AS b
		SET balance = v.balance::numeric,
		    reset_date = v.reset_date::date
		FROM unnest($1::text[], $2::text[], $3::text[]) AS v(id, balance, reset_date)
		WHERE b.id = v.id::uuid`

	if _, err = tx.ExecContext(ctx, q, ids, balances, resetDates); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BudgetRepository) GetByResetDate(ctx context.Context, resetDate time.Time) ([]dto.BudgetResponse, error) {
	const q = `
		SELECT ` + budgetSelectColumns + `
		FROM budgets b
		JOIN categories c ON c.id = b.category_id
		WHERE b.reset_date <= $1`
	var resp []dto.BudgetResponse
	if err := sqlscan.Select(ctx, r.db, &resp, q, resetDate); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *BudgetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM budgets WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
