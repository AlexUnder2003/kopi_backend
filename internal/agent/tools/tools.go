package tools

import (
	"KopiBackend/internal/services"

	"google.golang.org/adk/v2/tool"
)

func Collect(
	accounts *services.AccountService,
	categories *services.CategoryService,
	operations *services.OperationService,
	planned *services.PlannedOperationService,
) ([]tool.Tool, error) {
	groups := []interface {
		Tools() ([]tool.Tool, error)
	}{
		NewAccountsToolkit(accounts),
		NewCategoriesToolkit(categories),
		NewOperationsToolkit(operations),
		NewPlannedOperationsToolkit(planned),
	}

	collected := make([]tool.Tool, 0, len(groups))
	for _, group := range groups {
		items, err := group.Tools()
		if err != nil {
			return nil, err
		}
		collected = append(collected, items...)
	}
	return collected, nil
}
