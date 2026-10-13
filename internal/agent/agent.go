package agent

import (
	"context"
	"fmt"

	"KopiBackend/internal/config"

	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/openaimodel"
	"google.golang.org/adk/v2/tool"
)

func NewAgent(ctx context.Context, cfg config.AgentConfig, tools []tool.Tool) (adkagent.Agent, error) {
	llm, err := openaimodel.NewModel(ctx, cfg.Model, &openaimodel.ClientConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		API:     openaimodel.APIChatCompletions,
	})
	if err != nil {
		return nil, fmt.Errorf("create openai model: %w", err)
	}

	created, err := llmagent.New(llmagent.Config{
		Name:        cfg.Name,
		Instruction: cfg.Instruction,
		Model:       llm,
		Tools:       tools,
	})
	if err != nil {
		return nil, fmt.Errorf("create llm agent: %w", err)
	}

	return created, nil
}
