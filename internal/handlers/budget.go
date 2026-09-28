package handlers

import (
	"net/http"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/middleware"
	"KopiBackend/internal/services"
	"KopiBackend/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type BudgetHandler struct {
	budgetService *services.BudgetService
}

func NewBudgetHandler(budgetService *services.BudgetService) *BudgetHandler {
	return &BudgetHandler{budgetService: budgetService}
}

func (h *BudgetHandler) Register(g *echo.Group) {
	budgets := g.Group("/budgets")
	budgets.Use(middleware.UserMiddleware)

	budgets.POST("", h.Create)
	budgets.GET("", h.List)
	budgets.GET("/:id", h.GetByID)
	budgets.PATCH("/:id", h.Update)
	budgets.DELETE("/:id", h.Delete)
}

func (h *BudgetHandler) Create(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	var req dto.BudgetPost
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapBudgetValidationErrors(err)
	}

	resp, err := h.budgetService.Create(c.Request().Context(), userID, &req)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *BudgetHandler) List(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	resp, err := h.budgetService.List(c.Request().Context(), userID)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BudgetHandler) GetByID(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	resp, err := h.budgetService.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BudgetHandler) Update(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	var req dto.BudgetUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapBudgetValidationErrors(err)
	}

	resp, err := h.budgetService.Update(c.Request().Context(), id, userID, &req)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BudgetHandler) Delete(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	if err := h.budgetService.Delete(c.Request().Context(), id, userID); err != nil {
		return mapAppError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func mapBudgetValidationErrors(err error) error {
	switch err.Error() {
	case "name":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_name")
	case "amount":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_amount")
	case "interval_type":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_interval_type")
	case "interval":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_interval")
	case "start_date":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_start_date")
	case "category_id":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_category_id")
	case "currency":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_currency")
	default:
		return err
	}
}
