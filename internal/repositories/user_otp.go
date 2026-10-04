package repositories

import (
	"context"
	"database/sql"

	"KopiBackend/internal/models"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
)

type UserOTPRepository struct {
	db *sql.DB
}

func NewUserOTPRepository(db *sql.DB) *UserOTPRepository {
	return &UserOTPRepository{db: db}
}

func (r *UserOTPRepository) Create(ctx context.Context, userID uuid.UUID, otp string) error {
	const q = `
		INSERT INTO user_otps (user_id, otp)
		VALUES ($1, $2)`

	_, err := r.db.ExecContext(ctx, q, userID, otp)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserOTPRepository) GetByOTP(ctx context.Context, otp string) (*models.UserOTP, error) {
	const q = `SELECT user_id, expires_at FROM user_otps WHERE otp = $1`

	var row models.UserOTP
	if err := sqlscan.Get(ctx, r.db, &row, q, otp); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserOTPRepository) Delete(ctx context.Context, otp string) error {
	const q = `DELETE FROM user_otps WHERE otp = $1`

	_, err := r.db.ExecContext(ctx, q, otp)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserOTPRepository) DeleteExpired(ctx context.Context) error {
	const q = `DELETE FROM user_otps WHERE expires_at < now()`

	_, err := r.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	return nil
}
