package workers

import (
	"context"
	"time"

	"KopiBackend/internal/services"
)

type UserTokenWorker struct {
	userService *services.UserService
}

func NewUserTokenWorker(userService *services.UserService) *UserTokenWorker {
	return &UserTokenWorker{userService: userService}
}

func (w *UserTokenWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		w.userService.DeleteExpiredTokens(ctx)
	}
}
