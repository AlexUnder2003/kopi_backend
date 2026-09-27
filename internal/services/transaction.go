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

const errMsgTransactionNotFound = "not_found_transaction"
const errMsgSameTransfer = "bad_request_not_same_transfer"

type TransactionService struct {
	db              *sql.DB
	transactionRepo *repositories.TransactionRepository
	accountRepo     *repositories.AccountRepository
	categoryRepo    *repositories.CategoryRepository
	logger          *zap.SugaredLogger
}

func NewTransactionService(db *sql.DB, logger *zap.SugaredLogger) *TransactionService {
	return &TransactionService{
		db:              db,
		transactionRepo: repositories.NewTransactionRepository(db),
		accountRepo:     repositories.NewAccountRepository(db),
		categoryRepo:    repositories.NewCategoryRepository(db),
		logger:          logger,
	}
}

func (s *TransactionService) Create(ctx context.Context, userID uuid.UUID, transaction *models.Transaction) (*dto.TransactionResponse, error) {
	acc, err := s.accountRepo.GetByID(ctx, transaction.AccountID)
	if err != nil {
		return nil, err
	}

	if acc.UserID != userID {
		return nil, apperrors.NotFound(errMsgAccountNotFound)
	}

	created, err := s.transactionRepo.Create(ctx, transaction, nil)
	if err != nil {
		s.logger.Errorw("failed to create transaction", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return created, nil
}

func (s *TransactionService) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.TransactionResponse, error) {
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

func (s *TransactionService) CreateTransfer(ctx context.Context, userID uuid.UUID, from *models.Transaction, to *models.Transaction) ([]dto.TransactionResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	transferCategoryID, err := s.categoryRepo.GetTransferCategoryID(ctx)
	if err != nil {
		return nil, err
	}

	transactionUUID := uuid.New()

	from.CategoryID = transferCategoryID
	to.CategoryID = transferCategoryID
	from.TransferID = &transactionUUID
	to.TransferID = &transactionUUID

	fromAcc, err := s.accountRepo.GetByID(ctx, from.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgAccountNotFound)
		}
		s.logger.Errorw("failed to get from account", "error", err)
		return nil, err
	}

	if fromAcc.UserID != userID {
		return nil, apperrors.NotFound(errMsgAccountNotFound)
	}

	toAcc, err := s.accountRepo.GetByID(ctx, to.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgAccountNotFound)
		}
		s.logger.Errorw("failed to get to account", "error", err)
		return nil, err
	}

	if toAcc.UserID != userID {
		return nil, apperrors.NotFound(errMsgAccountNotFound)
	}

	fromTx, err := s.transactionRepo.Create(ctx, from, tx)
	if err != nil {
		return nil, err
	}

	toTx, err := s.transactionRepo.Create(ctx, to, tx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	response := make([]dto.TransactionResponse, 2)
	response[0] = *fromTx
	response[1] = *toTx

	return response, nil
}

func (s *TransactionService) List(ctx context.Context, userID uuid.UUID, params repositories.TransactionListParams) ([]dto.TransactionResponse, error) {
	transactions, err := s.transactionRepo.List(ctx, params, userID)
	if err != nil {
		s.logger.Errorw("failed to list transactions", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return transactions, nil
}

func (s *TransactionService) Update(ctx context.Context, userID uuid.UUID, transaction *models.Transaction) (*dto.TransactionResponse, error) {
	_, err := s.GetByID(ctx, transaction.ID, userID)
	if err != nil {
		return nil, err
	}

	if transaction.AccountID != uuid.Nil {
		acc, err := s.accountRepo.GetByID(ctx, transaction.AccountID)
		if err != nil {
			return nil, err
		}
		if acc.UserID != userID {
			return nil, apperrors.NotFound(errMsgAccountNotFound)
		}
	}

	updated, err := s.transactionRepo.Update(ctx, transaction, nil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgTransactionNotFound)
		}
		s.logger.Errorw("failed to update transaction", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return updated, nil
}

func (s *TransactionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.GetByID(ctx, id, userID); err != nil {
		return err
	}

	if err := s.transactionRepo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete transaction", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	return nil
}
