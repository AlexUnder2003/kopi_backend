package workers

import (
	"KopiBackend/internal/services"
	"context"
	"time"
)

type BudgetWorker struct {
	budgetService *services.BudgetService
}

func NewBudgetWorker(budgetService *services.BudgetService) *BudgetWorker {
	return &BudgetWorker{budgetService: budgetService}
}

func (w *BudgetWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		w.budgetService.Reset(ctx)
	}
}
