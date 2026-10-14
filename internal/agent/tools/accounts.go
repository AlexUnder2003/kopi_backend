package tools

import (
	"context"
	"strings"

	agentdto "KopiBackend/internal/agent/dto"
	"KopiBackend/internal/repositories"
	"KopiBackend/internal/services"

	"github.com/google/uuid"
	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type AccountsToolkit struct {
	accountsService *services.AccountService
}

func NewAccountsToolkit(accountsService *services.AccountService) *AccountsToolkit {
	return &AccountsToolkit{accountsService: accountsService}
}

func (t *AccountsToolkit) ListAccounts(ctx context.Context, userID uuid.UUID, params repositories.ListParams, currency string) ([]agentdto.AccountItem, error) {
	listed, err := t.accountsService.List(ctx, userID, params, &currency)
	if err != nil {
		return nil, err
	}

	accounts := make([]agentdto.AccountItem, 0, len(listed.Accounts))
	for _, account := range listed.Accounts {
		accounts = append(accounts, agentdto.AccountItem{
			ID:                   account.ID.String(),
			Name:                 account.Name,
			Currency:             account.Currency,
			Balance:              account.Balance.String(),
			IncludeInFreeBalance: account.IncludeInFreeBalance,
		})
	}
	return accounts, nil
}

func (t *AccountsToolkit) Tools() ([]tool.Tool, error) {
	list, err := functiontool.New(functiontool.Config{
		Name:        "list_accounts",
		Description: "Счета пользователя: id, имя, валюта и баланс. Id счёта нужен для создания операции.",
	}, func(ctx adkagent.Context, args agentdto.ListAccountsArgs) ([]agentdto.AccountItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, err
		}
		currency, err := requiredCurrency(args.Currency)
		if err != nil {
			return nil, err
		}

		return t.ListAccounts(ctx, userID, repositories.ListParams{
			Filter: strings.TrimSpace(args.Filter),
			Limit:  boundedLimit(args.Limit),
		}, currency)
	})

	if err != nil {
		return nil, err
	}

	return []tool.Tool{list}, nil
}
