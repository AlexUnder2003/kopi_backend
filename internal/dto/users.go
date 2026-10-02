package dto

import (
	"errors"

	"github.com/google/uuid"
)

type UserUpdate struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type UserResponse struct {
	ID    uuid.UUID `json:"id" db:"id"`
	Name  string    `json:"name" db:"name"`
	Email string    `json:"email" db:"email"`
}

type SendOTPRequest struct {
	Email string `json:"email"`
}

func (r *SendOTPRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	return nil
}

type LoginRequest struct {
	OTP string `json:"otp"`
}

func (r *LoginRequest) Validate() error {
	if r.OTP == "" {
		return errors.New("otp is required")
	}
	return nil
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r *RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}
	return nil
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}
