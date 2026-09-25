package services

import (
	"context"
	"database/sql"
	"errors"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrAccountNotFound = errors.New("account not found")

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
	return s.accountRepo.Create(ctx, account)
}

func (s *AccountService) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.AccountResponse, error) {
	acc, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if acc.UserID != userID {
		return nil, ErrAccountNotFound
	}

	return acc, nil
}

func (s *AccountService) List(ctx context.Context, userID uuid.UUID) ([]dto.AccountResponse, error) {
	return s.accountRepo.List(ctx, userID)
}

func (s *AccountService) Update(ctx context.Context, userID uuid.UUID, account *models.Account) (*dto.AccountResponse, error) {
	if _, err := s.GetByID(ctx, account.ID, userID); err != nil {
		return nil, err
	}

	return s.accountRepo.Update(ctx, account)
}

func (s *AccountService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.GetByID(ctx, id, userID); err != nil {
		return err
	}

	return s.accountRepo.Delete(ctx, id)
}
