package repositories

import (
	"context"
	"database/sql"

	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
)

const userTokenColumns = `id, user_id, token_hash, expires_at`

type UserTokenRepository struct {
	db *sql.DB
}

func NewUserTokenRepository(db *sql.DB) *UserTokenRepository {
	return &UserTokenRepository{db: db}
}

func (r *UserTokenRepository) Create(ctx context.Context, token *models.UserToken) error {
	const q = `
		INSERT INTO user_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, q, token.UserID, token.TokenHash, token.ExpiresAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.UserToken, error) {
	const q = `SELECT ` + userTokenColumns + ` FROM user_tokens WHERE token_hash = $1`

	var token models.UserToken
	if err := sqlscan.Get(ctx, r.db, &token, q, tokenHash); err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *UserTokenRepository) DeleteExpired(ctx context.Context) error {
	const q = `DELETE FROM user_tokens WHERE expires_at < now()`

	_, err := r.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM user_tokens WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	return nil
}
