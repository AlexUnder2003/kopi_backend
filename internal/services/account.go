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
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

const (
	errMsgAccountNotFound      = "not_found_account"
	errMsgAccountAlreadyExists = "already_exists_account"
)

type AccountService struct {
	accountRepo *repositories.AccountRepository
	logger      *zap.SugaredLogger
}

func NewAccountService(db *sql.DB, logger *zap.SugaredLogger) *AccountService {
	return &AccountService{
		accountRepo: repositories.NewAccountRepository(db),
		logger:      logger,
	}
}

func (s *AccountService) Create(ctx context.Context, account *models.Account) (*dto.AccountResponse, error) {
	acc, err := s.accountRepo.Create(ctx, account)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.Conflict(errMsgAccountAlreadyExists)
		}

		s.logger.Errorw("failed to create account", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return acc, nil
}

func (s *AccountService) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.AccountResponse, error) {
	acc, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgAccountNotFound)
		}

		s.logger.Errorw("failed to get account", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if acc.UserID != userID {
		return nil, apperrors.NotFound(errMsgAccountNotFound)
	}

	return acc, nil
}

func (s *AccountService) List(ctx context.Context, userID uuid.UUID) ([]dto.AccountResponse, error) {
	accounts, err := s.accountRepo.List(ctx, userID)
	if err != nil {
		s.logger.Errorw("failed to list accounts", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return accounts, nil
}

func (s *AccountService) Update(ctx context.Context, userID uuid.UUID, account *models.Account) (*dto.AccountResponse, error) {
	if _, err := s.GetByID(ctx, account.ID, userID); err != nil {
		return nil, err
	}

	updatedAccount, err := s.accountRepo.Update(ctx, account)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.Conflict(errMsgAccountAlreadyExists)
		}
		s.logger.Errorw("failed to update account", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return updatedAccount, nil
}

func (s *AccountService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.GetByID(ctx, id, userID); err != nil {
		return err
	}

	if err := s.accountRepo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete account", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	return nil
}
