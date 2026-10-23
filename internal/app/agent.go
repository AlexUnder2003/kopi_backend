package app

import (
	"context"
	"database/sql"

	agenttools "KopiBackend/internal/agent/tools"
	"KopiBackend/internal/config"
	"KopiBackend/internal/db"
	"KopiBackend/internal/handlers"
	"KopiBackend/internal/routes"
	"KopiBackend/internal/services"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type Agent struct {
	config *config.AppConfig
	router *echo.Echo
	db     *sql.DB
}

func NewAgent() *Agent {
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

	accountService := services.NewAccountService(database, cfg, sugaredLogger)
	categoryService := services.NewCategoryService(database, sugaredLogger)
	budgetService := services.NewBudgetService(database, categoryService, sugaredLogger)
	plannedOperationService := services.NewPlannedOperationService(database, accountService, sugaredLogger)
	goalService := services.NewGoalService(database, sugaredLogger)
	operationService := services.NewOperationService(database, accountService, budgetService, plannedOperationService, goalService, categoryService, sugaredLogger)
	subscriptionService := services.NewSubscriptionService(database, sugaredLogger)

	toolset, err := agenttools.Collect(accountService, categoryService, operationService, plannedOperationService, goalService)
	if err != nil {
		panic(err)
	}

	agentCfg, err := config.NewAgentConfig(sugaredLogger)
	if err != nil {
		panic(err)
	}

	agentService, err := services.NewAgentService(context.Background(), database, *agentCfg, sugaredLogger, toolset, subscriptionService)
	if err != nil {
		panic(err)
	}

	router := routes.Router([]routes.Handler{
		handlers.NewAgentHandler(agentService),
	})

	return &Agent{config: cfg, router: router, db: database}
}

func (a *Agent) Run() {
	addr := a.config.Address
	if addr == "" {
		addr = ":8081"
	}

	if err := a.router.Start(addr); err != nil {
		panic(err)
	}
}

func (a *Agent) Stop() {
	if a.db != nil {
		_ = a.db.Close()
	}
}
