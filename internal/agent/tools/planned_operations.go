package tools

import (
	"context"
	"errors"
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

	items := make([]agentdto.PlannedOperationItem, 0, len(listed))
	for _, op := range listed {
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

func (t *PlannedOperationsToolkit) Tools() ([]tool.Tool, error) {
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

func (t *PlannedOperationsToolkit) listTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "list_planned_operations",
		Description: "Запланированные доходы и расходы. Даты ограничивают окно следующего запуска.",
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

		op, err := newPlannedOperation(args)
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

func newPlannedOperation(args agentdto.CreatePlannedOperationArgs) (*dto.PlannedOperationPost, error) {
	name := strings.TrimSpace(args.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	amount, err := positiveAmount(args.Amount)
	if err != nil {
		return nil, err
	}

	opType := models.OperationType(strings.ToLower(strings.TrimSpace(args.Type)))
	if opType != models.OperationTypeIncome && opType != models.OperationTypeExpense {
		return nil, errors.New("type must be income or expense")
	}

	intervalType := models.IntervalType(strings.ToLower(strings.TrimSpace(args.IntervalType)))
	switch intervalType {
	case models.IntervalTypeDaily, models.IntervalTypeWeekly, models.IntervalTypeBiweekly, models.IntervalTypeMonthly, models.IntervalTypeCustom:
	default:
		return nil, errors.New("interval_type must be daily, weekly, biweekly, monthly, or custom")
	}
	if intervalType == models.IntervalTypeCustom && args.Interval < 1 {
		return nil, errors.New("interval must be at least 1 day when interval_type is custom")
	}

	accountID, err := requiredUUID(args.AccountID, "account_id")
	if err != nil {
		return nil, err
	}
	categoryID, err := requiredUUID(args.CategoryID, "category_id")
	if err != nil {
		return nil, err
	}

	plannedAt, err := parseOptionalTime(args.PlannedDate)
	if err != nil || plannedAt.IsZero() {
		return nil, errors.New("planned_date is required, use RFC3339 or YYYY-MM-DD")
	}

	return &dto.PlannedOperationPost{
		Name:         name,
		AccountID:    accountID,
		Amount:       amount,
		Type:         opType,
		IntervalType: intervalType,
		Interval:     args.Interval,
		CategoryID:   categoryID,
		PlannedAt:    plannedAt,
		IsRecurring:  args.IsRecurring,
	}, nil
}

func plannedOperationView(op dto.PlannedOperationResponse) agentdto.PlannedOperationItem {
	item := agentdto.PlannedOperationItem{
		ID:           op.ID.String(),
		Name:         op.Name,
		Amount:       op.Amount.String(),
		Type:         string(op.Type),
		AccountID:    op.Account.ID.String(),
		Account:      op.Account.Name,
		Currency:     op.Account.Currency,
		CategoryID:   op.Category.ID.String(),
		Category:     op.Category.Name,
		IntervalType: string(op.IntervalType),
		PlannedDate:  op.PlannedAt.UTC().Format(time.RFC3339),
		IsRecurring:  op.IsRecurring,
	}
	if op.NextRunAt != nil {
		item.NextRunDate = op.NextRunAt.UTC().Format(time.RFC3339)
	}
	return item
}
