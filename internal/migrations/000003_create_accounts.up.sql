CREATE TYPE currency_code AS ENUM ('USD', 'EUR', 'RUB', 'KZT', 'BYN');

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    currency currency_code NOT NULL,
    icon VARCHAR(255) NOT NULL,
    balance NUMERIC(19, 4) NOT NULL DEFAULT 0,
    user_id UUID NOT NULL REFERENCES users(id),
    include_in_free_balance BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE TRIGGER update_accounts_updated_at BEFORE UPDATE ON accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at();