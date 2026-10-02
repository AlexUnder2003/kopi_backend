package config

import (
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type AppConfig struct {
	Address        string
	DatabaseURL    string
	CurrencyAPIURL string
	SecretKey      string
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPassword   string
	SMTPFrom       string
	Development    bool
}

func NewAppConfig(logger *zap.SugaredLogger) (*AppConfig, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		logger.Errorf("Error loading .env file: %v", err)
		return nil, err
	}

	return &AppConfig{
		Address:        os.Getenv("ADDRESS"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		CurrencyAPIURL: "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1",
		SecretKey:      os.Getenv("SECRET_KEY"),
		SMTPHost:       os.Getenv("SMTP_HOST"),
		SMTPPort:       os.Getenv("SMTP_PORT"),
		SMTPUser:       os.Getenv("SMTP_USER"),
		SMTPPassword:   os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:       os.Getenv("SMTP_FROM"),
		Development:    os.Getenv("DEVELOPMENT") == "true",
	}, nil
}
