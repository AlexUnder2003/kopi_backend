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
	db            *sql.DB
	operationRepo *repositories.OperationRepository
	accountRepo   *repositories.AccountRepository
	budgetRepo    *repositories.BudgetRepository
	logger        *zap.SugaredLogger
}

func NewOperationService(db *sql.DB, logger *zap.SugaredLogger) *OperationService {
	return &OperationService{
		db:            db,
		operationRepo: repositories.NewOperationRepository(db),
		accountRepo:   repositories.NewAccountRepository(db),
		budgetRepo:    repositories.NewBudgetRepository(db),
		logger:        logger,
	}
}

func (s *OperationService) Create(ctx context.Context, userID uuid.UUID, operation *models.Operation) (any, error) {
	account, err := s.accountRepo.GetByID(ctx, operation.AccountID)
	if err != nil {
		return nil, err
	}

	if account.UserID != userID {
		return nil, apperrors.NotFound(errMsgAccountNotFound)
	}

	switch operation.Type {
	case models.OperationTypeTransfer:
		if operation.FromAccountID == uuid.Nil {
			return nil, apperrors.BadRequest("from account ID is required")
		}

		fromAccount, err := s.accountRepo.GetByID(ctx, operation.FromAccountID)
		if err != nil {
			return nil, err
		}

		if fromAccount.UserID != userID {
			return nil, apperrors.NotFound(errMsgAccountNotFound)
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

	if err := s.applyEffects(
		ctx,
		tx,
		operation.Type,
		operation.AccountID,
		operation.FromAccountID,
		operation.CategoryID,
		userID,
		operation.Amount,
		account.Currency,
	); err != nil {
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

func (s *OperationService) List(ctx context.Context, userID uuid.UUID, params repositories.OperationListParams) ([]any, error) {
	operations, err := s.operationRepo.List(ctx, params, userID)
	if err != nil {
		s.logger.Errorw("failed to list operations", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	response := make([]any, len(operations))
	for i, operation := range operations {
		response[i] = operation.Response()
	}
	return response, nil
}

func (s *OperationService) Update(ctx context.Context, userID uuid.UUID, operation *models.Operation) (any, error) {
	existing, err := s.get(ctx, operation.ID, userID)
	if err != nil {
		return nil, err
	}

	account, err := s.accountRepo.GetByID(ctx, existing.Account.ID)
	if err != nil {
		return nil, err
	}

	if operation.AccountID != uuid.Nil {
		account, err = s.accountRepo.GetByID(ctx, operation.AccountID)
		if err != nil {
			return nil, err
		}
	}

	delta := operation.Amount.Sub(existing.Amount)

	switch existing.Type {
	case models.OperationTypeTransfer:
		if existing.FromAccount.ID == uuid.Nil {
			return nil, apperrors.BadRequest("from account ID is required")
		}

		fromAccount, err := s.accountRepo.GetByID(ctx, existing.FromAccount.ID)
		if err != nil {
			return nil, err
		}

		if fromAccount.Currency != account.Currency {
			return nil, apperrors.BadRequest("currency mismatch")
		}

		if delta.IsPositive() && fromAccount.Balance.LessThan(delta) {
			return nil, apperrors.BadRequest("insufficient balance")
		}
	case models.OperationTypeExpense:
		if delta.IsPositive() && account.Balance.LessThan(delta) {
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
		if err := s.applyEffects(
			ctx,
			tx,
			existing.Type,
			existing.Account.ID,
			existing.FromAccount.ID,
			existing.Category.ID,
			userID,
			operation.Amount.Sub(existing.Amount),
			account.Currency,
		); err != nil {
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

	account, err := s.accountRepo.GetByID(ctx, operation.Account.ID)
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

	if err := s.applyEffects(
		ctx,
		tx,
		operation.Type,
		operation.Account.ID,
		operation.FromAccount.ID,
		operation.Category.ID,
		userID,
		operation.Amount.Neg(),
		account.Currency,
	); err != nil {
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
	opType models.OperationType,
	accountID, fromAccountID, categoryID, userID uuid.UUID,
	amount decimal.Decimal,
	currency string,
) error {
	budgetID, err := s.budgetRepo.GetByCategoryID(ctx, userID, categoryID, currency, tx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Errorw("failed to get budget by category ID", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	switch opType {
	case models.OperationTypeTransfer:
		if err := s.accountRepo.UpdateBalance(ctx, accountID, amount, tx); err != nil {
			s.logger.Errorw("failed to update account balance", "error", err)
			return apperrors.Internal(errInternalServerError)
		}
		if err := s.accountRepo.UpdateBalance(ctx, fromAccountID, amount.Neg(), tx); err != nil {
			s.logger.Errorw("failed to update account balance", "error", err)
			return apperrors.Internal(errInternalServerError)
		}
	default:
		if err := s.accountRepo.UpdateBalance(ctx, accountID, signed(opType, amount), tx); err != nil {
			s.logger.Errorw("failed to update account balance", "error", err)
			return apperrors.Internal(errInternalServerError)
		}

		if opType == models.OperationTypeExpense && budgetID != uuid.Nil {
			if err := s.budgetRepo.UpdateBalance(ctx, budgetID, signed(opType, amount), tx); err != nil {
				s.logger.Errorw("failed to update budget balance", "error", err)
				return apperrors.Internal(errInternalServerError)
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
