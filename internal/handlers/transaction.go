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

const defaultTransactionListLimit = 50

type TransactionHandler struct {
	transactionService *services.TransactionService
}

func NewTransactionHandler(transactionService *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionService: transactionService}
}

func (h *TransactionHandler) Register(g *echo.Group) {
	transactions := g.Group("/transactions")
	transactions.Use(middleware.UserMiddleware)

	transactions.POST("", h.Create)
	transactions.GET("", h.List)
	transactions.GET("/:id", h.GetByID)
	transactions.PATCH("/:id", h.Update)
	transactions.DELETE("/:id", h.Delete)
}

func (h *TransactionHandler) Create(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	var req dto.TransactionPost
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapTransactionValidationErrors(err)
	}

	resp, err := h.transactionService.Create(c.Request().Context(), userID, &models.Transaction{
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

func (h *TransactionHandler) List(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	params, err := transactionListParams(c)
	if err != nil {
		return err
	}

	resp, err := h.transactionService.List(c.Request().Context(), userID, params)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *TransactionHandler) GetByID(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	resp, err := h.transactionService.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *TransactionHandler) Update(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	var req dto.TransactionUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapTransactionValidationErrors(err)
	}

	resp, err := h.transactionService.Update(c.Request().Context(), userID, &models.Transaction{
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

func (h *TransactionHandler) Delete(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	if err := h.transactionService.Delete(c.Request().Context(), id, userID); err != nil {
		return mapAppError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func transactionListParams(c *echo.Context) (repositories.TransactionListParams, error) {
	var params repositories.TransactionListParams

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
		parsed, err := parseTransactionDate(value)
		if err != nil {
			return params, echo.NewHTTPError(http.StatusBadRequest, "bad_request_transaction_date")
		}
		params.StartDate = parsed
	}

	if value := c.QueryParam("end_date"); value != "" {
		parsed, err := parseTransactionDate(value)
		if err != nil {
			return params, echo.NewHTTPError(http.StatusBadRequest, "bad_request_transaction_date")
		}
		params.EndDate = parsed
	}

	params.Limit = defaultTransactionListLimit
	if value := c.QueryParam("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 1 {
			return params, echo.NewHTTPError(http.StatusBadRequest, "bad_request_transaction_limit")
		}
		params.Limit = limit
	}

	if value := c.QueryParam("offset"); value != "" {
		offset, err := strconv.Atoi(value)
		if err != nil || offset < 0 {
			return params, echo.NewHTTPError(http.StatusBadRequest, "bad_request_transaction_offset")
		}
		params.Offset = offset
	}

	return params, nil
}

func parseTransactionDate(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.DateOnly, value); err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, value)
}

func mapTransactionValidationErrors(err error) error {
	switch err.Error() {
	case "name":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_transaction_name")
	case "amount":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_transaction_amount")
	case "type":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_transaction_type")
	case "occurrence_date":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_transaction_occurrence_date")
	case "from_account_id":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_transaction_from_account_id")
	default:
		return err
	}
}
