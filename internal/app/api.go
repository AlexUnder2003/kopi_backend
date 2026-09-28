package app

import (
	"KopiBackend/internal/config"
	"KopiBackend/internal/db"
	"KopiBackend/internal/handlers"
	"KopiBackend/internal/routes"
	"KopiBackend/internal/services"
	"KopiBackend/internal/workers"
	"context"

	"database/sql"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type Worker interface {
	Run(ctx context.Context)
}

type App struct {
	config  *config.AppConfig
	router  *echo.Echo
	db      *sql.DB
	workers []Worker
}

func NewApp() *App {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	sugaredLogger := logger.Sugar()

	cfg, err := config.NewAppConfig(sugaredLogger)
	if err != nil {
		panic(err)
	}

	database, err := db.NewDB(cfg, sugaredLogger)
	if err != nil {
		panic(err)
	}

	userService := services.NewUserService(database, cfg, sugaredLogger)
	accountService := services.NewAccountService(database, sugaredLogger)
	categoryService := services.NewCategoryService(database, sugaredLogger)
	transactionService := services.NewTransactionService(database, sugaredLogger)
	budgetService := services.NewBudgetService(database, categoryService, sugaredLogger)

	hs := []routes.Handler{
		handlers.NewUserHandler(userService),
		handlers.NewAccountHandler(accountService),
		handlers.NewCategoryHandler(categoryService),
		handlers.NewTransactionHandler(transactionService),
		handlers.NewBudgetHandler(budgetService),
	}
	router := routes.Router(hs)
	budgetWorker := workers.NewBudgetWorker(budgetService)
	return &App{config: cfg, router: router, db: database, workers: []Worker{budgetWorker}}
}

func (a *App) Run() {
	addr := a.config.Address
	if addr == "" {
		addr = ":8080"
	}

	for _, worker := range a.workers {
		go worker.Run(context.Background())
	}

	if err := a.router.Start(addr); err != nil {
		panic(err)
	}
}

func (a *App) Stop() {
	if a.db != nil {
		_ = a.db.Close()
	}
}
