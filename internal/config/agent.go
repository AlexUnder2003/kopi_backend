package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const DefaultAgentInstruction = `Ты финансовый помощник Kopi. Отвечай на языке пользователя.
Помогаешь смотреть счета и категории, а также смотреть, создавать, изменять и удалять операции и запланированные операции.
Не выдумывай id: бери их из list_accounts, list_categories, list_operations и list_planned_operations.
Суммы — положительные десятичные числа. Даты — YYYY-MM-DD или RFC3339.
Перед созданием, изменением или удалением кратко опиши, что именно изменится.`

type AgentConfig struct {
	Name        string
	Instruction string
	Model       string
	APIKey      string
	BaseURL     string
}

func NewAgentConfig(logger *zap.SugaredLogger) (*AgentConfig, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		logger.Errorf("Error loading .env file: %v", err)
		return nil, fmt.Errorf("load .env file: %w", err)
	}

	return &AgentConfig{
		Name:        "kopi",
		Instruction: DefaultAgentInstruction,
		Model:       os.Getenv("OPENAI_MODEL"),
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		BaseURL:     os.Getenv("OPENAI_BASE_URL"),
	}, nil
}
