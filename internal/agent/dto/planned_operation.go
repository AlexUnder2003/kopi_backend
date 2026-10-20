package dto

import (
	"errors"

	appdto "KopiBackend/internal/dto"
	"KopiBackend/internal/models"

	"github.com/google/uuid"
)

type PlannedOperationItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Amount       string `json:"amount"`
	Balance      string `json:"balance"`
	Type         string `json:"type"`
	AccountID    string `json:"account_id"`
	Account      string `json:"account"`
	Currency     string `json:"currency"`
	CategoryID   string `json:"category_id"`
	Category     string `json:"category"`
	IntervalType string `json:"interval_type,omitempty"`
	PlannedDate  string `json:"planned_date"`
	NextRunDate  string `json:"next_run_date,omitempty"`
	IsRecurring  bool   `json:"is_recurring"`
}

type ListPlannedOperationsArgs struct {
	Filter    string `json:"filter,omitempty" jsonschema:"Optional planned operation name substring."`
	StartDate string `json:"start_date,omitempty" jsonschema:"Optional start of the next run window, YYYY-MM-DD."`
	EndDate   string `json:"end_date,omitempty" jsonschema:"Optional end of the next run window, YYYY-MM-DD."`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of rows. Defaults to 50, maximum 100."`
}

type CreatePlannedOperationArgs struct {
	Name         string `json:"name" jsonschema:"Short name, for example rent."`
	Amount       string `json:"amount" jsonschema:"Positive decimal amount."`
	Type         string `json:"type" jsonschema:"income, expense, or reservation."`
	AccountID    string `json:"account_id" jsonschema:"Account UUID from list_accounts."`
	CategoryID   string `json:"category_id" jsonschema:"Category UUID from list_categories."`
	IntervalType string `json:"interval_type,omitempty" jsonschema:"Optional. daily, weekly, biweekly, monthly, or custom. Omit for a one-time plan."`
	Interval     int    `json:"interval,omitempty" jsonschema:"Number of days. Required when interval_type is custom."`
	PlannedDate  string `json:"planned_date" jsonschema:"First run date, YYYY-MM-DD."`
	IsRecurring  bool   `json:"is_recurring,omitempty" jsonschema:"Repeat on the interval. Defaults to false."`
}

func (o *CreatePlannedOperationArgs) Model() (*appdto.PlannedOperationPost, error) {
	var p parsed
	intervalType := p.optionalIntervalType(o.IntervalType)
	if intervalType == models.IntervalTypeCustom && o.Interval < 1 && p.err == nil {
		p.err = errors.New("interval must be at least 1 day when interval_type is custom")
	}
	operation := &appdto.PlannedOperationPost{
		Name:         p.requiredString(o.Name, "name"),
		AccountID:    p.requiredUUID(o.AccountID, "account_id"),
		Amount:       p.requiredAmount(o.Amount),
		Type:         p.requiredOperationType(o.Type, false),
		IntervalType: intervalType,
		Interval:     o.Interval,
		CategoryID:   p.requiredUUID(o.CategoryID, "category_id"),
		PlannedAt:    p.optionalTime(o.PlannedDate),
		IsRecurring:  o.IsRecurring,
	}
	if operation.PlannedAt.IsZero() && p.err == nil {
		p.err = errors.New("planned_date is required")
	}
	if p.err != nil {
		return nil, p.err
	}
	return operation, nil
}

type UpdatePlannedOperationArgs struct {
	ID           string `json:"id" jsonschema:"Planned operation UUID from list_planned_operations."`
	Name         string `json:"name,omitempty" jsonschema:"New name. Omit to keep the current name."`
	Amount       string `json:"amount,omitempty" jsonschema:"New positive decimal amount. Omit to keep the current amount."`
	Type         string `json:"type,omitempty" jsonschema:"income, expense, or reservation. Omit to keep the current type."`
	AccountID    string `json:"account_id,omitempty" jsonschema:"New account UUID. Omit to keep the current account."`
	CategoryID   string `json:"category_id,omitempty" jsonschema:"New category UUID. Omit to keep the current category."`
	IntervalType string `json:"interval_type,omitempty" jsonschema:"daily, weekly, biweekly, monthly, or custom. Omit to keep the current interval."`
	Interval     int    `json:"interval,omitempty" jsonschema:"Number of days. Required when setting interval_type to custom."`
	PlannedDate  string `json:"planned_date,omitempty" jsonschema:"New first run date, YYYY-MM-DD or RFC3339. Omit to keep the current date."`
	IsRecurring  *bool  `json:"is_recurring,omitempty" jsonschema:"Whether the plan repeats. Omit to keep the current value."`
}

func (o *UpdatePlannedOperationArgs) Model() (uuid.UUID, *appdto.PlannedOperationUpdate, error) {
	var p parsed
	id := p.requiredUUID(o.ID, "id")
	intervalType := p.optionalIntervalType(o.IntervalType)
	if intervalType == models.IntervalTypeCustom && o.Interval < 1 && p.err == nil {
		p.err = errors.New("interval must be at least 1 day when interval_type is custom")
	}
	operation := &appdto.PlannedOperationUpdate{
		Name:         p.string(o.Name),
		AccountID:    p.optionalUUID(o.AccountID, "account_id"),
		Amount:       p.optionalAmount(o.Amount),
		Type:         p.optionalOperationType(o.Type, false),
		IntervalType: intervalType,
		Interval:     o.Interval,
		CategoryID:   p.optionalUUID(o.CategoryID, "category_id"),
		PlannedAt:    p.optionalTime(o.PlannedDate),
		IsRecurring:  o.IsRecurring,
	}
	if p.err != nil {
		return uuid.Nil, nil, p.err
	}
	return id, operation, nil
}

func (o *DeletePlannedOperationArgs) UUID() (uuid.UUID, error) {
	var p parsed
	id := p.requiredUUID(o.ID, "id")
	return id, p.err
}

type DeletePlannedOperationArgs struct {
	ID string `json:"id" jsonschema:"Planned operation UUID from list_planned_operations."`
}

type DeletedPlannedOperation struct {
	ID string `json:"id"`
}
