package repositories

import (
	"context"
	"database/sql"

	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
)

const aiExecutionSelectColumns = `
	id,
	user_id,
	left_quota,
	next_reset_at`

type AIExecutionRepository struct {
	db *sql.DB
}

func NewAIExecutionRepository(db *sql.DB) *AIExecutionRepository {
	return &AIExecutionRepository{db: db}
}

func (r *AIExecutionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.AIExecution, error) {
	const q = `
		SELECT ` + aiExecutionSelectColumns + `
		FROM ai_executions
		WHERE user_id = $1`

	var execution models.AIExecution
	if err := sqlscan.Get(ctx, r.db, &execution, q, userID); err != nil {
		return nil, err
	}
	return &execution, nil
}

func (r *AIExecutionRepository) Create(ctx context.Context, execution *models.AIExecution) (*models.AIExecution, error) {
	const q = `
		INSERT INTO ai_executions (user_id, left_quota, next_reset_at)
		VALUES ($1, $2, $3)
		RETURNING ` + aiExecutionSelectColumns

	var created models.AIExecution
	if err := sqlscan.Get(
		ctx, r.db, &created, q,
		execution.UserID, execution.LeftQuota, execution.NextResetAt,
	); err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *AIExecutionRepository) Update(ctx context.Context, execution *models.AIExecution) (*models.AIExecution, error) {
	const q = `
		UPDATE ai_executions
		SET left_quota = $2,
		    next_reset_at = $3,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $4
		RETURNING ` + aiExecutionSelectColumns

	var updated models.AIExecution
	if err := sqlscan.Get(
		ctx, r.db, &updated, q,
		execution.ID, execution.LeftQuota, execution.NextResetAt, execution.UserID,
	); err != nil {
		return nil, err
	}
	return &updated, nil
}
