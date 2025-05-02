package database

import (
	"context"
	"fmt"
	"starter-app/config"
	"starter-app/pkg/env"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool creates and returns a new PostgreSQL connection pool
func NewPostgresPool(cfg *config.Config) (*pgxpool.Pool, error) {
	// Create connection string
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	// Create pool configuration
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing postgres config: %w", err)
	}

	// Set pool configurations from environment variables or use defaults
	poolConfig.MaxConns = int32(env.GetInt("DB_POOL_MAX_CONNS", 10))
	poolConfig.MinConns = int32(env.GetInt("DB_POOL_MIN_CONNS", 2))
	poolConfig.MaxConnLifetime = env.GetDuration("DB_POOL_MAX_CONN_LIFETIME", time.Hour)
	poolConfig.MaxConnIdleTime = env.GetDuration("DB_POOL_MAX_CONN_IDLE_TIME", 30*time.Minute)
	poolConfig.HealthCheckPeriod = env.GetDuration("DB_POOL_HEALTH_CHECK_PERIOD", time.Minute)

	// Create the connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating postgres pool: %w", err)
	}

	// Check the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error connecting to postgres: %w", err)
	}

	// Log pool configuration
	fmt.Printf("Database pool initialized with: MaxConns=%d, MinConns=%d, MaxConnLifetime=%s, MaxConnIdleTime=%s\n",
		poolConfig.MaxConns, poolConfig.MinConns, poolConfig.MaxConnLifetime, poolConfig.MaxConnIdleTime)

	return pool, nil
}
