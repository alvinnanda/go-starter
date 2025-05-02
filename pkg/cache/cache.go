package cache

import (
	"starter-app/pkg/env"
	"time"
)

// Store defines a cache store interface
type Store interface {
	// Get retrieves an item from the cache by key
	Get(key string) ([]byte, bool)

	// Set adds an item to the cache with an expiration time
	Set(key string, value []byte, expiration time.Duration)

	// Delete removes an item from the cache
	Delete(key string)

	// Reset clears the entire cache
	Reset()

	// Close closes any connections
	Close()
}

// Config represents cache configuration
type Config struct {
	Driver    string
	TTL       time.Duration
	RedisURL  string
	RedisPass string
	RedisDB   int
}

// DefaultConfig returns the default cache configuration
func DefaultConfig() *Config {
	return &Config{
		Driver:    env.Get("CACHE_DRIVER", "memory"),
		TTL:       env.GetDuration("CACHE_TTL", 5*time.Minute),
		RedisURL:  env.Get("REDIS_URL", "localhost:6379"),
		RedisPass: env.Get("REDIS_PASSWORD", ""),
		RedisDB:   env.GetInt("REDIS_DB", 0),
	}
}

// New returns a configured cache store
func New(config *Config) Store {
	switch config.Driver {
	case "redis":
		return NewRedisStore(config)
	default:
		return NewMemoryStore(config)
	}
}
