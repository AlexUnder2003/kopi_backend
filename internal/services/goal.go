package services

import (
	"context"
	"database/sql"
	"errors"

	"KopiBackend/internal/apperrors"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const errMsgGoalNotFound = "not_found_goal"

type GoalService struct {
	goalRepo *repositories.GoalRepository
	logger   *zap.SugaredLogger
}

func NewGoalService(db *sql.DB, logger *zap.SugaredLogger) *GoalService {
	return &GoalService{
		goalRepo: repositories.NewGoalRepository(db),
		logger:   logger,
	}
}

func (s *GoalService) Create(ctx context.Context, userID uuid.UUID, goal *dto.GoalPost) (*dto.GoalResponse, error) {
	currentAmount := goal.CurrentAmount
	created, err := s.goalRepo.Create(ctx, &models.Goal{
		UserID:        userID,
		Name:          goal.Name,
		Currency:      models.CurrencyCode(goal.Currency),
		Deadline:      goal.Deadline,
		TargetAmount:  goal.TargetAmount,
		CurrentAmount: &currentAmount,
	})
	if err != nil {
		s.logger.Errorw("failed to create goal", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return created, nil
}

func (s *GoalService) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.GoalResponse, error) {
	goal, err := s.goalRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgGoalNotFound)
		}
		s.logger.Errorw("failed to get goal", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return goal, nil
}

func (s *GoalService) List(ctx context.Context, userID uuid.UUID, params repositories.ListParams) (*dto.GoalResponseList, error) {
	goals, total, err := s.goalRepo.List(ctx, userID, params)
	if err != nil {
		s.logger.Errorw("failed to list goals", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if goals == nil {
		goals = []dto.GoalResponse{}
	}

	return &dto.GoalResponseList{
		Goals:      goals,
		TotalCount: total,
	}, nil
}

func (s *GoalService) Update(ctx context.Context, id, userID uuid.UUID, goal *dto.GoalUpdate) (*dto.GoalResponse, error) {
	if _, err := s.GetByID(ctx, id, userID); err != nil {
		return nil, err
	}

	updated, err := s.goalRepo.Update(ctx, &models.Goal{
		ID:           id,
		UserID:       userID,
		Name:         goal.Name,
		Deadline:     goal.Deadline,
		TargetAmount: goal.TargetAmount,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgGoalNotFound)
		}
		s.logger.Errorw("failed to update goal", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return updated, nil
}

func (s *GoalService) UpdateBalance(ctx context.Context, id uuid.UUID, delta decimal.Decimal, tx *sql.Tx) error {
	if err := s.goalRepo.UpdateBalance(ctx, id, delta, tx); err != nil {
		s.logger.Errorw("failed to update goal balance", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	return nil
}

func (s *GoalService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.GetByID(ctx, id, userID); err != nil {
		return err
	}

	if err := s.goalRepo.Delete(ctx, id, userID); err != nil {
		s.logger.Errorw("failed to delete goal", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	return nil
}
