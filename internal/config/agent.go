package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const DefaultAgentInstruction = `Ты финансовый помощник Kopi. Отвечай на языке пользователя.
Помогаешь смотреть счета, категории, операции и запланированные операции, а также создавать доходы, расходы, переводы и планы.
Не выдумывай id счетов и категорий: бери их из list_accounts и list_categories.
Суммы — положительные десятичные числа. Даты — YYYY-MM-DD или RFC3339.
Перед созданием операции или плана кратко опиши, что именно будет создано.`

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
