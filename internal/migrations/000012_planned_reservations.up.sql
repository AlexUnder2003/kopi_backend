ALTER TABLE planned_operations
    ADD COLUMN balance NUMERIC(19, 4) NOT NULL DEFAULT 0;
