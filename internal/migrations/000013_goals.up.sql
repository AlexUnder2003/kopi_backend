CREATE TABLE goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    currency currency_code NOT NULL,
    deadline TIMESTAMP,
    target_amount DECIMAL(10, 2) NOT NULL,
    current_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_goals_user_id ON goals(user_id);

ALTER TABLE operations
    ADD CONSTRAINT operations_goal_id_fkey
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE SET NULL;