ALTER TABLE budgets
    ADD CONSTRAINT budgets_id_user_id_key UNIQUE (id, user_id);

CREATE TABLE budget_categories (
    budget_id UUID NOT NULL,
    user_id UUID NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    PRIMARY KEY (budget_id, category_id),
    UNIQUE (user_id, category_id),
    FOREIGN KEY (budget_id, user_id) REFERENCES budgets (id, user_id) ON DELETE CASCADE
);

INSERT INTO budget_categories (budget_id, user_id, category_id)
SELECT id, user_id, category_id
FROM budgets;

DROP INDEX idx_budgets_user_id_category_id;

ALTER TABLE budgets DROP COLUMN category_id;
