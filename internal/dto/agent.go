package dto

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type AgentMessagePost struct {
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message"`
}

type AgentMessageResponse struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

func (r *AgentMessagePost) Validate() error {
	if r.SessionID != "" {
		if _, err := uuid.Parse(r.SessionID); err != nil {
			return errors.New("session_id")
		}
	}
	if strings.TrimSpace(r.Message) == "" {
		return errors.New("message")
	}
	return nil
}
