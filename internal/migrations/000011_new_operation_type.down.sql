DROP INDEX idx_operations_planned_operation_id;

ALTER TABLE operations DROP COLUMN planned_operation_id;

ALTER TYPE operation_type RENAME TO operation_type_old;

CREATE TYPE operation_type AS ENUM ('income', 'expense', 'transfer');

ALTER TABLE operations
    ALTER COLUMN type TYPE operation_type
    USING type::text::operation_type;

ALTER TABLE planned_operations
    ALTER COLUMN type TYPE operation_type
    USING type::text::operation_type;

DROP TYPE operation_type_old;
