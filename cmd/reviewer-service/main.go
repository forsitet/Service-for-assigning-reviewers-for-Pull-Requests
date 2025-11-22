package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/config"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/repo/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("Reviewer service started")

	cfg, err := config.ParseConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err.Error())
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := postgres.NewDB(ctx, cfg.DB.ConnString(), logger)
	if err != nil {
		logger.Error("failed to connect to db", "error", err.Error())
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close db", "error", err.Error())
		}
	}()
	if err := postgres.InitTables(ctx, db, logger); err != nil {
		logger.Error("failed to init db schema", "error", err.Error())
		os.Exit(1)
	}

	// TODO: Сделать Грейсфулл Шатдаун
}
