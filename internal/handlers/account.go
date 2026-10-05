package handlers

import (
	"net/http"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/middleware"
	"KopiBackend/internal/models"
	"KopiBackend/internal/services"
	"KopiBackend/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type AccountHandler struct {
	accountService *services.AccountService
}

func NewAccountHandler(accountService *services.AccountService) *AccountHandler {
	return &AccountHandler{accountService: accountService}
}

func (h *AccountHandler) Register(g *echo.Group) {
	accounts := g.Group("/accounts")
	accounts.Use(middleware.UserMiddleware)

	accounts.POST("", h.Create)
	accounts.GET("", h.List)
	accounts.GET("/:id", h.GetByID)
	accounts.PATCH("/:id", h.Update)
	accounts.DELETE("/:id", h.Delete)
}

func (h *AccountHandler) Create(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	var req dto.AccountPost
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}

	if err := req.Validate(); err != nil {
		return mapAccountValidationErrors(err)
	}

	resp, err := h.accountService.Create(c.Request().Context(), &models.Account{
		Name:                 req.Name,
		Currency:             models.CurrencyCode(req.Currency),
		Icon:                 req.Icon,
		Balance:              req.Balance,
		UserID:               userID,
		IncludeInFreeBalance: &req.IncludeInFreeBalance,
	})
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *AccountHandler) List(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	var currency *string
	if value := c.QueryParam("currency"); value != "" {
		if !dto.ValidCurrency(value) {
			return echo.NewHTTPError(http.StatusBadRequest, "bad_request_currency")
		}
		currency = &value
	}

	if currency == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_currency")
	}

	params, err := listParams(c)
	if err != nil {
		return err
	}

	resp, err := h.accountService.List(c.Request().Context(), userID, params, currency)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AccountHandler) GetByID(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	resp, err := h.accountService.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AccountHandler) Update(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	var req dto.AccountUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}

	resp, err := h.accountService.Update(c.Request().Context(), userID, &models.Account{
		ID:                   id,
		Name:                 req.Name,
		Icon:                 req.Icon,
		IncludeInFreeBalance: req.IncludeInFreeBalance,
	})
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AccountHandler) Delete(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	if err := h.accountService.Delete(c.Request().Context(), id, userID); err != nil {
		return mapAppError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func mapAccountValidationErrors(err error) error {
	if err.Error() == "name" {
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_account_name")
	}
	if err.Error() == "currency" {
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_currency")
	}
	return err
}
