package repositories

import (
	"context"
	"database/sql"
	"strconv"
	"time"

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

func (r *AIExecutionRepository) Create(ctx context.Context, tx *sql.Tx, execution *models.AIExecution) (*models.AIExecution, error) {
	const q = `
		INSERT INTO ai_executions (user_id, left_quota, next_reset_at)
		VALUES ($1, $2, $3)
		RETURNING ` + aiExecutionSelectColumns

	var created models.AIExecution
	if err := sqlscan.Get(
		ctx, tx, &created, q,
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

func (r *AIExecutionRepository) GetByNextResetAt(ctx context.Context, tx *sql.Tx) ([]models.AIExecution, error) {
	const q = `
		SELECT ` + aiExecutionSelectColumns + `
		FROM ai_executions
		WHERE next_reset_at <= CURRENT_TIMESTAMP
		FOR UPDATE SKIP LOCKED`

	var executions []models.AIExecution
	if err := sqlscan.Select(ctx, tx, &executions, q); err != nil {
		return nil, err
	}
	return executions, nil
}

func (r *AIExecutionRepository) BulkUpdate(ctx context.Context, executions []models.AIExecution, tx *sql.Tx) error {
	if len(executions) == 0 {
		return nil
	}

	ids := make([]string, len(executions))
	userIDs := make([]string, len(executions))
	leftQuotas := make([]string, len(executions))
	nextResetAts := make([]string, len(executions))

	for i, execution := range executions {
		ids[i] = execution.ID.String()
		userIDs[i] = execution.UserID.String()
		leftQuotas[i] = strconv.Itoa(execution.LeftQuota)
		nextResetAts[i] = execution.NextResetAt.UTC().Format(time.DateTime)
	}

	const q = `
		UPDATE ai_executions AS a
		SET left_quota = v.left_quota::int,
		    next_reset_at = v.next_reset_at::timestamp,
		    updated_at = CURRENT_TIMESTAMP
		FROM unnest($1::text[], $2::text[], $3::text[], $4::text[]) AS v(id, user_id, left_quota, next_reset_at)
		WHERE a.id = v.id::uuid AND a.user_id = v.user_id::uuid`

	_, err := DBorTx(r.db, tx).ExecContext(ctx, q, ids, userIDs, leftQuotas, nextResetAts)
	return err
}
