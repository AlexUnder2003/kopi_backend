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

type GoalHandler struct {
	goalService *services.GoalService
}

func NewGoalHandler(goalService *services.GoalService) *GoalHandler {
	return &GoalHandler{goalService: goalService}
}

func (h *GoalHandler) Register(g *echo.Group) {
	goals := g.Group("/goals")
	goals.Use(middleware.UserMiddleware)

	goals.POST("", h.Create)
	goals.GET("", h.List)
	goals.GET("/:id", h.GetByID)
	goals.PATCH("/:id", h.Update)
	goals.DELETE("/:id", h.Delete)
}

func (h *GoalHandler) Create(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	var req dto.GoalPost
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapGoalValidationErrors(err)
	}

	resp, err := h.goalService.Create(c.Request().Context(), userID, &req)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *GoalHandler) List(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	params, err := listParams(c)
	if err != nil {
		return err
	}

	resp, err := h.goalService.List(c.Request().Context(), userID, params)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *GoalHandler) GetByID(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	resp, err := h.goalService.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *GoalHandler) Update(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	var req dto.GoalUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapGoalValidationErrors(err)
	}

	resp, err := h.goalService.Update(c.Request().Context(), id, userID, &req)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *GoalHandler) Delete(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	if err := h.goalService.Delete(c.Request().Context(), id, userID); err != nil {
		return mapAppError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func mapGoalValidationErrors(err error) error {
	switch err.Error() {
	case "name":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_goal_name")
	case "currency":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_goal_currency")
	case "target_amount":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_goal_target_amount")
	case "current_amount":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_goal_current_amount")
	default:
		return err
	}
}
