package dto

import (
	"errors"
	"time"

	"KopiBackend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (r *OperationPost) Validate() error {
	if r.Name == "" {
		return errors.New("name")
	}
	if !r.Amount.IsPositive() {
		return errors.New("amount")
	}
	if !validOperationType(r.Type) {
		return errors.New("type")
	}
	if r.OccurrenceDate.IsZero() {
		return errors.New("occurrence_date")
	}
	if r.Type == models.OperationTypeTransfer && (r.FromAccountID == uuid.Nil || r.FromAccountID == r.AccountID) {
		return errors.New("from_account_id")
	}
	return nil
}

func (r *OperationUpdate) Validate() error {
	if r.Type != "" && !validOperationType(r.Type) {
		return errors.New("type")
	}
	return nil
}

type OperationPost struct {
	Name               string               `json:"name"`
	Amount             decimal.Decimal      `json:"amount"`
	Type               models.OperationType `json:"type"`
	AccountID          uuid.UUID            `json:"account_id,omitempty"`
	FromAccountID      uuid.UUID            `json:"from_account_id,omitempty"`
	CategoryID         uuid.UUID            `json:"category_id"`
	PlannedOperationID uuid.UUID            `json:"planned_operation_id,omitempty"`
	GoalID             uuid.UUID            `json:"goal_id,omitempty"`
	OccurrenceDate     time.Time            `json:"occurrence_date"`
}

type OperationUpdate struct {
	Name           string               `json:"name,omitempty"`
	Amount         decimal.Decimal      `json:"amount,omitempty"`
	Type           models.OperationType `json:"type,omitempty"`
	CategoryID     uuid.UUID            `json:"category_id,omitempty"`
	OccurrenceDate time.Time            `json:"occurrence_date,omitempty"`
}

type OperationResponseList struct {
	Operations []any `json:"operations"`
	TotalCount int   `json:"total_count"`
}

type OperationResponse struct {
	ID                 uuid.UUID            `json:"id" db:"id"`
	Name               string               `json:"name" db:"name"`
	Amount             decimal.Decimal      `json:"amount" db:"amount"`
	Type               models.OperationType `json:"type" db:"type"`
	Account            AccountResponseShort `json:"account,omitempty" db:"account"`
	Category           CategoryResponse     `json:"category" db:"category"`
	PlannedOperationID *uuid.UUID           `json:"planned_operation_id,omitempty" db:"planned_operation_id"`
	GoalID             *uuid.UUID           `json:"goal_id,omitempty" db:"goal_id"`
	OccurrenceDate     time.Time            `json:"occurrence_date" db:"occurrence_date"`
}

type OperationResponseReservation struct {
	OperationResponse
	PlannedOperation *PlannedOperationResponseShort `json:"planned_operation,omitempty"`
	Goal             *GoalResponse                  `json:"goal,omitempty"`
}

type OperationResponseTransfer struct {
	OperationResponse
	FromAccount      AccountResponseShort          `json:"from_account" db:"from_account"`
	PlannedOperation PlannedOperationResponseShort `json:"-" db:"planned_operation"`
	Goal             GoalResponse                  `json:"-" db:"goal"`
}

func (r OperationResponseTransfer) Response() any {
	switch r.Type {
	case models.OperationTypeTransfer:
		return r
	case models.OperationTypeReservation:
		reservation := OperationResponseReservation{OperationResponse: r.OperationResponse}
		if r.PlannedOperation.ID != uuid.Nil {
			planned := r.PlannedOperation
			reservation.PlannedOperation = &planned
		} else if r.Goal.ID != uuid.Nil {
			goal := r.Goal
			reservation.Goal = &goal
		}
		return reservation
	default:
		return r.OperationResponse
	}
}
