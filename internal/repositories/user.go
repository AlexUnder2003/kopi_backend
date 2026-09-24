package repositories

import (
	"context"
	"database/sql"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (*dto.UserResponse, error) {
	const q = `
		INSERT INTO users (name, email)
		VALUES ($1, $2)
		RETURNING id, name, email`

	var resp dto.UserResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, user.Name, user.Email); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	const q = `SELECT id, name, email FROM users WHERE id = $1`

	var resp dto.UserResponse
	if err := sqlscan.Get(ctx, r.db, &resp, q, id); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	if err := sqlscan.Get(ctx, r.db, &exists, q, email); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) (*dto.UserResponse, error) {
	const q = `
		UPDATE users
		SET name = COALESCE($2, name),
		    email = COALESCE($3, email)
		WHERE id = $1
		RETURNING id, name, email`

	var resp dto.UserResponse
	if err := sqlscan.Get(
		ctx, r.db, &resp, q,
		user.ID, nullString(user.Name), nullString(user.Email),
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM users WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
