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

func (s *CategoryService) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.CategoryResponse, error) {
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

func (s *CategoryService) GetTransferCategoryID(ctx context.Context) (uuid.UUID, error) {
	id, err := s.categoryRepo.GetTransferCategoryID(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, apperrors.NotFound(errMsgCategoryNotFound)
		}
		s.logger.Errorw("failed to get transfer category", "error", err)
		return uuid.Nil, apperrors.Internal(errInternalServerError)
	}
	return id, nil
}

func (s *CategoryService) GetReservationCategoryID(ctx context.Context) (uuid.UUID, error) {
	id, err := s.categoryRepo.GetReservationCategoryID(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, apperrors.NotFound(errMsgCategoryNotFound)
		}
		s.logger.Errorw("failed to get reservation category", "error", err)
		return uuid.Nil, apperrors.Internal(errInternalServerError)
	}
	return id, nil
}

func (s *CategoryService) List(ctx context.Context, userID uuid.UUID, params repositories.ListParams) (*dto.CategoryResponseList, error) {
	categories, total, err := s.categoryRepo.List(ctx, userID, params)
	if err != nil {
		s.logger.Errorw("failed to list categories", "error", err)
		return nil, apperrors.Internal(errInternalServerError)
	}

	if categories == nil {
		categories = []dto.CategoryResponse{}
	}

	return &dto.CategoryResponseList{
		Categories: categories,
		TotalCount: total,
	}, nil
}

func (s *CategoryService) Update(ctx context.Context, userID uuid.UUID, category *models.Category) (*dto.CategoryResponse, error) {
	cat, err := s.GetByID(ctx, category.ID, userID)
	if err != nil {
		return nil, err
	}

	if cat.IsSystem {
		return nil, apperrors.BadRequest("cannot update default category")
	}

	if cat.UserID != userID {
		return nil, apperrors.BadRequest("cannot update category not owned by user")
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

func (s *CategoryService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	cat, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	if cat.IsSystem {
		return apperrors.BadRequest("cannot delete default category")
	}

	if cat.UserID != userID {
		return apperrors.BadRequest("cannot delete category not owned by user")
	}

	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete category", "error", err)
		return apperrors.Internal(errInternalServerError)
	}
	return nil
}
