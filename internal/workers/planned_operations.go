package workers

import (
	"KopiBackend/internal/services"
	"context"
	"time"
)

type PlannedOperationsWorker struct {
	plannedOperationService *services.PlannedOperationService
}

func NewPlannedOperationsWorker(plannedOperationService *services.PlannedOperationService) *PlannedOperationsWorker {
	return &PlannedOperationsWorker{plannedOperationService: plannedOperationService}
}

func (w *PlannedOperationsWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		w.plannedOperationService.Execute(ctx)
	}
}
