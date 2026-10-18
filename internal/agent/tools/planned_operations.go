package tools

import (
	"context"
	"errors"
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

type PlannedOperationsToolkit struct {
	plannedOperationsService *services.PlannedOperationService
}

func NewPlannedOperationsToolkit(plannedOperationsService *services.PlannedOperationService) *PlannedOperationsToolkit {
	return &PlannedOperationsToolkit{plannedOperationsService: plannedOperationsService}
}

func (t *PlannedOperationsToolkit) ListPlannedOperations(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, params repositories.ListParams) ([]agentdto.PlannedOperationItem, error) {
	listed, err := t.plannedOperationsService.List(ctx, userID, startDate, endDate, params)
	if err != nil {
		return nil, err
	}

	items := make([]agentdto.PlannedOperationItem, 0, len(listed.PlannedOperations))
	for _, op := range listed.PlannedOperations {
		items = append(items, plannedOperationView(op))
	}
	return items, nil
}

func (t *PlannedOperationsToolkit) CreatePlannedOperation(ctx context.Context, userID uuid.UUID, op *dto.PlannedOperationPost) (*dto.PlannedOperationResponse, error) {
	created, err := t.plannedOperationsService.Create(ctx, userID, op)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (t *PlannedOperationsToolkit) UpdatePlannedOperation(ctx context.Context, id, userID uuid.UUID, op *dto.PlannedOperationUpdate) (*dto.PlannedOperationResponse, error) {
	return t.plannedOperationsService.Update(ctx, id, userID, op)
}

func (t *PlannedOperationsToolkit) DeletePlannedOperation(ctx context.Context, id, userID uuid.UUID) error {
	return t.plannedOperationsService.Delete(ctx, id, userID)
}

func (t *PlannedOperationsToolkit) Tools() ([]tool.Tool, error) {
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

func (t *PlannedOperationsToolkit) listTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "list_planned_operations",
		Description: "Запланированные доходы и расходы. Id нужен для изменения и удаления. Даты ограничивают окно следующего запуска.",
	}, func(ctx adkagent.Context, args agentdto.ListPlannedOperationsArgs) ([]agentdto.PlannedOperationItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, errors.New("failed to list planned operations")
		}
		start, err := parseOptionalTime(args.StartDate)
		if err != nil {
			return nil, errors.New("start_date: use YYYY-MM-DD")
		}
		end, err := parseOptionalTime(args.EndDate)
		if err != nil {
			return nil, errors.New("end_date: use YYYY-MM-DD")
		}

		return t.ListPlannedOperations(ctx, userID, start, end, repositories.ListParams{
			Filter: strings.TrimSpace(args.Filter),
			Limit:  boundedLimit(args.Limit),
		})
	})
}

func (t *PlannedOperationsToolkit) createTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "create_planned_operation",
		Description: "Планирует будущий доход или расход, разово или по расписанию.",
	}, func(ctx adkagent.Context, args agentdto.CreatePlannedOperationArgs) (agentdto.PlannedOperationItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return agentdto.PlannedOperationItem{}, err
		}

		op, err := args.Model()
		if err != nil {
			return agentdto.PlannedOperationItem{}, err
		}

		created, err := t.CreatePlannedOperation(ctx, userID, op)
		if err != nil {
			return agentdto.PlannedOperationItem{}, err
		}
		return plannedOperationView(*created), nil
	})
}

func (t *PlannedOperationsToolkit) updateTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "update_planned_operation",
		Description: "Изменяет запланированный доход или расход. Передавай только поля, которые нужно поменять.",
	}, func(ctx adkagent.Context, args agentdto.UpdatePlannedOperationArgs) (agentdto.PlannedOperationItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return agentdto.PlannedOperationItem{}, err
		}

		id, op, err := args.Model()
		if err != nil {
			return agentdto.PlannedOperationItem{}, err
		}

		updated, err := t.UpdatePlannedOperation(ctx, id, userID, op)
		if err != nil {
			return agentdto.PlannedOperationItem{}, err
		}
		return plannedOperationView(*updated), nil
	})
}

func (t *PlannedOperationsToolkit) deleteTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "delete_planned_operation",
		Description: "Удаляет запланированный доход или расход. Уже созданные операции не затрагиваются.",
	}, func(ctx adkagent.Context, args agentdto.DeletePlannedOperationArgs) (agentdto.DeletedPlannedOperation, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return agentdto.DeletedPlannedOperation{}, err
		}

		id, err := args.UUID()
		if err != nil {
			return agentdto.DeletedPlannedOperation{}, err
		}

		if err := t.DeletePlannedOperation(ctx, id, userID); err != nil {
			return agentdto.DeletedPlannedOperation{}, err
		}
		return agentdto.DeletedPlannedOperation{ID: id.String()}, nil
	})
}

func plannedOperationView(op dto.PlannedOperationResponse) agentdto.PlannedOperationItem {
	item := agentdto.PlannedOperationItem{
		ID:          op.ID.String(),
		Name:        op.Name,
		Amount:      op.Amount.String(),
		Type:        string(op.Type),
		AccountID:   op.Account.ID.String(),
		Account:     op.Account.Name,
		Currency:    op.Account.Currency,
		CategoryID:  op.Category.ID.String(),
		Category:    op.Category.Name,
		PlannedDate: op.PlannedAt.UTC().Format(time.RFC3339),
		IsRecurring: op.IsRecurring,
	}
	if op.IntervalType != nil {
		item.IntervalType = string(*op.IntervalType)
	}
	if op.NextRunAt != nil {
		item.NextRunDate = op.NextRunAt.UTC().Format(time.RFC3339)
	}
	return item
}
