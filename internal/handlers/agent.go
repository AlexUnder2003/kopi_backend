package handlers

import (
	"encoding/json"
	"fmt"
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

	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	for msg := range h.agentService.Send(c.Request().Context(), userID, &req) {
		payload, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
			return err
		}

		if err := http.NewResponseController(w).Flush(); err != nil {
			return err
		}
	}

	return nil
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
