CREATE TABLE planned_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount NUMERIC(19, 4) NOT NULL,
    type transaction_type NOT NULL,
    interval_type interval_type NOT NULL,
    "interval" INT,
    category_id UUID NOT NULL REFERENCES categories(id),
    planned_at DATE NOT NULL,
    next_run_at DATE,
    is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_planned_operations_next_run_at ON planned_operations(next_run_at);
CREATE INDEX idx_planned_operations_account_id ON planned_operations(account_id);

CREATE TRIGGER update_planned_operations_updated_at BEFORE UPDATE ON planned_operations FOR EACH ROW EXECUTE FUNCTION update_updated_at();