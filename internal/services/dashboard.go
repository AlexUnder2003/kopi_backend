package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"KopiBackend/internal/config"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var errCurrencyRateNotFound = errors.New("currency_rate_not_found")

type DashboardService struct {
	config                  *config.AppConfig
	accountService          *AccountService
	transactionService      *TransactionService
	budgetService           *BudgetService
	plannedOperationService *PlannedOperationService
}

func NewDashboardService(
	config *config.AppConfig,
	accountService *AccountService,
	transactionService *TransactionService,
	budgetService *BudgetService,
	plannedOperationService *PlannedOperationService,
) *DashboardService {
	return &DashboardService{
		config:                  config,
		accountService:          accountService,
		transactionService:      transactionService,
		budgetService:           budgetService,
		plannedOperationService: plannedOperationService,
	}
}

func (s *DashboardService) getCurrencyRates(base string) (map[string]decimal.Decimal, error) {
	url := s.config.CurrencyAPIURL + "/currencies/" + strings.ToLower(base) + ".json"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var currencyResponse dto.CurrencyResponse
	if err := json.Unmarshal(body, &currencyResponse); err != nil {
		return nil, err
	}

	rates := make(map[string]decimal.Decimal, len(currencyResponse.Rates))
	for code, rate := range currencyResponse.Rates {
		rates[code] = decimal.NewFromFloat(rate)
	}
	return rates, nil
}

func (s *DashboardService) GetDashboard(ctx context.Context, userID uuid.UUID, currency string) (*dto.Dashboard, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, -1)

	var (
		accounts          []dto.AccountResponse
		transactions      []dto.TransactionResponse
		budgets           []dto.BudgetResponse
		plannedOperations []dto.PlannedOperationResponse

		totalBalance     decimal.Decimal
		availableBalance decimal.Decimal
		totalIncome      decimal.Decimal
		totalExpenses    decimal.Decimal

		fetchErr      error
		currencyRates map[string]decimal.Decimal
		mu            sync.Mutex
	)

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(5)

	go func() {
		defer waitGroup.Done()
		result, err := s.plannedOperationService.List(ctx, userID, monthStart, monthEnd)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			fetchErr = err
			return
		}
		plannedOperations = result
	}()

	go func() {
		defer waitGroup.Done()
		result, err := s.accountService.List(ctx, userID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			fetchErr = err
			return
		}
		accounts = result
	}()

	go func() {
		defer waitGroup.Done()
		result, err := s.transactionService.List(ctx, userID, repositories.TransactionListParams{
			StartDate: monthStart,
			EndDate:   monthEnd,
		})
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			fetchErr = err
			return
		}
		for _, transaction := range result {
			transactions = append(transactions, transaction.(dto.TransactionResponse))
		}
	}()

	go func() {
		defer waitGroup.Done()
		result, err := s.budgetService.List(ctx, userID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			fetchErr = err
			return
		}
		budgets = result
	}()

	go func() {
		defer waitGroup.Done()
		rates, err := s.getCurrencyRates(currency)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			fetchErr = err
			return
		}
		currencyRates = rates
	}()

	waitGroup.Wait()
	if fetchErr != nil {
		return nil, fetchErr
	}

	for _, account := range accounts {
		if account.Currency == currency {
			totalBalance = totalBalance.Add(account.Balance)
			continue
		}

		rate, ok := currencyRates[strings.ToLower(account.Currency)]
		if !ok || rate.IsZero() {
			return nil, errCurrencyRateNotFound
		}
		totalBalance = totalBalance.Add(account.Balance.Div(rate))
	}

	for _, transaction := range transactions {
		if transaction.Account.Currency != currency {
			rate, ok := currencyRates[strings.ToLower(transaction.Account.Currency)]
			if !ok || rate.IsZero() {
				return nil, errCurrencyRateNotFound
			}
			transaction.Amount = transaction.Amount.Div(rate)
		}

		switch transaction.Type {
		case models.TransactionTypeIncome:
			totalIncome = totalIncome.Add(transaction.Amount)

		case models.TransactionTypeExpense:
			totalExpenses = totalExpenses.Add(transaction.Amount)
		}
	}

	availableBalance = totalBalance

	for _, plannedOperation := range plannedOperations {
		if plannedOperation.Account.Currency != currency {
			rate, ok := currencyRates[strings.ToLower(plannedOperation.Account.Currency)]
			if !ok || rate.IsZero() {
				return nil, errCurrencyRateNotFound
			}

			plannedOperation.Amount = plannedOperation.Amount.Div(rate)
		}

		if plannedOperation.Type == models.TransactionTypeExpense {
			availableBalance = availableBalance.Sub(plannedOperation.Amount)
		}
	}

	for _, budget := range budgets {
		if budget.Currency == currency {
			availableBalance = availableBalance.Sub(budget.Balance)
		} else {
			rate, ok := currencyRates[strings.ToLower(budget.Currency)]
			if !ok || rate.IsZero() {
				return nil, errCurrencyRateNotFound
			}
			availableBalance = availableBalance.Sub(budget.Balance.Div(rate))
		}
	}

	return &dto.Dashboard{
		TotalBalance:     totalBalance,
		AvailableBalance: availableBalance,
		TotalIncome:      totalIncome,
		TotalExpenses:    totalExpenses,
	}, nil
}
