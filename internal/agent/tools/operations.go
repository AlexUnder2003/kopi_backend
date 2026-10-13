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

	items := make([]agentdto.OperationItem, 0, len(listed))
	for _, item := range listed {
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

func (t *OperationsToolkit) Tools() ([]tool.Tool, error) {
	list, err := t.listTool()
	if err != nil {
		return nil, err
	}
	create, err := t.createTool()
	if err != nil {
		return nil, err
	}
	return []tool.Tool{list, create}, nil
}

func (t *OperationsToolkit) listTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "list_operations",
		Description: "Операции пользователя: доходы, расходы и переводы. Фильтры по имени, счёту, категории и датам необязательны.",
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

		operation, err := newOperation(args)
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

func newOperation(args agentdto.CreateOperationArgs) (*models.Operation, error) {
	name := strings.TrimSpace(args.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	amount, err := positiveAmount(args.Amount)
	if err != nil {
		return nil, err
	}

	opType := models.OperationType(strings.ToLower(strings.TrimSpace(args.Type)))
	switch opType {
	case models.OperationTypeIncome, models.OperationTypeExpense, models.OperationTypeTransfer:
	default:
		return nil, errors.New("type must be income, expense, or transfer")
	}

	accountID, err := requiredUUID(args.AccountID, "account_id")
	if err != nil {
		return nil, err
	}
	categoryID, err := requiredUUID(args.CategoryID, "category_id")
	if err != nil {
		return nil, err
	}

	var fromAccountID uuid.UUID
	if opType == models.OperationTypeTransfer {
		fromAccountID, err = requiredUUID(args.FromAccountID, "from_account_id")
		if err != nil {
			return nil, err
		}
	}

	occurred, err := parseOptionalTime(args.OccurrenceDate)
	if err != nil {
		return nil, fmt.Errorf("occurrence_date: %w", err)
	}
	if occurred.IsZero() {
		occurred = time.Now().UTC()
	}

	return &models.Operation{
		Name:           name,
		Amount:         amount,
		Type:           opType,
		AccountID:      accountID,
		FromAccountID:  fromAccountID,
		CategoryID:     categoryID,
		OccurrenceDate: occurred,
	}, nil
}

func operationView(item any) (agentdto.OperationItem, bool) {
	switch op := item.(type) {
	case dto.OperationResponse:
		return agentdto.OperationItem{
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
