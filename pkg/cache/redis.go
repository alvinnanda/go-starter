package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore provides a Redis-based cache implementation
type RedisStore struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// NewRedisStore creates a new Redis store
func NewRedisStore(config *Config) Store {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisURL,
		Password: config.RedisPass,
		DB:       config.RedisDB,
	})

	ttl := config.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	return &RedisStore{
		client:     client,
		defaultTTL: ttl,
	}
}

// Get retrieves an item from the cache by key
func (s *RedisStore) Get(key string) ([]byte, bool) {
	ctx := context.Background()
	val, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	return val, true
}

// Set adds an item to the cache with an expiration time
func (s *RedisStore) Set(key string, value []byte, expiration time.Duration) {
	ctx := context.Background()

	if expiration == 0 {
		expiration = s.defaultTTL
	}

	s.client.Set(ctx, key, value, expiration)
}

// Delete removes an item from the cache
func (s *RedisStore) Delete(key string) {
	ctx := context.Background()
	s.client.Del(ctx, key)
}

// Reset clears the entire cache
func (s *RedisStore) Reset() {
	ctx := context.Background()
	s.client.FlushDB(ctx)
}

// Close closes the Redis client
func (s *RedisStore) Close() {
	s.client.Close()
}
