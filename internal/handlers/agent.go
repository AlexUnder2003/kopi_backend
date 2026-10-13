package handlers

import (
	"net/http"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/middleware"
	"KopiBackend/internal/services"
	"KopiBackend/internal/utils"

	"github.com/labstack/echo/v5"
)

type AgentHandler struct {
	agentService *services.AgentService
}

func NewAgentHandler(agentService *services.AgentService) *AgentHandler {
	return &AgentHandler{agentService: agentService}
}

func (h *AgentHandler) Register(g *echo.Group) {
	agentGroup := g.Group("/agent")
	agentGroup.Use(middleware.UserMiddleware)

	agentGroup.POST("/messages", h.Send)
}

func (h *AgentHandler) Send(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	var req dto.AgentMessagePost
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}

	if err := req.Validate(); err != nil {
		return mapAgentValidationErrors(err)
	}

	resp, err := h.agentService.Send(c.Request().Context(), userID, &req)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func mapAgentValidationErrors(err error) error {
	switch err.Error() {
	case "message":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_agent_message")
	case "session_id":
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_agent_session_id")
	default:
		return err
	}
}
