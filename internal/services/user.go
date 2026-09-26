package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"KopiBackend/internal/apperrors"
	"KopiBackend/internal/config"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"
	emailtpl "KopiBackend/internal/templates/email"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	errMsgUserNotFound        = "not_found_user"
	errMsgInvalidOTP          = "invalid_otp"
	errMsgRefreshTokenExpired = "refresh_token_expired"
	errMsgUnauthorized        = "unauthorized"
)

type UserService struct {
	userRepo      *repositories.UserRepository
	userTokenRepo *repositories.UserTokenRepository
	appConfig     *config.AppConfig
	logger        *zap.SugaredLogger
	smtpService   *SMTPService
}

func NewUserService(db *sql.DB, appConfig *config.AppConfig, logger *zap.SugaredLogger) *UserService {
	userRepo := repositories.NewUserRepository(db)
	userTokenRepo := repositories.NewUserTokenRepository(db)
	smtpService := NewSMTPService(appConfig, logger)
	return &UserService{userRepo: userRepo, userTokenRepo: userTokenRepo, appConfig: appConfig, logger: logger, smtpService: smtpService}
}

func (s *UserService) generateJWTTokens(userId uuid.UUID) (string, string, error) {
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString([]byte(s.appConfig.SecretKey))
	if err != nil {
		s.logger.Errorw("failed to create access token", "error", err)
		return "", "", apperrors.Internal(errInternalServerError)
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(90 * 12 * time.Hour).Unix(),
	}).SignedString([]byte(s.appConfig.SecretKey))
	if err != nil {
		s.logger.Errorw("failed to create refresh token", "error", err)
		return "", "", apperrors.Internal(errInternalServerError)
	}

	return accessToken, refreshToken, nil
}

func (s *UserService) generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgUserNotFound)
		}
		s.logger.Errorw("failed to get user", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return user, nil
}

func (s *UserService) SendOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			s.logger.Errorw("failed to get user by email", "error", err)
			return apperrors.Internal(errInternalServerError)
		}

		user, err = s.userRepo.Create(ctx, &models.User{Email: email})
		if err != nil {
			s.logger.Errorw("failed to create user", "error", err)
			return apperrors.Internal(errInternalServerError)
		}
	}

	otp, err := s.generateOTP()
	if err != nil {
		s.logger.Errorw("failed to generate OTP", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	if err = s.userRepo.CreateOTP(ctx, user.ID, otp); err != nil {
		s.logger.Errorw("failed to create OTP", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	html, err := emailtpl.RenderOTP(otp)
	if err != nil {
		s.logger.Errorw("failed to render OTP email", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	if err = s.smtpService.Send(email, "Код входа в Kopi", html); err != nil {
		s.logger.Errorw("failed to send OTP", "error", err)
		return apperrors.Internal(errInternalServerError)
	}

	return nil
}

func (s *UserService) Login(ctx context.Context, otp string) (*dto.LoginResponse, error) {
	userOTP, err := s.userRepo.GetUserIdByOTP(ctx, otp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.Unauthorized(errMsgInvalidOTP)
		}
		s.logger.Errorw("failed to get user by OTP", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if time.Now().After(userOTP.ExpiresAt) {
		return nil, apperrors.Unauthorized(errMsgInvalidOTP)
	}

	accessToken, refreshToken, err := s.generateJWTTokens(userOTP.UserID)
	if err != nil {
		return nil, err
	}

	refreshTokenHash := sha256.Sum256([]byte(refreshToken))
	if err = s.userTokenRepo.Create(ctx, &models.UserToken{
		UserID:    userOTP.UserID,
		TokenHash: hex.EncodeToString(refreshTokenHash[:]),
		ExpiresAt: time.Now().Add(90 * 12 * time.Hour),
	}); err != nil {
		s.logger.Errorw("failed to create user token", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if err = s.userRepo.DeleteOTP(ctx, otp); err != nil {
		s.logger.Errorw("failed to delete OTP", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Refresh(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
	refreshTokenHash := sha256.Sum256([]byte(refreshToken))

	userToken, err := s.userTokenRepo.GetByTokenHash(ctx, hex.EncodeToString(refreshTokenHash[:]))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.Unauthorized(errMsgUnauthorized)
		}
		s.logger.Errorw("failed to get user token", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if userToken.ExpiresAt.Before(time.Now()) {
		return nil, apperrors.Unauthorized(errMsgRefreshTokenExpired)
	}

	accessToken, _, err := s.generateJWTTokens(userToken.UserID)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *UserService) Create(ctx context.Context, user *models.User) (*dto.UserResponse, error) {
	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Errorw("failed to create user", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return created, nil
}

func (s *UserService) Update(ctx context.Context, user *models.User) (*dto.UserResponse, error) {
	updated, err := s.userRepo.Update(ctx, user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgUserNotFound)
		}
		s.logger.Errorw("failed to update user", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return updated, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete user", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	return nil
}
