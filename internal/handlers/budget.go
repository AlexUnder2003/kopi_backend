package handlers

import (
	"net/http"
	"strings"
	"time"

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
	budgets.GET("/already-spent", h.GetAlreadySpent)
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

	params, err := listParams(c)
	if err != nil {
		return err
	}

	resp, err := h.budgetService.List(c.Request().Context(), userID, params)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BudgetHandler) GetAlreadySpent(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	categoryIDs := strings.Split(c.QueryParam("category_ids"), ",")
	if len(categoryIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_category_ids")
	}

	startDate, err := time.Parse(time.RFC3339, c.QueryParam("start_date"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_start_date")
	}

	endDate, err := time.Parse(time.RFC3339, c.QueryParam("end_date"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_end_date")
	}

	currency := c.QueryParam("currency")
	if !dto.ValidCurrency(currency) {
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_currency")
	}

	if endDate.Before(startDate) {
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_end_date")
	}

	categoryIDsUUID := make([]uuid.UUID, len(categoryIDs))
	for i, categoryID := range categoryIDs {
		categoryIDsUUID[i], err = uuid.Parse(categoryID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "bad_request_budget_category_ids")
		}
	}

	alreadySpent, err := h.budgetService.GetAlreadySpent(c.Request().Context(), userID, categoryIDsUUID, startDate, endDate, currency)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, dto.AlreadySpentResponse{AlreadySpent: alreadySpent})
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
