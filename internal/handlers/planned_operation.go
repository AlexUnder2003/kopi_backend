package handlers

import (
	"net/http"
	"time"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/middleware"
	"KopiBackend/internal/services"
	"KopiBackend/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type PlannedOperationHandler struct {
	plannedOperationService *services.PlannedOperationService
}

func NewPlannedOperationHandler(plannedOperationService *services.PlannedOperationService) *PlannedOperationHandler {
	return &PlannedOperationHandler{plannedOperationService: plannedOperationService}
}

func (h *PlannedOperationHandler) Register(g *echo.Group) {
	plannedOperations := g.Group("/planned-operations")
	plannedOperations.Use(middleware.UserMiddleware)

	plannedOperations.POST("", h.Create)
	plannedOperations.GET("", h.List)
	plannedOperations.GET("/:id", h.GetByID)
	plannedOperations.PATCH("/:id", h.Update)
	plannedOperations.DELETE("/:id", h.Delete)
}

func (h *PlannedOperationHandler) Create(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	var req dto.PlannedOperationPost
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapPlannedOperationValidationErrors(err)
	}

	resp, err := h.plannedOperationService.Create(c.Request().Context(), userID, &req)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *PlannedOperationHandler) List(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	resp, err := h.plannedOperationService.List(c.Request().Context(), userID, time.Time{}, time.Time{})
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PlannedOperationHandler) GetByID(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	resp, err := h.plannedOperationService.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PlannedOperationHandler) Update(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	var req dto.PlannedOperationUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapPlannedOperationValidationErrors(err)
	}

	resp, err := h.plannedOperationService.Update(c.Request().Context(), id, userID, &req)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PlannedOperationHandler) Delete(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	if err := h.plannedOperationService.Delete(c.Request().Context(), id, userID); err != nil {
		return mapAppError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func mapPlannedOperationValidationErrors(err error) error {
	switch err.Error() {
	case "name":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_planned_operation_name")
	case "amount":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_planned_operation_amount")
	case "type":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_planned_operation_type")
	case "interval_type":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_planned_operation_interval_type")
	case "interval":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_planned_operation_interval")
	case "account_id":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_planned_operation_account_id")
	case "category_id":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_planned_operation_category_id")
	case "planned_date":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_planned_operation_planned_date")
	default:
		return err
	}
}
