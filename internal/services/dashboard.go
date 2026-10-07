package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"KopiBackend/internal/config"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

var errCurrencyRateNotFound = errors.New("currency_rate_not_found")

type DashboardService struct {
	config                  *config.AppConfig
	accountService          *AccountService
	operationService        *OperationService
	budgetService           *BudgetService
	plannedOperationService *PlannedOperationService
}

func NewDashboardService(
	config *config.AppConfig,
	accountService *AccountService,
	operationService *OperationService,
	budgetService *BudgetService,
	plannedOperationService *PlannedOperationService,
) *DashboardService {
	return &DashboardService{
		config:                  config,
		accountService:          accountService,
		operationService:        operationService,
		budgetService:           budgetService,
		plannedOperationService: plannedOperationService,
	}
}

func (s *DashboardService) GetDashboard(ctx context.Context, userID uuid.UUID, currency string) (*dto.Dashboard, error) {
	now := time.Now()
	period, err := s.dashboardPeriod(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	data, err := s.loadDashboardData(ctx, userID, currency, period)
	if err != nil {
		return nil, err
	}

	converter := currencyConverter{target: currency, rates: data.currencyRates}

	totalIncome, totalExpenses, err := calculateOperationTotals(data.operations, converter)
	if err != nil {
		return nil, err
	}

	accountReserve, err := calculateAccountReserve(data.accounts.Accounts, converter)
	if err != nil {
		return nil, err
	}

	plannedReserve, err := calculatePlannedReserve(data.plannedOperations, converter)
	if err != nil {
		return nil, err
	}

	budgetReserve, err := calculateBudgetReserve(data.budgets, period.end, now, converter)
	if err != nil {
		return nil, err
	}

	return &dto.Dashboard{
		TotalBalance: data.accounts.TotalBalance,
		AvailableBalance: data.accounts.TotalBalance.
			Sub(accountReserve).
			Sub(plannedReserve).
			Sub(budgetReserve),
		TotalIncome:     totalIncome,
		TotalExpenses:   totalExpenses,
		PlannedExpenses: plannedReserve,
		DateEnd:         period.end,
	}, nil
}

type dashboardPeriod struct {
	start time.Time
	end   time.Time
}

type dashboardData struct {
	accounts          dto.AccountResponseList
	operations        []dto.OperationResponse
	budgets           []dto.BudgetResponse
	plannedOperations []dto.PlannedOperationResponse
	currencyRates     map[string]decimal.Decimal
}

type currencyConverter struct {
	target string
	rates  map[string]decimal.Decimal
}

func (s *DashboardService) dashboardPeriod(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
) (dashboardPeriod, error) {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	nearestIncomeAt, err := s.plannedOperationService.GetNearestIncomeDate(ctx, userID, now)
	if err != nil {
		return dashboardPeriod{}, err
	}
	if nearestIncomeAt != nil {
		end = *nearestIncomeAt
	}

	return dashboardPeriod{start: start, end: end}, nil
}

func (s *DashboardService) loadDashboardData(
	ctx context.Context,
	userID uuid.UUID,
	currency string,
	period dashboardPeriod,
) (dashboardData, error) {
	var data dashboardData
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
			data.plannedOperations = result
		}
		return err
	})

	group.Go(func() error {
		result, err := s.accountService.List(groupCtx, userID, repositories.ListParams{}, &currency)
		if err == nil {
			data.accounts = *result
		}
		return err
	})

	group.Go(func() error {
		result, err := s.operationService.List(groupCtx, userID, repositories.OperationListParams{
			StartDate: period.start,
			EndDate:   period.end,
		})
		if err == nil {
			data.operations = operationResponses(result)
		}
		return err
	})

	group.Go(func() error {
		result, err := s.budgetService.List(groupCtx, userID, repositories.ListParams{})
		if err == nil {
			data.budgets = result
		}
		return err
	})

	group.Go(func() error {
		rates, err := GetCurrencyRates(s.config, currency)
		if err == nil {
			data.currencyRates = rates
		}
		return err
	})

	if err := group.Wait(); err != nil {
		return dashboardData{}, err
	}
	return data, nil
}

func operationResponses(items []any) []dto.OperationResponse {
	operations := make([]dto.OperationResponse, 0, len(items))
	for _, item := range items {
		switch operation := item.(type) {
		case dto.OperationResponse:
			operations = append(operations, operation)
		case dto.OperationResponseTransfer:
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

func calculateAccountReserve(
	accounts []dto.AccountResponse,
	converter currencyConverter,
) (decimal.Decimal, error) {
	var reserve decimal.Decimal

	for _, account := range accounts {
		if account.IncludeInFreeBalance {
			continue
		}

		amount, err := converter.convert(account.Balance, account.Currency)
		if err != nil {
			return decimal.Zero, err
		}
		reserve = reserve.Add(amount)
	}
	return reserve, nil
}

func calculatePlannedReserve(
	operations []dto.PlannedOperationResponse,
	converter currencyConverter,
) (decimal.Decimal, error) {
	var reserve decimal.Decimal

	for _, operation := range operations {
		if operation.Type != models.OperationTypeExpense {
			continue
		}

		amount, err := converter.convert(operation.Amount, operation.Account.Currency)
		if err != nil {
			return decimal.Zero, err
		}
		reserve = reserve.Add(amount)
	}
	return reserve, nil
}

func calculateBudgetReserve(
	budgets []dto.BudgetResponse,
	dateEnd time.Time,
	now time.Time,
	converter currencyConverter,
) (decimal.Decimal, error) {
	var total decimal.Decimal
	today := startOfDay(now, now.Location())

	for _, budget := range budgets {
		resetDay := startOfDay(budget.ResetDate, now.Location())
		horizon := startOfDay(dateEnd, now.Location())
		if !resetDay.After(today) || !horizon.After(today) {
			continue
		}

		balance, err := converter.convert(budget.Balance, budget.Currency)
		if err != nil {
			return decimal.Zero, err
		}

		daysLeft := daysBetween(today, resetDay)
		daysToReserve := daysBetween(today, horizon)
		reserve := balance.
			Div(decimal.NewFromInt(int64(daysLeft))).
			Mul(decimal.NewFromInt(int64(daysToReserve)))
		total = total.Add(reserve)
	}
	return total, nil
}

func (c currencyConverter) convert(amount decimal.Decimal, currency string) (decimal.Decimal, error) {
	if strings.EqualFold(currency, c.target) {
		return amount, nil
	}

	rate, ok := c.rates[strings.ToLower(currency)]
	if !ok || rate.IsZero() {
		return decimal.Zero, errCurrencyRateNotFound
	}
	return amount.Div(rate), nil
}

func startOfDay(date time.Time, location *time.Location) time.Time {
	date = date.In(location)
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, location)
}

func daysBetween(start, end time.Time) int {
	startDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)
	return int(endDate.Sub(startDate).Hours() / 24)
}
