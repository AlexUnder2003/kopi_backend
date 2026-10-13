package services

import (
	"context"
	"strings"

	"KopiBackend/internal/agent"
	"KopiBackend/internal/apperrors"
	"KopiBackend/internal/config"
	"KopiBackend/internal/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"
	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/genai"
)

const agentAppName = "kopi"

type AgentService struct {
	runner *runner.Runner
	logger *zap.SugaredLogger
}

func NewAgentService(ctx context.Context, cfg config.AgentConfig, logger *zap.SugaredLogger, tools []tool.Tool) (*AgentService, error) {
	if cfg.Name == "" {
		cfg.Name = agentAppName
	}
	if cfg.Instruction == "" {
		cfg.Instruction = config.DefaultAgentInstruction
	}

	sessions := session.InMemoryService()
	created, err := agent.NewAgent(ctx, cfg, tools)
	if err != nil {
		return nil, err
	}

	adkRunner, err := runner.New(runner.Config{
		AppName:           cfg.Name,
		Agent:             created,
		SessionService:    sessions,
		AutoCreateSession: true,
	})
	if err != nil {
		return nil, err
	}

	return &AgentService{
		runner: adkRunner,
		logger: logger,
	}, nil
}

func (s *AgentService) Send(ctx context.Context, userID uuid.UUID, req *dto.AgentMessagePost) (*dto.AgentMessageResponse, error) {
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.NewString()
	}

	response := &dto.AgentMessageResponse{SessionID: sessionID}

	for event, err := range s.runner.Run(ctx, userID.String(), sessionID, genai.NewContentFromText(req.Message, genai.RoleUser), adkagent.RunConfig{
		StreamingMode: adkagent.StreamingModeNone,
	}) {
		if err != nil {
			s.logger.Errorw("agent run failed", "error", err, "user_id", userID)
			return nil, apperrors.Internal(errInternalServerError)
		}

		if event == nil || event.Partial || event.Content == nil || event.Content.Role != string(genai.RoleModel) {
			continue
		}

		var text strings.Builder
		for _, part := range event.Content.Parts {
			if part == nil || part.Text == "" {
				continue
			}
			text.WriteString(part.Text)
		}
		if text.Len() > 0 {
			response.Message = text.String()
		}
	}

	return response, nil
}
