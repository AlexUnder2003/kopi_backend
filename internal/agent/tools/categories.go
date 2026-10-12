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

type CategoriesToolkit struct {
	categoriesService *services.CategoryService
}

func NewCategoriesToolkit(categoriesService *services.CategoryService) *CategoriesToolkit {
	return &CategoriesToolkit{categoriesService: categoriesService}
}

func (t *CategoriesToolkit) ListCategories(ctx context.Context, userID uuid.UUID, params repositories.ListParams) ([]agentdto.CategoryItem, error) {
	categories, err := t.categoriesService.List(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	items := make([]agentdto.CategoryItem, 0, len(categories))
	for _, category := range categories {
		items = append(items, agentdto.CategoryItem{
			ID:       category.ID.String(),
			Name:     category.Name,
			IsSystem: category.IsSystem,
		})
	}
	return items, nil
}

func (t *CategoriesToolkit) Tools() ([]tool.Tool, error) {
	list, err := functiontool.New(functiontool.Config{
		Name:        "list_categories",
		Description: "Категории пользователя и системные категории. Id категории нужен для создания операции.",
	}, func(ctx adkagent.Context, args agentdto.ListCategoriesArgs) ([]agentdto.CategoryItem, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, err
		}

		return t.ListCategories(ctx, userID, repositories.ListParams{
			Filter: strings.TrimSpace(args.Filter),
			Limit:  boundedLimit(args.Limit),
		})
	})
	if err != nil {
		return nil, err
	}
	return []tool.Tool{list}, nil
}
