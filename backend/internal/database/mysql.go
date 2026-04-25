package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type mysqlConfig struct {
	host     string
	port     int
	user     string
	password string
	name     string
}

var isContainerRuntime = func() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func InitMySQL() error {
	cfg, err := loadMySQLConfig()
	if err != nil {
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s",
		cfg.user, cfg.password, cfg.host, cfg.port, cfg.name,
	)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql connection: %w", err)
	}

	conn.SetMaxIdleConns(5)
	conn.SetMaxOpenConns(10)
	conn.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return fmt.Errorf("ping mysql: %w", err)
	}

	db = conn
	return nil
}

func Ping() error {
	if db == nil {
		return fmt.Errorf("mysql not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

func GetDB() *sql.DB {
	return db
}

func CloseMySQL() error {
	if db == nil {
		return nil
	}
	return db.Close()
}

func loadMySQLConfig() (mysqlConfig, error) {
	host, hostSet := lookupEnv("DB_HOST", "localhost")
	if requiresExplicitDBHost() && !hostSet {
		return mysqlConfig{}, fmt.Errorf("DB_HOST is required when GIN_MODE=release or running in a container; pass DB_HOST with docker -e or --env-file")
	}

	return mysqlConfig{
		host:     host,
		port:     getEnvInt("DB_PORT", 3306),
		user:     getEnv("DB_USER", "root"),
		password: os.Getenv("DB_PASSWORD"),
		name:     getEnv("DB_NAME", "steellink"),
	}, nil
}

func requiresExplicitDBHost() bool {
	return os.Getenv("GIN_MODE") == "release" || isContainerRuntime()
}

func getEnv(key, fallback string) string {
	value, ok := lookupEnv(key, fallback)
	if ok {
		return value
	}
	return fallback
}

func lookupEnv(key, fallback string) (string, bool) {
	if value := os.Getenv(key); value != "" {
		return value, true
	}
	return fallback, false
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
