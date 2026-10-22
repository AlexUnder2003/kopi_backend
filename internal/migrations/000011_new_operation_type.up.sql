ALTER TYPE operation_type ADD VALUE 'reservation';

ALTER TABLE operations
    ADD COLUMN planned_operation_id UUID REFERENCES planned_operations(id) ON DELETE CASCADE,
    ADD COLUMN goal_id UUID;

CREATE INDEX idx_operations_planned_operation_id ON operations(planned_operation_id);
CREATE INDEX idx_operations_goal_id ON operations(goal_id);