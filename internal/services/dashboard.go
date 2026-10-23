package services

import (
	"context"
	"time"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

type DashboardService struct {
	operationService        *OperationService
	freeBalanceService      *FreeBalanceService
	plannedOperationService *PlannedOperationService
}

func NewDashboardService(
	operationService *OperationService,
	freeBalanceService *FreeBalanceService,
	plannedOperationService *PlannedOperationService,
) *DashboardService {
	return &DashboardService{
		operationService:        operationService,
		freeBalanceService:      freeBalanceService,
		plannedOperationService: plannedOperationService,
	}
}

func (s *DashboardService) GetDashboard(ctx context.Context, userID uuid.UUID, currency string) (*dto.Dashboard, error) {
	now := time.Now()
	freeBalance, period, err := s.freeBalanceService.CalculateFreeBalance(ctx, userID, currency, now)
	if err != nil {
		return nil, err
	}

	var operations *dto.OperationResponseList
	var plannedOperations []dto.PlannedOperationResponse
	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		result, err := s.plannedOperationService.List(
			groupCtx,
			userID,
			period.start,
			period.end,
			repositories.ListParams{},
		)
		if err == nil {
			plannedOperations = result.PlannedOperations
		}
		return err
	})

	group.Go(func() error {
		result, err := s.operationService.List(groupCtx, userID, repositories.OperationListParams{
			StartDate: period.start,
			EndDate:   period.end,
		})
		if err == nil {
			operations = result
		}
		return err
	})

	if err := group.Wait(); err != nil {
		return nil, err
	}

	converter := currencyConverter{target: currency, rates: freeBalance.currencyRates}
	totalIncome, totalExpenses, err := calculateOperationTotals(operationResponses(operations.Operations), converter)
	if err != nil {
		return nil, err
	}

	plannedExpenses, err := calculatePlannedExpenses(plannedOperations, converter)
	if err != nil {
		return nil, err
	}

	return &dto.Dashboard{
		TotalBalance:     freeBalance.totalBalance,
		AvailableBalance: freeBalance.availableBalance,
		TotalIncome:      totalIncome,
		TotalExpenses:    totalExpenses,
		PlannedExpenses:  plannedExpenses,
		DateEnd:          period.end,
	}, nil
}

func operationResponses(items []any) []dto.OperationResponse {
	operations := make([]dto.OperationResponse, 0, len(items))
	for _, item := range items {
		switch operation := item.(type) {
		case dto.OperationResponse:
			operations = append(operations, operation)
		case dto.OperationResponseTransfer:
			operations = append(operations, operation.OperationResponse)
		case dto.OperationResponseReservation:
			operations = append(operations, operation.OperationResponse)
		}
	}
	return operations
}

func calculateOperationTotals(
	operations []dto.OperationResponse,
	converter currencyConverter,
) (decimal.Decimal, decimal.Decimal, error) {
	var income, expenses decimal.Decimal

	for _, operation := range operations {
		if operation.Type != models.OperationTypeIncome && operation.Type != models.OperationTypeExpense {
			continue
		}

		amount, err := converter.convert(operation.Amount, operation.Account.Currency)
		if err != nil {
			return decimal.Zero, decimal.Zero, err
		}

		switch operation.Type {
		case models.OperationTypeIncome:
			income = income.Add(amount)
		case models.OperationTypeExpense:
			expenses = expenses.Add(amount)
		}
	}
	return income, expenses, nil
}

func calculatePlannedExpenses(
	plannedOperations []dto.PlannedOperationResponse,
	converter currencyConverter,
) (decimal.Decimal, error) {
	expenses := decimal.Zero

	for _, operation := range plannedOperations {
		if operation.Type != models.OperationTypeExpense {
			continue
		}

		amount, err := converter.convert(operation.Amount, operation.Account.Currency)
		if err != nil {
			return decimal.Zero, err
		}

		expenses = expenses.Add(amount)
	}
	return expenses, nil
}
