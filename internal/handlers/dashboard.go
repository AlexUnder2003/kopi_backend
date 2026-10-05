package handlers

import (
	"net/http"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/middleware"
	"KopiBackend/internal/services"
	"KopiBackend/internal/utils"

	"github.com/labstack/echo/v5"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) Register(g *echo.Group) {
	dashboard := g.Group("/dashboard")
	dashboard.Use(middleware.UserMiddleware)

	dashboard.GET("", h.Get)
}

func (h *DashboardHandler) Get(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	currency := c.QueryParam("currency")
	if !dto.ValidCurrency(currency) {
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_currency")
	}

	resp, err := h.dashboardService.GetDashboard(c.Request().Context(), userID, currency)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}
