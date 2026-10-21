package workers

import (
	"context"
	"time"

	"KopiBackend/internal/services"
)

type AIExecutionWorker struct {
	aiExecutionService *services.AIExecutionService
}

func NewAIExecutionWorker(aiExecutionService *services.AIExecutionService) *AIExecutionWorker {
	return &AIExecutionWorker{aiExecutionService: aiExecutionService}
}

func (w *AIExecutionWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		w.aiExecutionService.ResetQuota(ctx)
	}
}
