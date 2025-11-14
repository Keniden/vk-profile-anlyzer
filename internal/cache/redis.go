package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"inteam/internal/config"
)

func NewRedis(cfg config.RedisConfig, logger *zap.Logger) redis.UniversalClient {
	if cfg.Addr == "" {
		logger.Info("redis disabled: addr is empty")
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("redis ping failed, disabling cache", zap.Error(err))
		return nil
	}

	logger.Info("redis connected", zap.String("addr", cfg.Addr))
	return client
}

