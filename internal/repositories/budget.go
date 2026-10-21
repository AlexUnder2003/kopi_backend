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
	b.is_active`

const budgetCategoriesColumn = `
	COALESCE(
		jsonb_agg(
			jsonb_build_object(
				'id', c.id,
				'name', c.name,
				'icon', c.icon,
				'is_system', c.user_id IS NULL
			)
			ORDER BY c.name
		) FILTER (WHERE c.id IS NOT NULL),
		'[]'::jsonb
	) AS categories`

const budgetCategoryJoins = `
	LEFT JOIN budget_categories bc ON bc.budget_id = b.id
	LEFT JOIN categories c ON c.id = bc.category_id`

type BudgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) Create(ctx context.Context, budget *models.Budget) (*dto.BudgetResponse, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	const insertBudget = `
		INSERT INTO budgets (
			name, amount, balance, user_id, currency, interval_type, "interval", start_date, reset_date
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, amount, balance, interval_type, "interval", currency, start_date, reset_date, is_active`

	var resp dto.BudgetResponse
	if err := sqlscan.Get(
		ctx, tx, &resp, insertBudget,
		budget.Name,
		budget.Amount,
		budget.Balance,
		budget.UserID,
		budget.Currency,
		budget.IntervalType,
		nullInt(budget.Interval),
		budget.StartDate,
		budget.ResetDate,
	); err != nil {
		return nil, err
	}

	const insertCategories = `
		WITH linked AS (
			INSERT INTO budget_categories (budget_id, user_id, category_id)
			SELECT $1, $2, category_id
			FROM unnest($3::uuid[]) AS category_id
			RETURNING category_id
		)
		SELECT ` + budgetCategoriesColumn + `
		FROM linked l
		JOIN categories c ON c.id = l.category_id`

	if err := tx.QueryRowContext(ctx, insertCategories, resp.ID, budget.UserID, budget.CategoryIDs).Scan(&resp.Categories); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *BudgetRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.BudgetResponse, error) {
	const q = `
		SELECT ` + budgetSelectColumns + `,
			` + budgetCategoriesColumn + `
		FROM budgets b
		` + budgetCategoryJoins + `
		WHERE b.id = $1 AND b.user_id = $2
		GROUP BY b.id`

	var resp dto.BudgetResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id, userID); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *BudgetRepository) List(ctx context.Context, userID uuid.UUID, params ListParams) ([]dto.BudgetResponse, int, error) {
	const q = `
		WITH page AS (
			SELECT ` + budgetSelectColumns + `
			FROM budgets b
			WHERE b.user_id = $1
				AND (b.name ILIKE '%' || $2 || '%' OR $2 IS NULL)
			ORDER BY b.name
			LIMIT $3 OFFSET $4
		)
		SELECT ` + budgetSelectColumns + `,
			` + budgetCategoriesColumn + `
		FROM page b
		` + budgetCategoryJoins + `
		GROUP BY b.id, b.name, b.amount, b.balance, b.interval_type, b."interval", b.currency, b.start_date, b.reset_date, b.is_active
		ORDER BY b.name`

	var resp []dto.BudgetResponse
	if err := sqlscan.Select(ctx, r.db, &resp, q, userID, nullString(params.Filter), nullInt(params.Limit), nullInt(params.Offset)); err != nil {
		return nil, 0, err
	}

	const countQuery = `
		SELECT COUNT(*)
		FROM budgets
		WHERE user_id = $1
			AND (name ILIKE '%' || $2 || '%' OR $2 IS NULL)`

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, userID, nullString(params.Filter)).Scan(&total); err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}

func (r *BudgetRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	const q = `SELECT COUNT(*) FROM budgets WHERE user_id = $1`

	var count int
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *BudgetRepository) GetAlreadySpent(ctx context.Context, userID uuid.UUID, categoryIDs []uuid.UUID, startDate time.Time, endDate time.Time, currency string) (decimal.Decimal, error) {
	const q = `
		SELECT COALESCE(SUM(o.amount), 0)
		FROM operations o
		JOIN accounts a ON a.id = o.account_id
		WHERE a.user_id = $1
			AND o.category_id = ANY($2::uuid[])
			AND o.type = 'expense'
			AND o.occurrence_date >= $3
			AND o.occurrence_date <= $4
			AND a.currency = $5`

	var amount decimal.Decimal
	if err := r.db.QueryRowContext(ctx, q, userID, categoryIDs, startDate, endDate, currency).Scan(&amount); err != nil {
		return decimal.Zero, err
	}

	return amount, nil
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
			    reset_date = COALESCE($7::date, reset_date),
			    is_active = COALESCE($8::boolean, is_active)
			WHERE id = $1
			RETURNING id, name, amount, balance, currency, interval_type, "interval", start_date, reset_date, is_active
		)
		SELECT ` + budgetSelectColumns + `,
			` + budgetCategoriesColumn + `
		FROM updated b
		` + budgetCategoryJoins + `
		GROUP BY b.id, b.name, b.amount, b.balance, b.interval_type, b."interval", b.currency, b.start_date, b.reset_date, b.is_active`

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
		budget.IsActive,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *BudgetRepository) GetByCategoryID(ctx context.Context, userID, categoryID uuid.UUID, currency string, tx *sql.Tx) (uuid.UUID, error) {
	const q = `
		SELECT b.id
		FROM budgets b
		JOIN budget_categories bc ON bc.budget_id = b.id
		WHERE b.user_id = $1 AND bc.category_id = $2 AND b.currency = $3
			AND b.is_active = TRUE
		FOR UPDATE OF b SKIP LOCKED`

	var id uuid.UUID
	if err := tx.QueryRowContext(ctx, q, userID, categoryID, currency).Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *BudgetRepository) BulkUpdateBalance(ctx context.Context, updates []dto.BudgetBalanceUpdate, tx *sql.Tx) error {
	if len(updates) == 0 {
		return nil
	}

	userIDs := make([]string, len(updates))
	categoryIDs := make([]string, len(updates))
	currencies := make([]string, len(updates))
	deltas := make([]string, len(updates))

	for i, update := range updates {
		userIDs[i] = update.UserID.String()
		categoryIDs[i] = update.CategoryID.String()
		currencies[i] = update.Currency
		deltas[i] = update.Delta.String()
	}

	const q = `
		UPDATE budgets AS b
		SET balance = b.balance + v.delta
		FROM (
			SELECT bc.budget_id, u.currency, SUM(u.delta::numeric) AS delta
			FROM unnest($1::text[], $2::text[], $3::text[], $4::text[]) AS u(user_id, category_id, currency, delta)
			JOIN budget_categories bc
				ON bc.user_id = u.user_id::uuid
				AND bc.category_id = u.category_id::uuid
			GROUP BY bc.budget_id, u.currency
		) AS v
		WHERE b.id = v.budget_id
			AND b.currency::text = v.currency`

	_, err := DBorTx(r.db, tx).ExecContext(ctx, q, userIDs, categoryIDs, currencies, deltas)
	return err
}

func (r *BudgetRepository) UpdateBalance(ctx context.Context, id uuid.UUID, delta decimal.Decimal, tx *sql.Tx) error {
	const q = `UPDATE budgets SET balance = balance + $2 WHERE id = $1`
	_, err := tx.ExecContext(ctx, q, id, delta)
	return err
}

func (r *BudgetRepository) BulkReset(ctx context.Context, budgets []models.Budget, tx *sql.Tx) error {
	if len(budgets) == 0 {
		return nil
	}

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

	_, err := tx.ExecContext(ctx, q, ids, balances, resetDates)
	return err
}

func (r *BudgetRepository) GetByResetDate(ctx context.Context, resetDate time.Time, tx *sql.Tx) ([]dto.BudgetResponse, error) {
	const q = `
		SELECT ` + budgetSelectColumns + `
		FROM budgets b
		WHERE b.reset_date <= $1 AND b.is_active = TRUE
		FOR UPDATE OF b SKIP LOCKED`

	var resp []dto.BudgetResponse
	if err := sqlscan.Select(ctx, tx, &resp, q, resetDate); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *BudgetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM budgets WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
