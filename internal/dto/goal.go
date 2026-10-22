package dto

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GoalPost struct {
	Name          string          `json:"name"`
	Currency      string          `json:"currency"`
	Deadline      time.Time       `json:"deadline,omitempty"`
	TargetAmount  decimal.Decimal `json:"target_amount"`
	CurrentAmount decimal.Decimal `json:"current_amount,omitempty"`
}

func (r *GoalPost) Validate() error {
	if r.Name == "" {
		return errors.New("name")
	}
	if !ValidCurrency(r.Currency) {
		return errors.New("currency")
	}
	if !r.TargetAmount.IsPositive() {
		return errors.New("target_amount")
	}
	if r.CurrentAmount.IsNegative() {
		return errors.New("current_amount")
	}
	return nil
}

type GoalUpdate struct {
	Name         string          `json:"name,omitempty"`
	Deadline     time.Time       `json:"deadline,omitempty"`
	TargetAmount decimal.Decimal `json:"target_amount,omitempty"`
}

func (r *GoalUpdate) Validate() error {
	if r.TargetAmount.IsNegative() {
		return errors.New("target_amount")
	}
	return nil
}

type GoalResponseList struct {
	Goals      []GoalResponse `json:"goals"`
	TotalCount int            `json:"total_count"`
}

type GoalResponse struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	UserID        uuid.UUID       `json:"-" db:"user_id"`
	Name          string          `json:"name" db:"name"`
	Currency      string          `json:"currency" db:"currency"`
	Deadline      *time.Time      `json:"deadline,omitempty" db:"deadline"`
	TargetAmount  decimal.Decimal `json:"target_amount" db:"target_amount"`
	CurrentAmount decimal.Decimal `json:"current_amount" db:"current_amount"`
}
