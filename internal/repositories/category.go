package repositories

import (
	"context"
	"database/sql"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
)

const transferCategoryName = "Переводы"

const categorySelectColumns = `
	id,
	name,
	icon,
	CASE WHEN user_id IS NULL THEN TRUE ELSE FALSE END AS is_system`

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
		RETURNING ` + categorySelectColumns

	var resp dto.CategoryResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, category.Name, category.Icon, nullUUID(category.UserID)); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.CategoryResponse, error) {
	const q = `SELECT ` + categorySelectColumns + ` FROM categories WHERE id = $1`

	var resp dto.CategoryResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *CategoryRepository) List(ctx context.Context, userID uuid.UUID, params ListParams) ([]dto.CategoryResponse, error) {
	const q = `
		SELECT ` + categorySelectColumns + `
 		FROM categories
		WHERE (user_id IS NULL OR user_id = $1)
			AND (name ILIKE '%' || $2 || '%' OR $2 IS NULL)
		ORDER BY name
		LIMIT $3 OFFSET $4`

	var resp []dto.CategoryResponse
	if err := sqlscan.Select(ctx, r.db, &resp, q, userID, nullString(params.Filter), nullInt(params.Limit), nullInt(params.Offset)); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *CategoryRepository) GetTransferCategoryID(ctx context.Context) (uuid.UUID, error) {
	const q = `SELECT id FROM categories WHERE name = $1`

	var resp uuid.UUID
	if err := sqlscan.Get(ctx, r.db, &resp, q, transferCategoryName); err != nil {
		return uuid.Nil, err
	}
	return resp, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) (*dto.CategoryResponse, error) {
	const q = `
		UPDATE categories
		SET name = COALESCE($2, name),
		    icon = COALESCE($3, icon)
		WHERE id = $1
		RETURNING ` + categorySelectColumns

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
