ALTER TYPE operation_type ADD VALUE 'reservation';

ALTER TABLE operations
    ADD COLUMN planned_operation_id UUID REFERENCES planned_operations(id) ON DELETE CASCADE;

CREATE INDEX idx_operations_planned_operation_id ON operations(planned_operation_id);
