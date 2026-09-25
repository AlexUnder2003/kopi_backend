package handlers

import (
	"net/http"
	"strings"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(g *echo.Group) {
	auth := g.Group("/auth")
	auth.POST("/otp", h.SendOTP)
	auth.POST("/login", h.Login)

	users := g.Group("/users")
	users.GET("/:id", h.GetByID)
}

func (h *UserHandler) SendOTP(c *echo.Context) error {
	var req dto.SendOTPRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Email) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	if err := h.userService.SendOTP(c.Request().Context(), req.Email); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send otp")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *UserHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.OTP) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "otp is required")
	}

	resp, err := h.userService.Login(c.Request().Context(), req.OTP)
	if err != nil {
		if err.Error() == "invalid OTP" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid otp")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to login")
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetByID(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	resp, err := h.userService.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user")
	}

	return c.JSON(http.StatusOK, resp)
}
