ALTER TABLE budgets
    ADD COLUMN category_id UUID REFERENCES categories(id);

UPDATE budgets AS b
SET category_id = picked.category_id
FROM (
    SELECT DISTINCT ON (budget_id) budget_id, category_id
    FROM budget_categories
    ORDER BY budget_id, category_id
) AS picked
WHERE picked.budget_id = b.id;

ALTER TABLE budgets
    ALTER COLUMN category_id SET NOT NULL;

CREATE UNIQUE INDEX idx_budgets_user_id_category_id ON budgets(user_id, category_id);

DROP TABLE budget_categories;

ALTER TABLE budgets DROP CONSTRAINT budgets_id_user_id_key;
