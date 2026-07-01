package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-image-demo-backend/internal/config"
	"ai-image-demo-backend/internal/database"
	"ai-image-demo-backend/internal/handlers"
	"ai-image-demo-backend/internal/middleware"
	"ai-image-demo-backend/internal/repositories"
	"ai-image-demo-backend/internal/services"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg.MySQLDSN)
	if err != nil {
		logger.Error("failed to connect mysql", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		logger.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	userRepo := repositories.NewUserRepository(db)
	imageRepo := repositories.NewImageRepository(db)
	authService := services.NewAuthService(userRepo, cfg)
	imageService := services.NewImageService(imageRepo, cfg, logger)
	imagePool := services.NewWorkerPool(cfg.WorkerPoolSize, cfg.WorkerPoolSize*4, imageService, logger)
	imagePool.Start()
	defer imagePool.Stop()

	authHandler := handlers.NewAuthHandler(authService)
	imageHandler := handlers.NewImageHandler(imageRepo, imagePool)

	mux := http.NewServeMux()
	mux.Handle("/storage/images/", http.StripPrefix("/storage/images/", http.FileServer(http.Dir(cfg.ImageOutputDir))))
	mux.HandleFunc("/api/health", handlers.Health)
	mux.HandleFunc("/api/auth/register", middleware.WithMethod(http.MethodPost, authHandler.Register))
	mux.HandleFunc("/api/auth/login", middleware.WithMethod(http.MethodPost, authHandler.Login))
	mux.HandleFunc("/api/auth/me", middleware.WithMethod(http.MethodGet, middleware.Auth(cfg, authHandler.Me)))
	mux.HandleFunc("/api/images/generate", middleware.WithMethod(http.MethodPost, middleware.Auth(cfg, imageHandler.Generate)))
	mux.HandleFunc("/api/images", middleware.WithMethod(http.MethodGet, middleware.Auth(cfg, imageHandler.List)))

	handler := middleware.Recover(logger, middleware.CORS(middleware.RequestLogger(logger, mux)))
	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("server started", "port", cfg.AppPort, "worker_pool_size", cfg.WorkerPoolSize)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
