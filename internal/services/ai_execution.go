package services

import (
	"context"
	"database/sql"
	"time"

	"KopiBackend/internal/apperrors"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"

	"go.uber.org/zap"
)

type AIExecutionService struct {
	db              *sql.DB
	aiExecutionRepo *repositories.AIExecutionRepository
	logger          *zap.SugaredLogger
}

func NewAIExecutionService(db *sql.DB, logger *zap.SugaredLogger) *AIExecutionService {
	return &AIExecutionService{
		db:              db,
		aiExecutionRepo: repositories.NewAIExecutionRepository(db),
		logger:          logger,
	}
}

func (s *AIExecutionService) ResetQuota(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Errorw("failed to begin transaction", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	defer tx.Rollback()

	executions, err := s.aiExecutionRepo.GetByNextResetAt(ctx, tx)
	if err != nil {
		s.logger.Errorw("failed to get ai executions by next reset", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	if len(executions) == 0 {
		return nil
	}

	nextResetAt := time.Now().Add(time.Hour * 24 * 30)
	updated := make([]models.AIExecution, len(executions))
	for i, execution := range executions {
		updated[i] = models.AIExecution{
			ID:          execution.ID,
			UserID:      execution.UserID,
			LeftQuota:   defaultAIExecutionQuota,
			NextResetAt: nextResetAt,
		}
	}

	if err := s.aiExecutionRepo.BulkUpdate(ctx, updated, tx); err != nil {
		s.logger.Errorw("failed to bulk update ai executions", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorw("failed to commit transaction", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	return nil
}
