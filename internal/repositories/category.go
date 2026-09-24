package repositories

import (
	"context"
	"database/sql"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) (*dto.CategoryResponse, error) {
	const q = `
		INSERT INTO categories (name, icon, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, icon`

	var resp dto.CategoryResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, category.Name, category.Icon, category.UserID); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.CategoryResponse, error) {
	const q = `SELECT id, name, icon FROM categories WHERE id = $1`

	var resp dto.CategoryResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *CategoryRepository) List(ctx context.Context, userID uuid.UUID) ([]dto.CategoryResponse, error) {
	const q = `
		SELECT id, name, icon
		FROM categories
		WHERE user_id IS NULL OR user_id = $1
		ORDER BY name`

	var resp []dto.CategoryResponse
	if err := sqlscan.Select(ctx, r.db, &resp, q, userID); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) (*dto.CategoryResponse, error) {
	const q = `
		UPDATE categories
		SET name = COALESCE($2, name),
		    icon = COALESCE($3, icon)
		WHERE id = $1
		RETURNING id, name, icon`

	var resp dto.CategoryResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		category.ID, nullString(category.Name), nullString(category.Icon),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM categories WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
