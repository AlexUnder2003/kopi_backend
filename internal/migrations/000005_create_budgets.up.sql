CREATE TYPE budget_frequency AS ENUM ('daily', 'weekly', 'monthly', 'yearly');

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    amount NUMERIC(19, 4) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    currency currency_code NOT NULL,
    frequency budget_frequency NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budgets_user_id ON budgets(user_id);
CREATE UNIQUE INDEX idx_budgets_category_id_frequency_user_id ON budgets(category_id, frequency, user_id);
CREATE TRIGGER update_budgets_updated_at BEFORE UPDATE ON budgets FOR EACH ROW EXECUTE FUNCTION update_updated_at();