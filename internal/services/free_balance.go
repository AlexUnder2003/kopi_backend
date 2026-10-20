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

type FreeBalanceService struct {
	config                  *config.AppConfig
	accountService          *AccountService
	budgetService           *BudgetService
	plannedOperationService *PlannedOperationService
}

func NewFreeBalanceService(
	config *config.AppConfig,
	accountService *AccountService,
	budgetService *BudgetService,
	plannedOperationService *PlannedOperationService,
) *FreeBalanceService {
	return &FreeBalanceService{
		config:                  config,
		accountService:          accountService,
		budgetService:           budgetService,
		plannedOperationService: plannedOperationService,
	}
}

type freeBalancePeriod struct {
	start time.Time
	end   time.Time
}

type freeBalanceData struct {
	plannedBalance    decimal.Decimal
	accounts          dto.AccountResponseList
	budgets           []dto.BudgetResponse
	plannedOperations []dto.PlannedOperationResponse
	currencyRates     map[string]decimal.Decimal
}

type freeBalanceResult struct {
	totalBalance     decimal.Decimal
	availableBalance decimal.Decimal
	plannedExpenses  decimal.Decimal
	currencyRates    map[string]decimal.Decimal
}

type currencyConverter struct {
	target string
	rates  map[string]decimal.Decimal
}

func (s *FreeBalanceService) GetPeriod(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
) (freeBalancePeriod, error) {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	nearestIncomeAt, err := s.plannedOperationService.GetNearestIncomeDate(ctx, userID, now)
	if err != nil {
		return freeBalancePeriod{}, err
	}
	if nearestIncomeAt != nil {
		end = *nearestIncomeAt
	}

	return freeBalancePeriod{start: start, end: end}, nil
}

func (s *FreeBalanceService) Calculate(
	ctx context.Context,
	userID uuid.UUID,
	currency string,
	period freeBalancePeriod,
	now time.Time,
) (freeBalanceResult, error) {
	data, err := s.loadData(ctx, userID, currency, period)
	if err != nil {
		return freeBalanceResult{}, err
	}

	converter := currencyConverter{target: currency, rates: data.currencyRates}

	accountReserve, err := calculateAccountReserve(data.accounts.Accounts, converter)
	if err != nil {
		return freeBalanceResult{}, err
	}

	plannedReserve, err := calculatePlannedReserve(data.plannedBalance, data.plannedOperations, converter)
	if err != nil {
		return freeBalanceResult{}, err
	}

	budgetReserve, err := calculateBudgetReserve(data.budgets, period.end, now, converter)
	if err != nil {
		return freeBalanceResult{}, err
	}

	return freeBalanceResult{
		totalBalance: data.accounts.TotalBalance,
		availableBalance: data.accounts.TotalBalance.
			Sub(accountReserve).
			Sub(plannedReserve).
			Sub(budgetReserve),
		plannedExpenses: plannedReserve,
		currencyRates:   data.currencyRates,
	}, nil
}

func (s *FreeBalanceService) loadData(
	ctx context.Context,
	userID uuid.UUID,
	currency string,
	period freeBalancePeriod,
) (freeBalanceData, error) {
	var data freeBalanceData
	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		balance, err := s.plannedOperationService.GetBalance(groupCtx, period.end, userID)
		if err == nil {
			data.plannedBalance = balance
		}
		return err
	})

	group.Go(func() error {
		result, err := s.plannedOperationService.List(
			groupCtx,
			userID,
			period.start,
			period.end,
			repositories.ListParams{},
		)
		if err == nil {
			data.plannedOperations = result.PlannedOperations
		}
		return err
	})

	group.Go(func() error {
		result, err := s.accountService.List(groupCtx, userID, repositories.ListParams{}, currency)
		if err == nil {
			data.accounts = *result
		}
		return err
	})

	group.Go(func() error {
		result, err := s.budgetService.List(groupCtx, userID, repositories.ListParams{})
		if err == nil {
			data.budgets = result.Budgets
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
		return freeBalanceData{}, err
	}
	return data, nil
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
	plannedBalance decimal.Decimal,
	operations []dto.PlannedOperationResponse,
	converter currencyConverter,
) (decimal.Decimal, error) {
	reserve := plannedBalance

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
