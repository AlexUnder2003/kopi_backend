package tools

import (
	"context"
	"strings"
	"time"

	agentdto "KopiBackend/internal/agent/dto"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/repositories"
	"KopiBackend/internal/services"

	"github.com/google/uuid"
	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type GoalsToolkit struct {
	goalsService *services.GoalService
}

func NewGoalsToolkit(goalsService *services.GoalService) *GoalsToolkit {
	return &GoalsToolkit{goalsService: goalsService}
}

func (t *GoalsToolkit) ListGoals(ctx context.Context, userID uuid.UUID, params repositories.ListParams) ([]agentdto.GoalItem, error) {
	listed, err := t.goalsService.List(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	items := make([]agentdto.GoalItem, 0, len(listed.Goals))
	for _, goal := range listed.Goals {
		items = append(items, goalView(goal))
	}
	return items, nil
}

func (t *GoalsToolkit) CreateGoal(ctx context.Context, userID uuid.UUID, goal *dto.GoalPost) (*dto.GoalResponse, error) {
	return t.goalsService.Create(ctx, userID, goal)
}

func (t *GoalsToolkit) UpdateGoal(ctx context.Context, id, userID uuid.UUID, goal *dto.GoalUpdate) (*dto.GoalResponse, error) {
	return t.goalsService.Update(ctx, id, userID, goal)
}

func (t *GoalsToolkit) DeleteGoal(ctx context.Context, id, userID uuid.UUID) error {
	return t.goalsService.Delete(ctx, id, userID)
}

func (t *GoalsToolkit) Tools() ([]tool.Tool, error) {
	list, err := t.listTool()
	if err != nil {
		return nil, err
	}
	create, err := t.createTool()
	if err != nil {
		return nil, err
	}
	update, err := t.updateTool()
	if err != nil {
		return nil, err
	}
	del, err := t.deleteTool()
	if err != nil {
		return nil, err
	}
	return []tool.Tool{list, create, update, del}, nil
}

func (t *GoalsToolkit) listTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "list_goals",
		Description: "Цели накопления: имя, валюта, цель и уже отложенная сумма. Id нужен, чтобы отложить деньги резервацией.",
	}, func(ctx adkagent.Context, args agentdto.ListGoalsArgs) ([]agentdto.GoalItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, err
		}

		return t.ListGoals(ctx, userID, repositories.ListParams{
			Filter: strings.TrimSpace(args.Filter),
			Limit:  boundedLimit(args.Limit),
		})
	})
}

func (t *GoalsToolkit) createTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "create_goal",
		Description: "Создаёт цель накопления. Деньги на неё откладывает отдельная операция type=reservation с этим goal_id.",
	}, func(ctx adkagent.Context, args agentdto.CreateGoalArgs) (agentdto.GoalItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return agentdto.GoalItem{}, err
		}

		goal, err := args.Model()
		if err != nil {
			return agentdto.GoalItem{}, err
		}

		created, err := t.CreateGoal(ctx, userID, goal)
		if err != nil {
			return agentdto.GoalItem{}, err
		}
		return goalView(*created), nil
	})
}

func (t *GoalsToolkit) updateTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "update_goal",
		Description: "Изменяет имя, сумму или срок цели. Передавай только поля, которые нужно поменять. Уже отложенную сумму меняет резервация.",
	}, func(ctx adkagent.Context, args agentdto.UpdateGoalArgs) (agentdto.GoalItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return agentdto.GoalItem{}, err
		}

		id, goal, err := args.Model()
		if err != nil {
			return agentdto.GoalItem{}, err
		}

		updated, err := t.UpdateGoal(ctx, id, userID, goal)
		if err != nil {
			return agentdto.GoalItem{}, err
		}
		return goalView(*updated), nil
	})
}

func (t *GoalsToolkit) deleteTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "delete_goal",
		Description: "Удаляет цель. Связанные резервации остаются, но больше не ссылаются на неё.",
	}, func(ctx adkagent.Context, args agentdto.DeleteGoalArgs) (agentdto.DeletedGoal, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return agentdto.DeletedGoal{}, err
		}

		id, err := args.UUID()
		if err != nil {
			return agentdto.DeletedGoal{}, err
		}

		if err := t.DeleteGoal(ctx, id, userID); err != nil {
			return agentdto.DeletedGoal{}, err
		}
		return agentdto.DeletedGoal{ID: id.String()}, nil
	})
}

func goalView(goal dto.GoalResponse) agentdto.GoalItem {
	return agentdto.GoalItem{
		ID:            goal.ID.String(),
		Name:          goal.Name,
		Currency:      goal.Currency,
		TargetAmount:  goal.TargetAmount.String(),
		CurrentAmount: goal.CurrentAmount.String(),
		Deadline:      goalDeadline(goal.Deadline),
	}
}

func goalDeadline(deadline *time.Time) string {
	if deadline == nil || deadline.IsZero() {
		return ""
	}
	return deadline.UTC().Format(time.RFC3339)
}
