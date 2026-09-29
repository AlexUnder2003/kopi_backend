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
		ResetDate:    NextDate(budget.StartDate, budget.IntervalType, budget.Interval),
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
		ID:       existingBudget.ID,
		Name:     budget.Name,
		Amount:   budget.Amount,
		IsActive: budget.IsActive,
	}

	if budgetScheduleChanged(budget) {
		intervalType := effectiveIntervalType(budget.IntervalType, existingBudget.IntervalType)
		interval := effectiveInterval(budget.Interval, existingInterval(existingBudget.Interval))
		startDate := effectiveTime(budget.StartDate, existingBudget.StartDate)

		if intervalType != models.IntervalTypeCustom {
			startDate = s.CalculateStartDate()
		}

		budgetModel.IntervalType = intervalType
		budgetModel.Interval = interval
		budgetModel.StartDate = startDate
		budgetModel.ResetDate = NextDate(startDate, intervalType, interval)
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
		budgetModels[i] = models.Budget{
			ID:        budget.ID,
			Balance:   budget.Amount,
			ResetDate: NextDate(budget.ResetDate, budget.IntervalType, existingInterval(budget.Interval)),
		}
	}

	if err := s.budgetRepo.BulkReset(ctx, budgetModels); err != nil {
		s.logger.Errorw("failed to bulk update budgets", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	return nil
}

func budgetScheduleChanged(budget *dto.BudgetUpdate) bool {
	return budget.IntervalType != "" || budget.Interval != 0 || !budget.StartDate.IsZero()
}

func existingInterval(interval *int) int {
	if interval == nil {
		return 0
	}
	return *interval
}

func (s *BudgetService) CalculateStartDate() time.Time {
	now := time.Now()
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	return firstDayOfMonth
}
