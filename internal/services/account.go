package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"KopiBackend/internal/apperrors"
	"KopiBackend/internal/config"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	errMsgAccountNotFound      = "not_found_account"
	errMsgAccountAlreadyExists = "already_exists_account"
	errMsgCurrencyRateNotFound = "not_found_currency_rate"
)

type AccountService struct {
	accountRepo         *repositories.AccountRepository
	subscriptionService *SubscriptionService
	config              *config.AppConfig
	logger              *zap.SugaredLogger
}

func NewAccountService(db *sql.DB, config *config.AppConfig, logger *zap.SugaredLogger) *AccountService {
	return &AccountService{
		accountRepo:         repositories.NewAccountRepository(db),
		subscriptionService: NewSubscriptionService(db, logger),
		config:              config,
		logger:              logger,
	}
}

func (s *AccountService) Create(ctx context.Context, account *models.Account) (*dto.AccountResponse, error) {
	hasAccess, err := s.subscriptionService.HasAccess(ctx, account.UserID, models.SubscriptionPlanPro)
	if err != nil {
		s.logger.Errorw("failed to check subscription plan", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	accountsCount, err := s.accountRepo.CountByUser(ctx, account.UserID)
	if err != nil {
		s.logger.Errorw("failed to count accounts", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if accountsCount >= 3 && !hasAccess {
		return nil, apperrors.Forbidden(errMsgSubscriptionPlanProRequired)
	}

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

func (s *AccountService) List(ctx context.Context, userID uuid.UUID, params repositories.ListParams, currency string) (*dto.AccountResponseList, error) {
	accounts, err := s.accountRepo.List(ctx, userID, params)
	if err != nil {
		s.logger.Errorw("failed to list accounts", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	var totalBalance decimal.Decimal

	if currency != "" {
		rates, err := GetCurrencyRates(s.config, strings.ToLower(currency))
		if err != nil {
			s.logger.Errorw("failed to get currency rates", "error", err)
			return nil, apperrors.Internal(errInternalServerError)
		}

		for _, account := range accounts {
			if account.Currency == currency {
				totalBalance = totalBalance.Add(account.Balance)
			} else {
				rate, ok := rates[strings.ToLower(account.Currency)]
				if !ok || rate.IsZero() {
					return nil, apperrors.NotFound(errMsgCurrencyRateNotFound)
				}

				totalBalance = totalBalance.Add(account.Balance.Div(rate))
			}
		}
	}

	return &dto.AccountResponseList{
		Accounts:     accounts,
		TotalBalance: totalBalance,
	}, nil

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

func (s *AccountService) UpdateBalance(ctx context.Context, accountID uuid.UUID, delta decimal.Decimal, tx *sql.Tx) error {
	if err := s.accountRepo.UpdateBalance(ctx, accountID, delta, tx); err != nil {
		s.logger.Errorw("failed to update account balance", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	return nil
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
