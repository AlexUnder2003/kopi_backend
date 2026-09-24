package db

import (
	"KopiBackend/internal/config"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func NewDB(config *config.AppConfig, logger *zap.SugaredLogger) (*sql.DB, error) {
	db, err := sql.Open("pgx", config.DatabaseURL)
	if err != nil {
		logger.Errorf("Failed to open database: %v", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Errorf("Failed to ping database: %v", err)
		return nil, err
	}

	if err := RunMigrations(db, logger); err != nil {
		logger.Errorf("Failed to run migrations: %v", err)
		return nil, err
	}

	return db, nil
}

func RunMigrations(db *sql.DB, logger *zap.SugaredLogger) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Errorf("Failed to create migration driver: %v", err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/migrations",
		"postgres",
		driver,
	)

	if err != nil {
		logger.Errorf("Failed to create migration instance: %v", err)
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Errorf("Failed to run migrations: %v", err)
		return err
	}

	return nil
}
