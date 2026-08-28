package main

import (
	"context"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/company/hrbot/internal/config"
	"github.com/company/hrbot/internal/delivery/http"
	"github.com/company/hrbot/internal/delivery/telegram"
	"gopkg.in/telebot.v3"
	"github.com/company/hrbot/internal/domain/admin"
	"github.com/company/hrbot/internal/domain/application"
	"github.com/company/hrbot/internal/domain/bottext"
	"github.com/company/hrbot/internal/domain/resume"
	"github.com/company/hrbot/internal/domain/user"
	"github.com/company/hrbot/internal/domain/vacancy"
	"github.com/company/hrbot/internal/repository/postgres"
	"github.com/company/hrbot/internal/worker"
	"github.com/hibiken/asynq"
	"github.com/company/hrbot/internal/repository/redis"
	"github.com/company/hrbot/pkg/i18n"
	"github.com/company/hrbot/pkg/logger"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		// Can't use slog yet if it needs config, but default logger works
		panic("failed to load config: " + err.Error())
	}

	// Init logger
	logger.Init(cfg.LogLevel)
	slog.Info("Starting HR Telegram Bot...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to PostgreSQL
	pgPool, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pgPool.Close()
	slog.Info("Connected to PostgreSQL")

	// Connect to Redis
	redisClient, err := redis.New(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("Failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	slog.Info("Connected to Redis")

	// Init Repositories & State
	adminRepo := admin.NewRepository(pgPool)
	userRepo := user.NewRepository(pgPool)
	resumeRepo := resume.NewRepository(pgPool)
	vacancyRepo := vacancy.NewRepository(pgPool)
	appRepo := application.NewRepository(pgPool)
	textRepo := bottext.NewRepository(pgPool)
	stateManager := telegram.NewStateManager(redisClient)

	// Init i18n
	translator, err := i18n.NewTranslator("internal/locales")
	if err != nil {
		slog.Error("Failed to initialize translator", "error", err)
		os.Exit(1)
	}

	// Setup Telegram Bot
	botClient, err := telegram.NewBot(cfg, userRepo, resumeRepo, vacancyRepo, appRepo, textRepo, stateManager, translator)
	if err != nil {
		slog.Error("Failed to initialize telegram bot", "error", err)
		os.Exit(1)
	}
	botClient.SetupHandlers()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Setup Asynq
	redisOpt, _ := asynq.ParseRedisURI(cfg.RedisURL)
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	// Initialize telebot without poller for worker
	tbot, err := telebot.NewBot(telebot.Settings{
		Token: cfg.TelegramBotToken,
	})
	if err != nil {
		slog.Error("Failed to init telebot for worker", "error", err)
		os.Exit(1)
	}

	broadcastProcessor := worker.NewBroadcastProcessor(userRepo, tbot)
	workerServer := worker.NewWorkerServer(cfg.RedisURL, broadcastProcessor)
	go func() {
		if err := workerServer.Start(); err != nil {
			slog.Error("Failed to start worker server", "error", err)
		}
	}()
	defer workerServer.Stop()

	// Setup HTTP Server
	router := http.SetupRouter(cfg.JWTSecret, asynqClient, adminRepo, vacancyRepo, appRepo)
	port := cfg.AppPort
	if port == "" {
		port = "8080"
	}
	
	srv := &nethttp.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		slog.Info("Starting HTTP server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			slog.Error("Failed to start HTTP server", "error", err)
		}
	}()

	slog.Info("Starting bot...")
	go botClient.Start()

	<-quit

	slog.Info("Shutting down gracefully...")
	botClient.Stop()

	// Shutdown HTTP Server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP Server forced to shutdown", "error", err)
	}
}
