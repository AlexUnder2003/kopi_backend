package repositories

import (
	"context"
	"database/sql"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
)

const budgetSelectColumns = `
	b.id,
	b.amount,
	b.balance,
	b.interval_type,
	b."interval",
	b.currency,
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
				amount, balance, user_id, currency, interval_type, "interval", reset_date, category_id
			)
			VALUES ($1, COALESCE($2, $1), $3, $4, $5, $6, $7, $8)
			RETURNING id, amount, balance, user_id, currency, interval_type, "interval", reset_date, category_id
		)
		SELECT ` + budgetSelectColumns + `
		FROM inserted b
		JOIN categories c ON c.id = b.category_id`

	var resp dto.BudgetResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		budget.Amount,
		nullDecimal(budget.Balance),
		budget.UserID,
		budget.Currency,
		budget.IntervalType,
		budget.Interval,
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
			SET amount = COALESCE($2, amount),
			    balance = COALESCE($3, balance),
			    interval_type = COALESCE($4::interval_type, interval_type),
			    "interval" = COALESCE($5, "interval"),
			    reset_date = COALESCE($6, reset_date),
			    category_id = COALESCE($7, category_id)
			WHERE id = $1
			RETURNING id, amount, balance, user_id, interval_type, "interval", reset_date, category_id
		)
		SELECT ` + budgetSelectColumns + `
		FROM updated b
		JOIN categories c ON c.id = b.category_id`

	var resp dto.BudgetResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		budget.ID,
		nullDecimal(budget.Amount),
		nullDecimal(budget.Balance),
		nullString(budget.Currency),
		nullString(string(budget.IntervalType)),
		nullInt(budget.Interval),
		nullTime(budget.ResetDate),
		nullUUID(budget.CategoryID),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *BudgetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM budgets WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
