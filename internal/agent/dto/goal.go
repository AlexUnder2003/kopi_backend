package dto

import (
	appdto "KopiBackend/internal/dto"

	"github.com/google/uuid"
)

type GoalItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Currency      string `json:"currency"`
	TargetAmount  string `json:"target_amount"`
	CurrentAmount string `json:"current_amount"`
	Deadline      string `json:"deadline,omitempty"`
}

type ListGoalsArgs struct {
	Filter string `json:"filter,omitempty" jsonschema:"Optional goal name substring."`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of rows. Defaults to 50, maximum 100."`
}

type CreateGoalArgs struct {
	Name          string `json:"name" jsonschema:"Short name, for example vacation or emergency fund."`
	Currency      string `json:"currency" jsonschema:"USD, EUR, RUB, KZT, or BYN. Reservations must use an account in this currency."`
	TargetAmount  string `json:"target_amount" jsonschema:"Positive decimal amount to save."`
	CurrentAmount string `json:"current_amount,omitempty" jsonschema:"Already saved amount. Defaults to 0."`
	Deadline      string `json:"deadline,omitempty" jsonschema:"Optional date, YYYY-MM-DD or RFC3339."`
}

func (o *CreateGoalArgs) Model() (*appdto.GoalPost, error) {
	var p parsed
	goal := &appdto.GoalPost{
		Name:          p.requiredString(o.Name, "name"),
		Currency:      p.requiredCurrency(o.Currency),
		TargetAmount:  p.requiredAmount(o.TargetAmount),
		CurrentAmount: p.optionalNonNegative(o.CurrentAmount, "current_amount"),
		Deadline:      p.optionalTime(o.Deadline),
	}
	if p.err != nil {
		return nil, p.err
	}
	return goal, nil
}

type UpdateGoalArgs struct {
	ID           string `json:"id" jsonschema:"Goal UUID from list_goals."`
	Name         string `json:"name,omitempty" jsonschema:"New name. Omit to keep the current name."`
	TargetAmount string `json:"target_amount,omitempty" jsonschema:"New positive decimal target. Omit to keep the current target."`
	Deadline     string `json:"deadline,omitempty" jsonschema:"New date, YYYY-MM-DD or RFC3339. Omit to keep the current deadline."`
}

func (o *UpdateGoalArgs) Model() (uuid.UUID, *appdto.GoalUpdate, error) {
	var p parsed
	id := p.requiredUUID(o.ID, "id")
	goal := &appdto.GoalUpdate{
		Name:         p.string(o.Name),
		TargetAmount: p.optionalAmount(o.TargetAmount),
		Deadline:     p.optionalTime(o.Deadline),
	}
	if p.err != nil {
		return uuid.Nil, nil, p.err
	}
	return id, goal, nil
}

func (o *DeleteGoalArgs) UUID() (uuid.UUID, error) {
	var p parsed
	id := p.requiredUUID(o.ID, "id")
	return id, p.err
}

type DeleteGoalArgs struct {
	ID string `json:"id" jsonschema:"Goal UUID from list_goals."`
}

type DeletedGoal struct {
	ID string `json:"id"`
}
