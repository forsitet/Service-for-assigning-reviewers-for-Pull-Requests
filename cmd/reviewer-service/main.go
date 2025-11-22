package main

import (
	"log/slog"
	"os"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("Reviewer service started")

	cfg, err := config.ParseConfig()

	if err != nil {
		logger.Error("Failed to load config", "error", err.Error())
		os.Exit(1)
	}
	_ = cfg

	// TODO: Сделать Грейсфулл Шатдаун
}
