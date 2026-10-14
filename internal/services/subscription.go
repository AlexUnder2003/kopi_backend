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
	"go.uber.org/zap"
)

const errMsgSubscriptionNotFound = "not_found_subscription"

type SubscriptionService struct {
	subscriptionRepo *repositories.SubscriptionRepository
	logger           *zap.SugaredLogger
}

func NewSubscriptionService(db *sql.DB, logger *zap.SugaredLogger) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepo: repositories.NewSubscriptionRepository(db),
		logger:           logger,
	}
}

func (s *SubscriptionService) HasAccess(ctx context.Context, userID uuid.UUID, plan models.SubscriptionPlan) (bool, error) {
	subscription, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	switch plan {
	case models.SubscriptionPlanFree:
		return true, nil
	case models.SubscriptionPlanPro:
		return subscription.Plan == models.SubscriptionPlanPro || subscription.Plan == models.SubscriptionPlanProPlus, nil
	case models.SubscriptionPlanProPlus:
		return subscription.Plan == models.SubscriptionPlanProPlus, nil
	}

	return false, nil
}

func (s *SubscriptionService) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	subscription, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgSubscriptionNotFound)
		}
		s.logger.Errorw("failed to get subscription", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return subscription, nil
}

func (s *SubscriptionService) Create(ctx context.Context, subscription *models.Subscription) (*dto.SubscriptionResponse, error) {
	created, err := s.subscriptionRepo.Create(ctx, subscription)
	if err != nil {
		s.logger.Errorw("failed to create subscription", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return created, nil
}

func (s *SubscriptionService) Update(ctx context.Context, userID uuid.UUID, subscription *models.Subscription) (*dto.SubscriptionResponse, error) {
	updated, err := s.subscriptionRepo.Update(ctx, userID, subscription)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgSubscriptionNotFound)
		}
		s.logger.Errorw("failed to update subscription", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return updated, nil
}
