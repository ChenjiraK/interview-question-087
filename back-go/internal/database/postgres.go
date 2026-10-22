package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open connects using DATABASE_URL, or the POSTGRES_* environment variables.
func Open(ctx context.Context) (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		name, username := os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_USER")
		if name == "" || username == "" {
			return nil, errors.New("set DATABASE_URL or POSTGRES_DB and POSTGRES_USER")
		}
		connection := url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(username, os.Getenv("POSTGRES_PASSWORD")),
			Host:   net.JoinHostPort(envOr("POSTGRES_HOST", "127.0.0.1"), envOr("POSTGRES_PORT", "5432")),
			Path:   "/" + name,
		}
		query := url.Values{"sslmode": {envOr("POSTGRES_SSLMODE", "disable")}}
		connection.RawQuery = query.Encode()
		dsn = connection.String()
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableAutomaticPing: true,
		Logger: logger.New(log.New(os.Stderr, "", log.LstdFlags), logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database connection pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	return db, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
