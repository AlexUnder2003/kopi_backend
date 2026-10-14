package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"KopiBackend/internal/agent"
	"KopiBackend/internal/apperrors"
	"KopiBackend/internal/config"
	"KopiBackend/internal/dto"
	"KopiBackend/internal/models"
	"KopiBackend/internal/repositories"

	"github.com/google/uuid"
	"go.uber.org/zap"
	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/genai"
)

const (
	agentAppName             = "kopi"
	errMsgForbidden          = "agent_forbidden"
	errMsgAgentQuotaExceeded = "agent_quota_exceeded"
	defaultAIExecutionQuota  = 15
)

type AgentService struct {
	runner              *runner.Runner
	subscriptionService *SubscriptionService
	aiExecutionRepo     *repositories.AIExecutionRepository
	logger              *zap.SugaredLogger
}

func NewAgentService(ctx context.Context, db *sql.DB, cfg config.AgentConfig, logger *zap.SugaredLogger, tools []tool.Tool, subscriptionService *SubscriptionService) (*AgentService, error) {
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
		runner:              adkRunner,
		subscriptionService: subscriptionService,
		aiExecutionRepo:     repositories.NewAIExecutionRepository(db),
		logger:              logger,
	}, nil
}

func (s *AgentService) Send(ctx context.Context, userID uuid.UUID, req *dto.AgentMessagePost) (<-chan dto.AgentMessageResponse, error) {
	if err := s.ensureQuota(ctx, userID); err != nil {
		return nil, err
	}

	out := make(chan dto.AgentMessageResponse)
	go func() {
		defer close(out)

		sessionID := req.SessionID
		if sessionID == "" {
			sessionID = uuid.NewString()
		}

		used := false

		for event, err := range s.runner.Run(ctx, userID.String(), sessionID, genai.NewContentFromText(req.Message, genai.RoleUser), adkagent.RunConfig{
			StreamingMode: adkagent.StreamingModeSSE,
		}) {
			if err != nil {
				s.logger.Errorw("agent run failed", "error", err, "user_id", userID)
				break
			}

			text := modelText(event)
			if text == "" || !event.Partial {
				continue
			}

			used = true

			select {
			case <-ctx.Done():
			case out <- dto.AgentMessageResponse{SessionID: sessionID, Message: text}:
			}
			if ctx.Err() != nil {
				break
			}
		}

		if !used {
			return
		}

		if err := s.consumeQuota(context.WithoutCancel(ctx), userID); err != nil {
			s.logger.Errorw("failed to consume ai quota", "error", err, "user_id", userID)
		}
	}()

	return out, nil
}

func (s *AgentService) ensureQuota(ctx context.Context, userID uuid.UUID) error {
	execution, err := s.aiExecutionRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		s.logger.Errorw("failed to get ai execution", "error", err, "user_id", userID)
		return apperrors.Internal(errInternalServerError)
	}

	if execution.LeftQuota <= 0 {
		return apperrors.Forbidden(errMsgAgentQuotaExceeded)
	}

	return nil
}

func (s *AgentService) consumeQuota(ctx context.Context, userID uuid.UUID) error {
	execution, err := s.aiExecutionRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Errorw("failed to get ai execution", "error", err, "user_id", userID)
		return apperrors.Internal(errInternalServerError)
	}

	execution.LeftQuota--
	if execution.LeftQuota < 0 {
		nextResetAt := time.Now().UTC().AddDate(0, 1, 0)

		_, err = s.aiExecutionRepo.Update(ctx, &models.AIExecution{
			UserID:      userID,
			LeftQuota:   0,
			NextResetAt: nextResetAt,
		})

		if err != nil {
			s.logger.Errorw("failed to update ai execution", "error", err, "user_id", userID)
			return apperrors.Internal(errInternalServerError)
		}

		return apperrors.Forbidden(errMsgAgentQuotaExceeded)
	}

	if _, err := s.aiExecutionRepo.Update(ctx, execution); err != nil {
		s.logger.Errorw("failed to update ai execution", "error", err, "user_id", userID)
		return apperrors.Internal(errInternalServerError)
	}

	return nil
}

func modelText(event *session.Event) string {
	if event == nil || event.Content == nil || event.Content.Role != string(genai.RoleModel) {
		return ""
	}
	var text strings.Builder
	for _, part := range event.Content.Parts {
		if part != nil && part.Text != "" {
			text.WriteString(part.Text)
		}
	}
	return text.String()
}
