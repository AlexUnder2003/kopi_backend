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
	b.amount AS total_amount,
	COALESCE(spent.amount_spent, 0) AS amount_spent,
	b.amount - COALESCE(spent.amount_spent, 0) AS amount_left,
	b.frequency,
	b.currency,
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
			INSERT INTO budgets (amount, user_id, currency, frequency, category_id)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, amount, user_id, currency, frequency, category_id
		)
		SELECT ` + budgetSelectColumns + `
		FROM inserted b
		JOIN categories c ON c.id = b.category_id`

	var resp dto.BudgetResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		budget.Amount, budget.UserID, budget.Currency, budget.Frequency, budget.CategoryID,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *BudgetRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.BudgetResponse, error) {
	const q = `
		SELECT ` + budgetSelectColumns + `
		FROM budgets b
		JOIN categories c ON c.id = b.category_id
		WHERE b.id = $1`

	var resp dto.BudgetResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id); err != nil {
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
			    currency = COALESCE($3, currency),
			    frequency = COALESCE($4::budget_frequency, frequency),
			    category_id = COALESCE($5, category_id)
			WHERE id = $1
			RETURNING id, amount, user_id, currency, frequency, category_id
		)
		SELECT ` + budgetSelectColumns + `
		FROM updated b
		JOIN categories c ON c.id = b.category_id`

	var resp dto.BudgetResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		budget.ID,
		nullDecimal(budget.Amount),
		nullString(budget.Currency),
		nullString(string(budget.Frequency)),
		nullUUID(budget.CategoryID),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *BudgetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM budgets WHERE id = $1;`

	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
