CREATE TYPE interval_type AS ENUM ('daily', 'weekly', 'biweekly', 'monthly', 'custom');

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    amount NUMERIC(19, 4) NOT NULL,
    balance NUMERIC(19, 4) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    category_id UUID NOT NULL REFERENCES categories(id),
    currency currency_code NOT NULL,
    interval_type interval_type NOT NULL,
    "interval" INT,
    start_date DATE NOT NULL,
    reset_date DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budgets_user_id ON budgets(user_id);
CREATE UNIQUE INDEX idx_budgets_user_id_category_id ON budgets(user_id, category_id);
CREATE TRIGGER update_budgets_updated_at BEFORE UPDATE ON budgets FOR EACH ROW EXECUTE FUNCTION update_updated_at();
