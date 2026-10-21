package workers

import (
	"context"
	"time"

	"KopiBackend/internal/services"
)

type SubscriptionWorker struct {
	subscriptionService *services.SubscriptionService
}

func NewSubscriptionWorker(subscriptionService *services.SubscriptionService) *SubscriptionWorker {
	return &SubscriptionWorker{subscriptionService: subscriptionService}
}

func (w *SubscriptionWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		w.subscriptionService.ResetSubscription(ctx)
	}
}
