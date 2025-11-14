package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	httpapi "inteam/internal/api/http"
	"inteam/internal/auth"
	"inteam/internal/cache"
	"inteam/internal/config"
	"inteam/internal/db"
	"inteam/internal/gigachat"
	"inteam/internal/logger"
	"inteam/internal/metrics"
	"inteam/internal/repository"
	"inteam/internal/service"
	"inteam/internal/storage"
	"inteam/internal/telemetry"
	"inteam/internal/vk"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file (yaml/json)")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	zapLogger, err := logger.New(cfg.Logging)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer func() {
		_ = zapLogger.Sync()
	}()

	ctx = context.WithValue(ctx, struct{}{}, nil) // ensure non-nil ctx for otel init

	otelProvider, err := telemetry.Init(ctx, cfg.Telemetry, zapLogger)
	if err != nil {
		zapLogger.Warn("failed to init telemetry", logger.Error(err))
	}
	defer func() {
		if otelProvider != nil && otelProvider.TracerProvider != nil {
			_ = otelProvider.TracerProvider.Shutdown(context.Background())
		}
	}()

	gormDB, err := db.New(cfg.DB, zapLogger)
	if err != nil {
		zapLogger.Fatal("failed to init db", logger.Error(err))
	}

	if err := db.AutoMigrate(gormDB); err != nil {
		zapLogger.Fatal("failed to run migrations", logger.Error(err))
	}

	httpClient := &http.Client{
		Timeout: cfg.HTTPClient.ClientTimeout,
	}

	redisClient := cache.NewRedis(cfg.Redis, zapLogger)
	minioStorage, err := storage.NewMinio(cfg.Minio, zapLogger)
	if err != nil {
		zapLogger.Warn("failed to init minio, continuing without object storage", logger.Error(err))
	}

	vkClient := vk.NewClient(cfg.VK, httpClient, zapLogger, redisClient)
	gigachatClient := gigachat.NewClient(cfg.GigaChat, httpClient, zapLogger)

	profileRepo := repository.NewProfileRepository(gormDB)
	userRepo := repository.NewUserRepository(gormDB)

	jwtManager := auth.NewJWTManager(cfg.Auth)

	profileService := service.NewProfileService(vkClient, gigachatClient, profileRepo, minioStorage, zapLogger)
	authService := service.NewAuthService(userRepo, jwtManager, zapLogger)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.GinMiddleware(zapLogger))
	if cfg.Metrics.Enabled {
		router.Use(metrics.GinMiddleware())
		router.GET("/metrics", metrics.MetricsHandler())
	}

	httpapi.RegisterRoutes(router, cfg, profileService, authService, jwtManager)

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    cfg.HTTP.ReadTimeout,
		WriteTimeout:   cfg.HTTP.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		zapLogger.Info("starting HTTP server", logger.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("http server error", logger.Error(err))
		}
	}()

	<-ctx.Done()
	stop()
	zapLogger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLogger.Error("server shutdown error", logger.Error(err))
	}

	if err := db.Close(gormDB); err != nil {
		zapLogger.Error("db close error", logger.Error(err))
	}

	os.Exit(0)
}
