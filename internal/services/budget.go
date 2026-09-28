package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"KopiBackend/internal/apperrors"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

const (
	errMsgBudgetNotFound      = "not_found_budget"
	errMsgBudgetAlreadyExists = "already_exists_budget"
)

type BudgetService struct {
	budgetRepo      *repositories.BudgetRepository
	categoryService *CategoryService
	logger          *zap.SugaredLogger
}

func NewBudgetService(db *sql.DB, categoryService *CategoryService, logger *zap.SugaredLogger) *BudgetService {
	return &BudgetService{
		budgetRepo:      repositories.NewBudgetRepository(db),
		categoryService: categoryService,
		logger:          logger,
	}
}

func (s *BudgetService) Create(ctx context.Context, userID uuid.UUID, budget *dto.BudgetPost) (*dto.BudgetResponse, error) {
	if budget.IntervalType != models.IntervalTypeCustom {
		budget.StartDate = s.CalculateStartDate()
	}

	budgetModel := &models.Budget{
		Name:         budget.Name,
		UserID:       userID,
		Amount:       budget.Amount,
		IntervalType: budget.IntervalType,
		Interval:     budget.Interval,
		StartDate:    budget.StartDate,
		ResetDate:    s.CalculateResetDate(ctx, budget.StartDate, budget.IntervalType, budget.Interval),
		Currency:     budget.Currency,
		CategoryID:   budget.CategoryID,
	}

	created, err := s.budgetRepo.Create(ctx, budgetModel)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.Conflict(errMsgBudgetAlreadyExists)
		}
		s.logger.Errorw("failed to create budget", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return created, nil
}

func (s *BudgetService) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.BudgetResponse, error) {
	budget, err := s.budgetRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgBudgetNotFound)
		}
		s.logger.Errorw("failed to get budget", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return budget, nil
}

func (s *BudgetService) List(ctx context.Context, userID uuid.UUID) ([]dto.BudgetResponse, error) {
	budgets, err := s.budgetRepo.List(ctx, userID)
	if err != nil {
		s.logger.Errorw("failed to list budgets", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return budgets, nil
}

func (s *BudgetService) Update(ctx context.Context, id, userID uuid.UUID, budget *dto.BudgetUpdate) (*dto.BudgetResponse, error) {
	existingBudget, err := s.budgetRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgBudgetNotFound)
		}
		s.logger.Errorw("failed to get budget", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	budgetModel := &models.Budget{
		ID:     existingBudget.ID,
		Name:   budget.Name,
		Amount: budget.Amount,
	}

	if budget.IntervalType != existingBudget.IntervalType || budget.IntervalType == models.IntervalTypeCustom {
		if budget.IntervalType == models.IntervalTypeCustom {
			budgetModel.StartDate = budget.StartDate
			budgetModel.Interval = budget.Interval
		} else {
			budgetModel.StartDate = s.CalculateStartDate()
		}

		budgetModel.IntervalType = budget.IntervalType
		budgetModel.ResetDate = s.CalculateResetDate(ctx, budgetModel.StartDate, budget.IntervalType, budgetModel.Interval)
	}

	updated, err := s.budgetRepo.Update(ctx, budgetModel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgBudgetNotFound)
		}
		s.logger.Errorw("failed to update budget", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return updated, nil
}

func (s *BudgetService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.GetByID(ctx, id, userID); err != nil {
		return err
	}

	if err := s.budgetRepo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete budget", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	return nil
}

func (s *BudgetService) Reset(ctx context.Context) error {
	budgets, err := s.budgetRepo.GetByResetDate(ctx, time.Now())
	if err != nil {
		s.logger.Errorw("failed to get budgets by reset date", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	if len(budgets) == 0 {
		return nil
	}

	budgetModels := make([]models.Budget, len(budgets))
	for i, budget := range budgets {
		var interval int

		if budget.Interval != nil {
			interval = *budget.Interval
		}

		budgetModels[i] = models.Budget{
			ID:        budget.ID,
			Balance:   budget.Amount,
			ResetDate: s.CalculateResetDate(ctx, budget.ResetDate, budget.IntervalType, interval),
		}
	}

	if err := s.budgetRepo.BulkUpdate(ctx, budgetModels); err != nil {
		s.logger.Errorw("failed to bulk update budgets", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	return nil
}

func (s *BudgetService) CalculateResetDate(ctx context.Context, date time.Time, intervalType models.IntervalType, interval int) time.Time {
	var resetDate time.Time

	now := time.Now()

	for {
		switch intervalType {
		case models.IntervalTypeDaily:
			resetDate = date.AddDate(0, 0, 1)
		case models.IntervalTypeWeekly:
			resetDate = date.AddDate(0, 0, 7)
		case models.IntervalTypeBiweekly:
			resetDate = date.AddDate(0, 0, 14)
		case models.IntervalTypeMonthly:
			resetDate = date.AddDate(0, 1, 0)
		case models.IntervalTypeCustom:
			resetDate = date.AddDate(0, 0, interval)
		}
		if resetDate.After(now) {
			return resetDate
		}
		date = resetDate
	}
}

func (s *BudgetService) CalculateStartDate() time.Time {
	now := time.Now()
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	return firstDayOfMonth
}
