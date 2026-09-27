CREATE TYPE transaction_type AS ENUM ('income', 'expense', 'transfer_in', 'transfer_out');

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type transaction_type NOT NULL,
    transfer_id UUID,
    account_id UUID NOT NULL REFERENCES accounts(id),
    category_id UUID NOT NULL REFERENCES categories(id),
    amount NUMERIC(19, 4) NOT NULL,
    occurrence_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_account_occurrence_date ON transactions(account_id, occurrence_date DESC);
CREATE INDEX idx_transactions_category_id ON transactions(category_id);
CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions FOR EACH ROW EXECUTE FUNCTION update_updated_at();