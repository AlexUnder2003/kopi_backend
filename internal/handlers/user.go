package handlers

import (
	"net/http"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/middleware"
	"KopiBackend/internal/services"
	"KopiBackend/internal/utils"

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
	users.Use(middleware.UserMiddleware)
	users.GET("/me", h.GetMe)
}

func (h *UserHandler) SendOTP(c *echo.Context) error {
	var req dto.SendOTPRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
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
	if err := req.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
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

func (h *UserHandler) GetMe(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	resp, err := h.userService.GetByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user")
	}

	return c.JSON(http.StatusOK, resp)
}
