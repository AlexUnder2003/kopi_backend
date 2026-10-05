package handlers

import (
	"net/http"
	"strconv"
	"time"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/middleware"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"
	"KopiBackend/internal/services"
	"KopiBackend/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const defaultOperationListLimit = 50

type OperationHandler struct {
	operationService *services.OperationService
}

func NewOperationHandler(operationService *services.OperationService) *OperationHandler {
	return &OperationHandler{operationService: operationService}
}

func (h *OperationHandler) Register(g *echo.Group) {
	operations := g.Group("/operations")
	operations.Use(middleware.UserMiddleware)

	operations.POST("", h.Create)
	operations.GET("", h.List)
	operations.GET("/:id", h.GetByID)
	operations.PATCH("/:id", h.Update)
	operations.DELETE("/:id", h.Delete)
}

func (h *OperationHandler) Create(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	var req dto.OperationPost
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapOperationValidationErrors(err)
	}

	resp, err := h.operationService.Create(c.Request().Context(), userID, &models.Operation{
		Name:           req.Name,
		Amount:         req.Amount,
		Type:           req.Type,
		AccountID:      req.AccountID,
		FromAccountID:  req.FromAccountID,
		CategoryID:     req.CategoryID,
		OccurrenceDate: req.OccurrenceDate,
	})
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *OperationHandler) List(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	params, err := operationListParams(c)
	if err != nil {
		return err
	}

	resp, err := h.operationService.List(c.Request().Context(), userID, params)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *OperationHandler) GetByID(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	resp, err := h.operationService.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *OperationHandler) Update(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	var req dto.OperationUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapOperationValidationErrors(err)
	}

	resp, err := h.operationService.Update(c.Request().Context(), userID, &models.Operation{
		ID:             id,
		Name:           req.Name,
		Amount:         req.Amount,
		Type:           req.Type,
		AccountID:      req.AccountID,
		FromAccountID:  req.FromAccountID,
		CategoryID:     req.CategoryID,
		OccurrenceDate: req.OccurrenceDate,
	})
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *OperationHandler) Delete(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	if err := h.operationService.Delete(c.Request().Context(), id, userID); err != nil {
		return mapAppError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func operationListParams(c *echo.Context) (repositories.OperationListParams, error) {
	var params repositories.OperationListParams

	if value := c.QueryParam("account_id"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			return params, echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
		}
		params.AccountID = id
	}

	if value := c.QueryParam("category_id"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			return params, echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
		}
		params.CategoryID = id
	}

	if value := c.QueryParam("start_date"); value != "" {
		parsed, err := parseOperationDate(value)
		if err != nil {
			return params, echo.NewHTTPError(http.StatusBadRequest, "bad_request_operation_date")
		}
		params.StartDate = parsed
	}

	if value := c.QueryParam("end_date"); value != "" {
		parsed, err := parseOperationDate(value)
		if err != nil {
			return params, echo.NewHTTPError(http.StatusBadRequest, "bad_request_operation_date")
		}
		params.EndDate = parsed
	}

	params.Limit = defaultOperationListLimit
	if value := c.QueryParam("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 1 {
			return params, echo.NewHTTPError(http.StatusBadRequest, "bad_request_operation_limit")
		}
		params.Limit = limit
	}

	if value := c.QueryParam("offset"); value != "" {
		offset, err := strconv.Atoi(value)
		if err != nil || offset < 0 {
			return params, echo.NewHTTPError(http.StatusBadRequest, "bad_request_operation_offset")
		}
		params.Offset = offset
	}

	return params, nil
}

func parseOperationDate(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.DateOnly, value); err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, value)
}

func mapOperationValidationErrors(err error) error {
	switch err.Error() {
	case "name":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_operation_name")
	case "amount":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_operation_amount")
	case "type":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_operation_type")
	case "occurrence_date":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_operation_occurrence_date")
	case "from_account_id":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_operation_from_account_id")
	default:
		return err
	}
}
