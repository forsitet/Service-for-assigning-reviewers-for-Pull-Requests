package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/migrations"
)

func RunMigrations(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	logger.Info("running DB migrations with goose")

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	logger.Info("DB migrations applied successfully")
	return nil
}
