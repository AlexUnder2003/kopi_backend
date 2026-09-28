DROP INDEX IF EXISTS idx_budgets_category_id_interval_user_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_budgets_user_id_category_id ON budgets(user_id, category_id);
