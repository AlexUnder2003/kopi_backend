package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	agentdto "KopiBackend/internal/agent/dto"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"
	"KopiBackend/internal/services"

	"github.com/google/uuid"
	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type OperationsToolkit struct {
	operationsService *services.OperationService
}

func NewOperationsToolkit(operationsService *services.OperationService) *OperationsToolkit {
	return &OperationsToolkit{
		operationsService: operationsService,
	}
}

func (t *OperationsToolkit) ListOperations(ctx context.Context, userID uuid.UUID, params repositories.OperationListParams) ([]agentdto.OperationItem, error) {
	listed, err := t.operationsService.List(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	items := make([]agentdto.OperationItem, 0, len(listed.Operations))
	for _, item := range listed.Operations {
		view, ok := operationView(item)
		if ok {
			items = append(items, view)
		}
	}
	return items, nil
}

func (t *OperationsToolkit) CreateOperation(ctx context.Context, userID uuid.UUID, operation *models.Operation) (any, error) {
	return t.operationsService.Create(ctx, userID, operation)
}

func (t *OperationsToolkit) UpdateOperation(ctx context.Context, userID uuid.UUID, operation *models.Operation) (any, error) {
	return t.operationsService.Update(ctx, userID, operation)
}

func (t *OperationsToolkit) DeleteOperation(ctx context.Context, userID, id uuid.UUID) error {
	return t.operationsService.Delete(ctx, id, userID)
}

func (t *OperationsToolkit) Tools() ([]tool.Tool, error) {
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

func (t *OperationsToolkit) listTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "list_operations",
		Description: "Операции пользователя: доходы, расходы и переводы. Id нужен для изменения и удаления. Фильтры по имени, счёту, категории и датам необязательны.",
	}, func(ctx adkagent.Context, args agentdto.ListOperationsArgs) ([]agentdto.OperationItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, errors.New("failed to list operations")
		}

		params, err := listParams(args)
		if err != nil {
			return nil, errors.New("failed to list operations")
		}

		return t.ListOperations(ctx, userID, params)
	})
}

func (t *OperationsToolkit) createTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "create_operation",
		Description: "Создаёт доход, расход или перевод между счетами.",
	}, func(ctx adkagent.Context, args agentdto.CreateOperationArgs) (agentdto.OperationItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return agentdto.OperationItem{}, errors.New("failed to create operation")
		}

		operation, err := args.Model()
		if err != nil {
			return agentdto.OperationItem{}, errors.New("failed to create operation")
		}

		created, err := t.CreateOperation(ctx, userID, operation)
		if err != nil {
			return agentdto.OperationItem{}, errors.New("failed to create operation")
		}

		view, ok := operationView(created)
		if !ok {
			return agentdto.OperationItem{}, errors.New("unexpected operation response")
		}
		return view, nil
	})
}

func (t *OperationsToolkit) updateTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "update_operation",
		Description: "Изменяет доход, расход или перевод. Передавай только поля, которые нужно поменять.",
	}, func(ctx adkagent.Context, args agentdto.UpdateOperationArgs) (agentdto.OperationItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return agentdto.OperationItem{}, errors.New("failed to update operation")
		}

		operation, err := args.Model()
		if err != nil {
			return agentdto.OperationItem{}, err
		}

		updated, err := t.UpdateOperation(ctx, userID, operation)
		if err != nil {
			return agentdto.OperationItem{}, errors.New("failed to update operation")
		}

		view, ok := operationView(updated)
		if !ok {
			return agentdto.OperationItem{}, errors.New("unexpected operation response")
		}
		return view, nil
	})
}

func (t *OperationsToolkit) deleteTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "delete_operation",
		Description: "Удаляет операцию и откатывает её влияние на баланс счёта.",
	}, func(ctx adkagent.Context, args agentdto.DeleteOperationArgs) (agentdto.DeletedOperation, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return agentdto.DeletedOperation{}, errors.New("failed to delete operation")
		}

		id, err := args.UUID()
		if err != nil {
			return agentdto.DeletedOperation{}, err
		}

		if err := t.DeleteOperation(ctx, userID, id); err != nil {
			return agentdto.DeletedOperation{}, errors.New("failed to delete operation")
		}
		return agentdto.DeletedOperation{ID: id.String()}, nil
	})
}

func listParams(args agentdto.ListOperationsArgs) (repositories.OperationListParams, error) {
	accountID, err := optionalUUID(args.AccountID, "account_id")
	if err != nil {
		return repositories.OperationListParams{}, err
	}
	categoryID, err := optionalUUID(args.CategoryID, "category_id")
	if err != nil {
		return repositories.OperationListParams{}, err
	}
	start, err := parseOptionalTime(args.StartDate)
	if err != nil {
		return repositories.OperationListParams{}, fmt.Errorf("start_date: %w", err)
	}
	end, err := parseOptionalTime(args.EndDate)
	if err != nil {
		return repositories.OperationListParams{}, fmt.Errorf("end_date: %w", err)
	}

	return repositories.OperationListParams{
		ListParams: repositories.ListParams{
			Filter: strings.TrimSpace(args.Filter),
			Limit:  boundedLimit(args.Limit),
		},
		AccountID:  accountID,
		CategoryID: categoryID,
		StartDate:  start,
		EndDate:    end,
	}, nil
}

func operationView(item any) (agentdto.OperationItem, bool) {
	switch op := item.(type) {
	case dto.OperationResponse:
		return agentdto.OperationItem{
			ID:             op.ID.String(),
			Name:           op.Name,
			Amount:         op.Amount.String(),
			Type:           string(op.Type),
			Account:        op.Account.Name,
			Currency:       op.Account.Currency,
			Category:       op.Category.Name,
			OccurrenceDate: op.OccurrenceDate.UTC().Format(time.RFC3339),
		}, true
	case dto.OperationResponseTransfer:
		view, _ := operationView(op.OperationResponse)
		view.FromAccount = op.FromAccount.Name
		return view, true
	default:
		return agentdto.OperationItem{}, false
	}
}
