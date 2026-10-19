package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OperationType string

const (
	OperationTypeIncome      OperationType = "income"
	OperationTypeExpense     OperationType = "expense"
	OperationTypeTransfer    OperationType = "transfer"
	OperationTypeReservation OperationType = "reservation"
)

type Operation struct {
	ID                 uuid.UUID       `db:"id"`
	Name               string          `db:"name"`
	Type               OperationType   `db:"type"`
	AccountID          uuid.UUID       `db:"account_id"`
	FromAccountID      uuid.UUID       `db:"from_account_id"`
	CategoryID         uuid.UUID       `db:"category_id"`
	PlannedOperationID uuid.UUID       `db:"planned_operation_id"`
	Amount             decimal.Decimal `db:"amount"`
	OccurrenceDate     time.Time       `db:"occurrence_date"`
}
