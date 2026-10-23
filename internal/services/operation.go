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

const errMsgOperationNotFound = "not_found_operation"

type OperationService struct {
	db                      *sql.DB
	operationRepo           *repositories.OperationRepository
	accountService          *AccountService
	budgetService           *BudgetService
	plannedOperationService *PlannedOperationService
	goalService             *GoalService
	logger                  *zap.SugaredLogger
}

func NewOperationService(db *sql.DB, accountService *AccountService, budgetService *BudgetService, plannedOperationService *PlannedOperationService, goalService *GoalService, logger *zap.SugaredLogger) *OperationService {
	return &OperationService{
		db:                      db,
		operationRepo:           repositories.NewOperationRepository(db),
		accountService:          accountService,
		budgetService:           budgetService,
		plannedOperationService: plannedOperationService,
		goalService:             goalService,
		logger:                  logger,
	}
}

func (s *OperationService) Create(ctx context.Context, userID uuid.UUID, operation *models.Operation) (any, error) {
	account, err := s.accountService.GetByID(ctx, operation.AccountID, userID)
	if err != nil {
		return nil, err
	}

	switch operation.Type {
	case models.OperationTypeTransfer:
		if operation.FromAccountID == uuid.Nil {
			return nil, apperrors.BadRequest("from account ID is required")
		}

		fromAccount, err := s.accountService.GetByID(ctx, operation.FromAccountID, userID)
		if err != nil {
			return nil, err
		}

		if fromAccount.Currency != account.Currency {
			return nil, apperrors.BadRequest("currency mismatch")
		}

		if fromAccount.Balance.LessThan(operation.Amount) {
			return nil, apperrors.BadRequest("insufficient balance")
		}
	case models.OperationTypeExpense:
		if account.Balance.LessThan(operation.Amount) {
			return nil, apperrors.BadRequest("insufficient balance")
		}
	case models.OperationTypeReservation:
		if operation.GoalID == uuid.Nil && operation.PlannedOperationID == uuid.Nil {
			return nil, apperrors.BadRequest("goal or planned operation is required")
		} else if operation.GoalID != uuid.Nil && operation.PlannedOperationID != uuid.Nil {
			return nil, apperrors.BadRequest("goal and planned operation cannot be set at the same time")
		}

		if operation.GoalID != uuid.Nil {
			goal, err := s.goalService.GetByID(ctx, operation.GoalID, userID)
			if err != nil {
				return nil, err
			}

			if account.Currency != goal.Currency {
				return nil, apperrors.BadRequest("currency mismatch")
			}

			if account.Balance.LessThan(operation.Amount) {
				return nil, apperrors.BadRequest("insufficient balance")
			}

		}

		if operation.PlannedOperationID != uuid.Nil {
			plannedOperation, err := s.plannedOperationService.GetByID(ctx, operation.PlannedOperationID, userID)
			if err != nil {
				return nil, err
			}

			accountPlannedOperation, err := s.accountService.GetByID(ctx, plannedOperation.Account.ID, userID)
			if err != nil {
				return nil, err
			}

			if accountPlannedOperation.Currency != account.Currency {
				return nil, apperrors.BadRequest("currency mismatch")
			}

			if accountPlannedOperation.Balance.LessThan(operation.Amount) {
				return nil, apperrors.BadRequest("insufficient balance")
			}

			operation.AccountID = accountPlannedOperation.ID
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Errorw("failed to begin transaction", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	defer tx.Rollback()

	created, err := s.operationRepo.Create(ctx, operation, tx)
	if err != nil {
		s.logger.Errorw("failed to create operation", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if err := s.applyEffects(ctx, tx, userID, operation, account.Currency); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorw("failed to commit transaction", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return created.Response(), nil
}

func (s *OperationService) GetByID(ctx context.Context, id, userID uuid.UUID) (any, error) {
	operation, err := s.get(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return operation.Response(), nil
}

func (s *OperationService) List(ctx context.Context, userID uuid.UUID, params repositories.OperationListParams) (*dto.OperationResponseList, error) {
	operations, total, err := s.operationRepo.List(ctx, params, userID)
	if err != nil {
		s.logger.Errorw("failed to list operations", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	if operations == nil {
		operations = []dto.OperationResponseTransfer{}
	}

	response := make([]any, len(operations))
	for i, operation := range operations {
		response[i] = operation.Response()
	}
	return &dto.OperationResponseList{
		Operations: response,
		TotalCount: total,
	}, nil
}

func (s *OperationService) Update(ctx context.Context, userID uuid.UUID, operation *models.Operation) (any, error) {
	existing, err := s.get(ctx, operation.ID, userID)
	if err != nil {
		return nil, err
	}

	account, err := s.accountService.GetByID(ctx, existing.Account.ID, userID)
	if err != nil {
		return nil, err
	}

	var fromAccount *dto.AccountResponse
	if existing.FromAccount.ID != uuid.Nil {
		fromAccount, err = s.accountService.GetByID(ctx, existing.FromAccount.ID, userID)
		if err != nil {
			return nil, err
		}
	}

	delta := operation.Amount.Sub(existing.Amount)

	switch existing.Type {
	case models.OperationTypeTransfer:
		if delta.IsPositive() && fromAccount.Balance.LessThan(delta) {
			return nil, apperrors.BadRequest("insufficient balance")
		}
	case models.OperationTypeExpense:
		if delta.IsPositive() && account.Balance.LessThan(delta) {
			return nil, apperrors.BadRequest("insufficient balance")
		}
	case models.OperationTypeReservation:
		if delta.IsPositive() && account.Balance.LessThan(operation.Amount) {
			return nil, apperrors.BadRequest("insufficient balance")
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Errorw("failed to begin transaction", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	defer tx.Rollback()

	updated, err := s.operationRepo.Update(ctx, operation, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgOperationNotFound)
		}
		s.logger.Errorw("failed to update operation", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if !operation.Amount.IsZero() {
		effect := &models.Operation{
			Type:               existing.Type,
			AccountID:          existing.Account.ID,
			FromAccountID:      existing.FromAccount.ID,
			CategoryID:         existing.Category.ID,
			PlannedOperationID: existing.PlannedOperation.ID,
			GoalID:             existing.Goal.ID,
			Amount:             operation.Amount.Sub(existing.Amount),
		}
		if err := s.applyEffects(ctx, tx, userID, effect, account.Currency); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorw("failed to commit transaction", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return updated.Response(), nil
}

func (s *OperationService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	operation, err := s.get(ctx, id, userID)
	if err != nil {
		return err
	}

	account, err := s.accountService.GetByID(ctx, operation.Account.ID, userID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Errorw("failed to begin transaction", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	defer tx.Rollback()

	if err := s.operationRepo.Delete(ctx, id, tx); err != nil {
		s.logger.Errorw("failed to delete operation", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	effect := &models.Operation{
		Type:               operation.Type,
		AccountID:          operation.Account.ID,
		FromAccountID:      operation.FromAccount.ID,
		CategoryID:         operation.Category.ID,
		PlannedOperationID: operation.PlannedOperation.ID,
		GoalID:             operation.Goal.ID,
		Amount:             operation.Amount.Neg(),
	}
	if err := s.applyEffects(ctx, tx, userID, effect, account.Currency); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorw("failed to commit transaction", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	return nil
}

func (s *OperationService) get(ctx context.Context, id, userID uuid.UUID) (*dto.OperationResponseTransfer, error) {
	operation, err := s.operationRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgOperationNotFound)
		}
		s.logger.Errorw("failed to get operation", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return operation, nil
}

func (s *OperationService) applyEffects(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	op *models.Operation,
	currency string,
) error {
	budgetID, err := s.budgetService.GetByCategoryID(ctx, userID, op.CategoryID, currency, tx)
	if err != nil {
		return err
	}

	switch op.Type {
	case models.OperationTypeTransfer:
		if err := s.accountService.UpdateBalance(ctx, op.AccountID, op.Amount, tx); err != nil {
			return err
		}
		if err := s.accountService.UpdateBalance(ctx, op.FromAccountID, op.Amount.Neg(), tx); err != nil {
			return err
		}
	case models.OperationTypeReservation:
		if op.GoalID == uuid.Nil && op.PlannedOperationID == uuid.Nil {
			return apperrors.BadRequest("goal or planned operation is required")
		}
		if op.GoalID != uuid.Nil {
			if err := s.goalService.UpdateBalance(ctx, op.GoalID, op.Amount, tx); err != nil {
				return err
			}
		}
		if op.PlannedOperationID != uuid.Nil {
			if err := s.plannedOperationService.UpdateBalance(ctx, op.PlannedOperationID, op.Amount, tx); err != nil {
				return err
			}
		}
	default:
		if err := s.accountService.UpdateBalance(ctx, op.AccountID, signed(op.Type, op.Amount), tx); err != nil {
			return err
		}

		if op.Type == models.OperationTypeExpense && budgetID != uuid.Nil {
			if err := s.budgetService.UpdateBalance(ctx, budgetID, signed(op.Type, op.Amount), tx); err != nil {
				return err
			}
		}
	}

	return nil
}

func signed(opType models.OperationType, amount decimal.Decimal) decimal.Decimal {
	if opType == models.OperationTypeExpense {
		return amount.Neg()
	}
	return amount
}
