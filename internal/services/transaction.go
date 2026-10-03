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

const errMsgTransactionNotFound = "not_found_transaction"

type TransactionService struct {
	db              *sql.DB
	transactionRepo *repositories.TransactionRepository
	accountRepo     *repositories.AccountRepository
	budgetRepo      *repositories.BudgetRepository
	logger          *zap.SugaredLogger
}

func NewTransactionService(db *sql.DB, logger *zap.SugaredLogger) *TransactionService {
	return &TransactionService{
		db:              db,
		transactionRepo: repositories.NewTransactionRepository(db),
		accountRepo:     repositories.NewAccountRepository(db),
		budgetRepo:      repositories.NewBudgetRepository(db),
		logger:          logger,
	}
}

func (s *TransactionService) Create(ctx context.Context, userID uuid.UUID, transaction *models.Transaction) (any, error) {
	account, err := s.accountRepo.GetByID(ctx, transaction.AccountID)
	if err != nil {
		return nil, err
	}

	switch transaction.Type {
	case models.TransactionTypeTransfer:
		fromAccount, err := s.accountRepo.GetByID(ctx, *transaction.FromAccountID)
		if err != nil {
			return nil, err
		}

		if transaction.FromAccountID == nil {
			return nil, apperrors.BadRequest("from account ID is required")
		}

		if fromAccount.Currency != account.Currency {
			return nil, apperrors.BadRequest("currency mismatch")
		}

		if fromAccount.Balance.LessThan(transaction.Amount) {
			return nil, apperrors.BadRequest("insufficient balance")
		}
	case models.TransactionTypeExpense:
		if account.Balance.LessThan(transaction.Amount) {
			return nil, apperrors.BadRequest("insufficient balance")
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Errorw("failed to begin transaction", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	defer tx.Rollback()

	created, err := s.transactionRepo.Create(ctx, transaction, tx)
	if err != nil {
		s.logger.Errorw("failed to create transaction", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if err := s.applyEffects(
		ctx,
		tx,
		transaction.Type,
		transaction.AccountID,
		uuidValue(transaction.FromAccountID),
		transaction.CategoryID,
		userID,
		transaction.Amount,
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

func (s *TransactionService) GetByID(ctx context.Context, id, userID uuid.UUID) (any, error) {
	transaction, err := s.get(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return transaction.Response(), nil
}

func (s *TransactionService) List(ctx context.Context, userID uuid.UUID, params repositories.TransactionListParams) ([]any, error) {
	transactions, err := s.transactionRepo.List(ctx, params, userID)
	if err != nil {
		s.logger.Errorw("failed to list transactions", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	response := make([]any, len(transactions))
	for i, transaction := range transactions {
		response[i] = transaction.Response()
	}
	return response, nil
}

func (s *TransactionService) Update(ctx context.Context, userID uuid.UUID, transaction *models.Transaction) (any, error) {
	existing, err := s.get(ctx, transaction.ID, userID)
	if err != nil {
		return nil, err
	}

	account, err := s.accountRepo.GetByID(ctx, transaction.AccountID)
	if err != nil {
		return nil, err
	}

	delta := transaction.Amount.Sub(existing.Amount)

	switch existing.Type {
	case models.TransactionTypeTransfer:
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
	case models.TransactionTypeExpense:
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

	updated, err := s.transactionRepo.Update(ctx, transaction, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgTransactionNotFound)
		}
		s.logger.Errorw("failed to update transaction", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if !transaction.Amount.IsZero() {
		if err := s.applyEffects(
			ctx,
			tx,
			existing.Type,
			existing.Account.ID,
			existing.FromAccount.ID,
			existing.Category.ID,
			userID,
			transaction.Amount.Sub(existing.Amount),
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

func (s *TransactionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	transaction, err := s.get(ctx, id, userID)
	if err != nil {
		return err
	}

	account, err := s.accountRepo.GetByID(ctx, transaction.Account.ID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Errorw("failed to begin transaction", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	defer tx.Rollback()

	if err := s.transactionRepo.Delete(ctx, id, tx); err != nil {
		s.logger.Errorw("failed to delete transaction", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	if err := s.applyEffects(
		ctx,
		tx,
		transaction.Type,
		transaction.Account.ID,
		transaction.FromAccount.ID,
		transaction.Category.ID,
		userID,
		transaction.Amount.Neg(),
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

func (s *TransactionService) get(ctx context.Context, id, userID uuid.UUID) (*dto.TransactionResponseTransfer, error) {
	transaction, err := s.transactionRepo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgTransactionNotFound)
		}
		s.logger.Errorw("failed to get transaction", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return transaction, nil
}

func (s *TransactionService) applyEffects(
	ctx context.Context,
	tx *sql.Tx,
	txType models.TransactionType,
	accountID, fromAccountID, categoryID, userID uuid.UUID,
	amount decimal.Decimal,
	currency string,
) error {
	budgetID, err := s.budgetRepo.GetByCategoryID(ctx, userID, categoryID, currency, tx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Errorw("failed to get budget by category ID", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	switch txType {
	case models.TransactionTypeTransfer:
		if err := s.accountRepo.UpdateBalance(ctx, accountID, amount, tx); err != nil {
			s.logger.Errorw("failed to update account balance", "error", err)
			return apperrors.Internal(errInternalServerError)
		}
		if err := s.accountRepo.UpdateBalance(ctx, fromAccountID, amount.Neg(), tx); err != nil {
			s.logger.Errorw("failed to update account balance", "error", err)
			return apperrors.Internal(errInternalServerError)
		}
	default:
		if err := s.accountRepo.UpdateBalance(ctx, accountID, signed(txType, amount), tx); err != nil {
			s.logger.Errorw("failed to update account balance", "error", err)
			return apperrors.Internal(errInternalServerError)
		}

		if txType == models.TransactionTypeExpense && budgetID != uuid.Nil {
			if err := s.budgetRepo.UpdateBalance(ctx, budgetID, signed(txType, amount), tx); err != nil {
				s.logger.Errorw("failed to update budget balance", "error", err)
				return apperrors.Internal(errInternalServerError)
			}
		}
	}

	return nil
}

func uuidValue(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return *id
}

func signed(txType models.TransactionType, amount decimal.Decimal) decimal.Decimal {
	if txType == models.TransactionTypeExpense {
		return amount.Neg()
	}
	return amount
}
