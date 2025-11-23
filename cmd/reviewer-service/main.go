package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apihttp "github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/api/http"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/config"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/repo/postgres"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service"
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
	if err := postgres.RunMigrations(ctx, db, logger); err != nil {
		logger.Error("failed to init db schema", "error", err.Error())
		os.Exit(1)
	}

	teamRepo := postgres.NewTeamRepo(db)
	userRepo := postgres.NewUserRepo(db)
	prRepo := postgres.NewPRRepo(db)

	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo, prRepo)
	prService := service.NewPRService(prRepo, userRepo, time.Now)

	app := service.NewApp(teamService, userService, prService)

	server := apihttp.NewServer(app, logger)
	router := apihttp.NewRouter(server, logger)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("http server starting", "addr", cfg.HTTPAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("shutting down reviewer-service")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown error", "error", err)
	} else {
		logger.Info("http server stopped gracefully")
	}
}
