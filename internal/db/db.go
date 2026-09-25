package db

import (
	"KopiBackend/internal/config"
	"KopiBackend/internal/migrations"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		logger.Errorf("Failed to create migration source: %v", err)
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
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
