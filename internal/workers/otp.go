package workers

import (
	"context"
	"time"

	"KopiBackend/internal/services"
)

type OTPWorker struct {
	userService *services.UserService
}

func NewOTPWorker(userService *services.UserService) *OTPWorker {
	return &OTPWorker{userService: userService}
}

func (w *OTPWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		w.userService.DeleteExpiredOTPs(ctx)
	}
}
