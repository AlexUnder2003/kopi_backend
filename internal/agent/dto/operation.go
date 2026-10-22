package dto

import (
	"errors"
	"time"

	"KopiBackend/internal/models"

	"github.com/google/uuid"
)

type OperationItem struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Amount             string `json:"amount"`
	Type               string `json:"type"`
	Account            string `json:"account"`
	Currency           string `json:"currency"`
	FromAccount        string `json:"from_account,omitempty"`
	Category           string `json:"category"`
	PlannedOperationID string `json:"planned_operation_id,omitempty"`
	GoalID             string `json:"goal_id,omitempty"`
	OccurrenceDate     string `json:"occurrence_date"`
}

type ListOperationsArgs struct {
	Filter     string `json:"filter,omitempty" jsonschema:"Optional operation name substring."`
	AccountID  string `json:"account_id,omitempty" jsonschema:"Optional account UUID."`
	CategoryID string `json:"category_id,omitempty" jsonschema:"Optional category UUID."`
	GoalID     string `json:"goal_id,omitempty" jsonschema:"Optional goal UUID."`
	StartDate  string `json:"start_date,omitempty" jsonschema:"Optional start date, YYYY-MM-DD."`
	EndDate    string `json:"end_date,omitempty" jsonschema:"Optional end date, YYYY-MM-DD."`
	Limit      int    `json:"limit,omitempty" jsonschema:"Maximum number of rows. Defaults to 50, maximum 100."`
}

type CreateOperationArgs struct {
	Name               string `json:"name" jsonschema:"Short name of the operation, for example coffee or salary."`
	Amount             string `json:"amount" jsonschema:"Positive decimal amount, for example 500 or 12.50."`
	Type               string `json:"type" jsonschema:"income, expense, transfer, or reservation."`
	AccountID          string `json:"account_id" jsonschema:"Destination account UUID. For a transfer this is the account that receives money."`
	FromAccountID      string `json:"from_account_id,omitempty" jsonschema:"Source account UUID. Required for transfer."`
	CategoryID         string `json:"category_id" jsonschema:"Category UUID."`
	PlannedOperationID string `json:"planned_operation_id,omitempty" jsonschema:"Optional planned operation UUID this operation belongs to."`
	GoalID             string `json:"goal_id,omitempty" jsonschema:"Optional goal UUID this operation contributes to."`
	OccurrenceDate     string `json:"occurrence_date,omitempty" jsonschema:"When it happened, YYYY-MM-DD. Defaults to now."`
}

func (o *CreateOperationArgs) Model() (*models.Operation, error) {
	var p parsed
	operation := &models.Operation{
		Name:               p.requiredString(o.Name, "name"),
		Amount:             p.requiredAmount(o.Amount),
		Type:               p.requiredOperationType(o.Type, true),
		AccountID:          p.requiredUUID(o.AccountID, "account_id"),
		CategoryID:         p.requiredUUID(o.CategoryID, "category_id"),
		PlannedOperationID: p.optionalUUID(o.PlannedOperationID, "planned_operation_id"),
		GoalID:             p.optionalUUID(o.GoalID, "goal_id"),
		OccurrenceDate:     p.optionalTime(o.OccurrenceDate),
	}
	if operation.Type == models.OperationTypeTransfer {
		operation.FromAccountID = p.requiredUUID(o.FromAccountID, "from_account_id")
	}
	if operation.OccurrenceDate.IsZero() {
		operation.OccurrenceDate = time.Now().UTC()
	}
	if operation.Type == models.OperationTypeTransfer && operation.FromAccountID == operation.AccountID && p.err == nil {
		p.err = errors.New("from_account_id must differ from account_id")
	}
	if p.err != nil {
		return nil, p.err
	}
	return operation, nil
}

type UpdateOperationArgs struct {
	ID             string `json:"id" jsonschema:"Operation UUID from list_operations."`
	Name           string `json:"name,omitempty" jsonschema:"New name. Omit to keep the current name."`
	Amount         string `json:"amount,omitempty" jsonschema:"New positive decimal amount. Omit to keep the current amount."`
	Type           string `json:"type,omitempty" jsonschema:"income, expense, transfer, or reservation. Omit to keep the current type."`
	AccountID      string `json:"account_id,omitempty" jsonschema:"New account UUID. Omit to keep the current account."`
	FromAccountID  string `json:"from_account_id,omitempty" jsonschema:"New source account UUID for a transfer. Omit to keep the current source."`
	CategoryID     string `json:"category_id,omitempty" jsonschema:"New category UUID. Omit to keep the current category."`
	OccurrenceDate string `json:"occurrence_date,omitempty" jsonschema:"New date, YYYY-MM-DD or RFC3339. Omit to keep the current date."`
}

func (o *UpdateOperationArgs) Model() (*models.Operation, error) {
	var p parsed
	operation := &models.Operation{
		ID:             p.requiredUUID(o.ID, "id"),
		Name:           p.string(o.Name),
		Amount:         p.optionalAmount(o.Amount),
		Type:           p.optionalOperationType(o.Type, true),
		AccountID:      p.optionalUUID(o.AccountID, "account_id"),
		FromAccountID:  p.optionalUUID(o.FromAccountID, "from_account_id"),
		CategoryID:     p.optionalUUID(o.CategoryID, "category_id"),
		OccurrenceDate: p.optionalTime(o.OccurrenceDate),
	}
	if p.err != nil {
		return nil, p.err
	}
	return operation, nil
}

func (o *DeleteOperationArgs) UUID() (uuid.UUID, error) {
	var p parsed
	id := p.requiredUUID(o.ID, "id")
	return id, p.err
}

type DeleteOperationArgs struct {
	ID string `json:"id" jsonschema:"Operation UUID from list_operations."`
}

type DeletedOperation struct {
	ID string `json:"id"`
}
