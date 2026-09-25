package handlers

import (
	"errors"
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
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req dto.AccountPost
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := req.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.accountService.Create(c.Request().Context(), &models.Account{
		Name:     req.Name,
		Currency: models.CurrencyCode(req.Currency),
		Icon:     req.Icon,
		Balance:  req.Balance,
		UserID:   userID,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create account")
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *AccountHandler) List(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	resp, err := h.accountService.List(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list accounts")
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AccountHandler) GetByID(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id")
	}

	resp, err := h.accountService.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrAccountNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "account not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get account")
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AccountHandler) Update(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id")
	}

	var req dto.AccountUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	resp, err := h.accountService.Update(c.Request().Context(), userID, &models.Account{
		ID:   id,
		Name: req.Name,
		Icon: req.Icon,
	})
	if err != nil {
		if errors.Is(err, services.ErrAccountNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "account not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update account")
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AccountHandler) Delete(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id")
	}

	if err := h.accountService.Delete(c.Request().Context(), id, userID); err != nil {
		if errors.Is(err, services.ErrAccountNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "account not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete account")
	}

	return c.NoContent(http.StatusNoContent)
}
