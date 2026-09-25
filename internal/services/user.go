package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"KopiBackend/internal/config"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"
	"crypto/sha256"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserService struct {
	userRepo      *repositories.UserRepository
	userTokenRepo *repositories.UserTokenRepository
	appConfig     *config.AppConfig
	logger        *zap.SugaredLogger
}

func NewUserService(db *sql.DB, appConfig *config.AppConfig, logger *zap.SugaredLogger) *UserService {
	userRepo := repositories.NewUserRepository(db)
	userTokenRepo := repositories.NewUserTokenRepository(db)
	return &UserService{userRepo: userRepo, userTokenRepo: userTokenRepo, appConfig: appConfig, logger: logger}
}

func (s *UserService) generateJWTTokens(userId uuid.UUID) (string, string, error) {
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString([]byte(s.appConfig.SecretKey))

	if err != nil {
		s.logger.Errorf("Failed to create access token: %v", err)
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(90 * 12 * time.Hour).Unix(),
	}).SignedString([]byte(s.appConfig.SecretKey))

	if err != nil {
		s.logger.Errorf("Failed to create refresh token: %v", err)
		return "", "", err
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
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) SendOTP(ctx context.Context, email string) error {
	var user *dto.UserResponse

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			user, err = s.userRepo.Create(ctx, &models.User{Email: email})
			if err != nil {
				s.logger.Errorf("Failed to create user: %v", err)
				return err
			}
		}
		return err
	}

	otp, err := s.generateOTP()
	if err != nil {
		s.logger.Errorf("Failed to generate OTP: %v", err)
		return err
	}

	err = s.userRepo.CreateOTP(ctx, user.ID, otp)
	if err != nil {
		s.logger.Errorf("Failed to create OTP: %v", err)
		return err
	}

	return nil
}

func (s *UserService) Login(ctx context.Context, email, otp string) (*dto.LoginResponse, error) {
	userId, err := s.userRepo.GetUserIdByOTP(ctx, otp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid OTP")
		}
		s.logger.Errorf("Failed to get user ID by OTP: %v", err)
		return nil, err
	}

	accessToken, refreshToken, err := s.generateJWTTokens(userId)
	if err != nil {
		s.logger.Errorf("Failed to generate JWT tokens: %v", err)
		return nil, err
	}

	refreshTokenHash := sha256.Sum256([]byte(refreshToken))

	err = s.userTokenRepo.Create(ctx, &models.UserToken{
		UserID:    userId,
		TokenHash: hex.EncodeToString(refreshTokenHash[:]),
		ExpiresAt: time.Now().Add(90 * 12 * time.Hour),
	})
	if err != nil {
		s.logger.Errorf("Failed to create user token: %v", err)
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Create(ctx context.Context, user *models.User) (*dto.UserResponse, error) {
	return s.userRepo.Create(ctx, user)
}

func (s *UserService) Update(ctx context.Context, user *models.User) (*dto.UserResponse, error) {
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}
