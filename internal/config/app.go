package config

import (
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type AppConfig struct {
	Address     string
	DatabaseURL string
	SecretKey   string
	SMTPAddress string
}

func NewAppConfig(logger *zap.SugaredLogger) (*AppConfig, error) {
	err := godotenv.Load()

	if err != nil {
		logger.Errorf("Error loading .env file: %v", err)
		return nil, err
	}

	return &AppConfig{
		Address:     os.Getenv("ADDRESS"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		SecretKey:   os.Getenv("SECRET_KEY"),
		SMTPAddress: os.Getenv("SMTP_ADDRESS"),
	}, nil
}
