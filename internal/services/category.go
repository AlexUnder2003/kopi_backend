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
	errMsgCategoryNotFound      = "not_found_category"
	errMsgCategoryAlreadyExists = "already_exists_category"
)

type CategoryService struct {
	categoryRepo *repositories.CategoryRepository
	logger       *zap.SugaredLogger
}

func NewCategoryService(db *sql.DB, logger *zap.SugaredLogger) *CategoryService {
	return &CategoryService{
		categoryRepo: repositories.NewCategoryRepository(db),
		logger:       logger,
	}
}

func (s *CategoryService) Create(ctx context.Context, category *models.Category) (*dto.CategoryResponse, error) {
	cat, err := s.categoryRepo.Create(ctx, category)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.Conflict(errMsgCategoryAlreadyExists)
		}
		s.logger.Errorw("failed to create category", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return cat, nil
}

func (s *CategoryService) GetByID(ctx context.Context, id uuid.UUID) (*dto.CategoryResponse, error) {
	cat, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(errMsgCategoryNotFound)
		}
		s.logger.Errorw("failed to get category", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return cat, nil
}

func (s *CategoryService) List(ctx context.Context, userID uuid.UUID, params repositories.ListParams) ([]dto.CategoryResponse, error) {
	categories, err := s.categoryRepo.List(ctx, userID, params)
	if err != nil {
		s.logger.Errorw("failed to list categories", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return categories, nil
}

func (s *CategoryService) Update(ctx context.Context, category *models.Category) (*dto.CategoryResponse, error) {
	if _, err := s.GetByID(ctx, category.ID); err != nil {
		return nil, err
	}

	updated, err := s.categoryRepo.Update(ctx, category)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.Conflict(errMsgCategoryAlreadyExists)
		}
		s.logger.Errorw("failed to update category", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}
	return updated, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete category", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	return nil
}
