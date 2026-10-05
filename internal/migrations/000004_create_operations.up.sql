CREATE TYPE operation_type AS ENUM ('income', 'expense', 'transfer');

CREATE TABLE operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type operation_type NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    from_account_id UUID REFERENCES accounts(id),
    category_id UUID NOT NULL REFERENCES categories(id),
    amount NUMERIC(19, 4) NOT NULL,
    occurrence_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_operations_account_occurrence_date ON operations(account_id, occurrence_date DESC);
CREATE INDEX idx_operations_from_account_id ON operations(from_account_id);
CREATE INDEX idx_operations_category_id ON operations(category_id);
CREATE TRIGGER update_operations_updated_at BEFORE UPDATE ON operations FOR EACH ROW EXECUTE FUNCTION update_updated_at();