package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

type redisConfig struct {
	host     string
	port     int
	password string
	db       int
}

func InitRedis() error {
	cfg := loadRedisConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.host, cfg.port),
		Password: cfg.password,
		DB:       cfg.db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return fmt.Errorf("ping redis: %w", err)
	}

	redisClient = client
	return nil
}

func GetRedis() *redis.Client {
	return redisClient
}

func CloseRedis() error {
	if redisClient == nil {
		return nil
	}
	return redisClient.Close()
}

func loadRedisConfig() redisConfig {
	return redisConfig{
		host:     getEnv("REDIS_HOST", "localhost"),
		port:     getEnvInt("REDIS_PORT", 6379),
		password: os.Getenv("REDIS_PASSWORD"),
		db:       getEnvInt("REDIS_DB", 0),
	}
}
