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
	"go.uber.org/zap"
)

const errMsgPlannedOperationNotFound = "not_found_planned_operation"

type PlannedOperationService struct {
	db                   *sql.DB
	plannedOperationRepo *repositories.PlannedOperationRepository
	transactionRepo      *repositories.TransactionRepository
	accountService       *AccountService
	logger               *zap.SugaredLogger
}

func NewPlannedOperationService(db *sql.DB, accountService *AccountService, logger *zap.SugaredLogger) *PlannedOperationService {
	return &PlannedOperationService{
		db:                   db,
		plannedOperationRepo: repositories.NewPlannedOperationRepository(db),
		transactionRepo:      repositories.NewTransactionRepository(db),
		accountService:       accountService,
		logger:               logger,
	}
}

func (s *PlannedOperationService) Create(ctx context.Context, userID uuid.UUID, op *dto.PlannedOperationPost) (*dto.PlannedOperationResponse, error) {
	_, err := s.accountService.GetByID(ctx, op.AccountID, userID)
	if err != nil {
		return nil, err
	}

	isRecurring := op.IsRecurring
	model := &models.PlannedOperation{
		Name:         op.Name,
		AccountID:    op.AccountID,
		Amount:       op.Amount,
		Type:         op.Type,
		IntervalType: op.IntervalType,
		Interval:     op.Interval,
		CategoryID:   op.CategoryID,
		PlannedAt:    op.PlannedAt,
		NextRunAt:    nextRunAt(op.PlannedAt, op.IntervalType, op.Interval),
		IsRecurring:  &isRecurring,
	}

	created, err := s.plannedOperationRepo.Create(ctx, model)
	if err != nil {
		s.logger.Errorw("failed to create planned operation", "error", err, "user_id", userID)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return created, nil
}

func (s *PlannedOperationService) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.PlannedOperationResponse, error) {
	op, err := s.plannedOperationRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgPlannedOperationNotFound)
		}
		s.logger.Errorw("failed to get planned operation", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return op, nil
}

func (s *PlannedOperationService) List(ctx context.Context, userID uuid.UUID) ([]dto.PlannedOperationResponse, error) {
	ops, err := s.plannedOperationRepo.List(ctx, repositories.PlannedOperationListParams{UserID: userID})
	if err != nil {
		s.logger.Errorw("failed to list planned operations", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return ops, nil
}

func (s *PlannedOperationService) Update(ctx context.Context, id, userID uuid.UUID, op *dto.PlannedOperationUpdate) (*dto.PlannedOperationResponse, error) {
	existing, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	model := &models.PlannedOperation{
		ID:           existing.ID,
		Name:         op.Name,
		AccountID:    op.AccountID,
		Amount:       op.Amount,
		Type:         op.Type,
		IntervalType: op.IntervalType,
		Interval:     op.Interval,
		CategoryID:   op.CategoryID,
		PlannedAt:    op.PlannedAt,
		IsRecurring:  op.IsRecurring,
	}

	if scheduleChanged(op) {
		model.NextRunAt = nextRunAt(
			effectiveTime(op.PlannedAt, existing.PlannedAt),
			effectiveIntervalType(op.IntervalType, existing.IntervalType),
			effectiveInterval(op.Interval, existing.Interval),
		)
	}

	updated, err := s.plannedOperationRepo.Update(ctx, model)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgPlannedOperationNotFound)
		}
		s.logger.Errorw("failed to update planned operation", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return updated, nil
}

func (s *PlannedOperationService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.GetByID(ctx, id, userID); err != nil {
		return err
	}

	if err := s.plannedOperationRepo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete planned operation", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	return nil
}

func (s *PlannedOperationService) Execute(ctx context.Context) error {
	ops, err := s.plannedOperationRepo.GetByNextRunAt(ctx, time.Now())
	if err != nil {
		s.logger.Errorw("failed to get planned operations by next run at", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	if len(ops) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Errorw("failed to begin transaction", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	defer tx.Rollback()

	var transactions []models.Transaction

	for _, op := range ops {
		transactions = append(transactions, models.Transaction{
			Name:           op.Name,
			Type:           op.Type,
			AccountID:      op.Account.ID,
			Amount:         op.Amount,
			OccurrenceDate: op.PlannedAt,
			CategoryID:     op.Category.ID,
		})
	}

	if err := s.transactionRepo.BulkCreate(ctx, transactions, tx); err != nil {
		s.logger.Errorw("failed to bulk create transactions", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	var updatedOps []models.PlannedOperation
	for _, op := range ops {
		if op.IsRecurring {
			updatedOps = append(updatedOps, models.PlannedOperation{
				ID:        op.ID,
				NextRunAt: nextRunAt(*op.NextRunAt, op.IntervalType, op.Interval),
			})
		} else {
			updatedOps = append(updatedOps, models.PlannedOperation{
				ID: op.ID,
			})
		}
	}

	if err := s.plannedOperationRepo.BulkUpdate(ctx, updatedOps, tx); err != nil {
		s.logger.Errorw("failed to bulk update planned operations", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorw("failed to commit transaction", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	return nil
}

func nextRunAt(plannedAt time.Time, intervalType models.IntervalType, interval int) time.Time {
	if plannedAt.After(time.Now()) {
		return plannedAt
	}
	return NextDate(plannedAt, intervalType, interval)
}

func scheduleChanged(op *dto.PlannedOperationUpdate) bool {
	return !op.PlannedAt.IsZero() || op.IntervalType != "" || op.Interval != 0
}
