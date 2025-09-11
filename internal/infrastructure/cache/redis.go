package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"go-web-starter/internal/config"
	"go-web-starter/internal/infrastructure/logger"
)

// RedisClient wraps redis.Client with additional functionality
type RedisClient struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisClient creates a new Redis client instance
func NewRedisClient(cfg *config.RedisConfig, log *logger.Logger) (*RedisClient, error) {
	// Create Redis client options
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
		
		// Connection pool settings
		PoolSize:     10,
		MinIdleConns: 5,
		
		// Timeouts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		
		// Retry settings
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	}

	// Create Redis client
	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.WithField("host", cfg.Host).WithField("port", cfg.Port).WithField("database", cfg.Database).Info("Redis client connected successfully")

	return &RedisClient{
		client: client,
		logger: log,
	}, nil
}

// Get retrieves a value from Redis
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	
	result, err := r.client.Get(ctx, key).Result()
	duration := time.Since(start)
	
	if err != nil {
		if err == redis.Nil {
			r.logger.LogCacheOperation("GET", key, false, duration)
			return "", ErrCacheMiss
		}
		r.logger.WithError(err).WithField("key", key).WithField("duration", duration).Error("Redis GET failed")
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	
	r.logger.LogCacheOperation("GET", key, true, duration)
	return result, nil
}

// Set stores a value in Redis with expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	
	err := r.client.Set(ctx, key, value, expiration).Err()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("key", key).WithField("duration", duration).Error("Redis SET failed")
		return fmt.Errorf("redis set failed: %w", err)
	}
	
	r.logger.LogCacheOperation("SET", key, true, duration)
	return nil
}

// Delete removes a key from Redis
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	start := time.Now()
	
	err := r.client.Del(ctx, keys...).Err()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("keys", keys).WithField("duration", duration).Error("Redis DELETE failed")
		return fmt.Errorf("redis delete failed: %w", err)
	}
	
	r.logger.LogCacheOperation("DELETE", fmt.Sprintf("%v", keys), true, duration)
	return nil
}

// Exists checks if a key exists in Redis
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	start := time.Now()
	
	result, err := r.client.Exists(ctx, keys...).Result()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("keys", keys).WithField("duration", duration).Error("Redis EXISTS failed")
		return 0, fmt.Errorf("redis exists failed: %w", err)
	}
	
	r.logger.LogCacheOperation("EXISTS", fmt.Sprintf("%v", keys), true, duration)
	return result, nil
}

// Expire sets expiration time for a key
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	start := time.Now()
	
	err := r.client.Expire(ctx, key, expiration).Err()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("key", key).WithField("duration", duration).Error("Redis EXPIRE failed")
		return fmt.Errorf("redis expire failed: %w", err)
	}
	
	r.logger.LogCacheOperation("EXPIRE", key, true, duration)
	return nil
}

// TTL returns the time to live for a key
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	start := time.Now()
	
	result, err := r.client.TTL(ctx, key).Result()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("key", key).WithField("duration", duration).Error("Redis TTL failed")
		return 0, fmt.Errorf("redis ttl failed: %w", err)
	}
	
	r.logger.LogCacheOperation("TTL", key, true, duration)
	return result, nil
}

// Increment increments a key's value
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	
	result, err := r.client.Incr(ctx, key).Result()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("key", key).WithField("duration", duration).Error("Redis INCR failed")
		return 0, fmt.Errorf("redis incr failed: %w", err)
	}
	
	r.logger.LogCacheOperation("INCR", key, true, duration)
	return result, nil
}

// IncrementBy increments a key's value by a specific amount
func (r *RedisClient) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	start := time.Now()
	
	result, err := r.client.IncrBy(ctx, key, value).Result()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("key", key).WithField("value", value).WithField("duration", duration).Error("Redis INCRBY failed")
		return 0, fmt.Errorf("redis incrby failed: %w", err)
	}
	
	r.logger.LogCacheOperation("INCRBY", key, true, duration)
	return result, nil
}

// SetNX sets a key only if it doesn't exist
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	start := time.Now()
	
	result, err := r.client.SetNX(ctx, key, value, expiration).Result()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("key", key).WithField("duration", duration).Error("Redis SETNX failed")
		return false, fmt.Errorf("redis setnx failed: %w", err)
	}
	
	r.logger.LogCacheOperation("SETNX", key, result, duration)
	return result, nil
}

// GetSet atomically sets a key and returns its old value
func (r *RedisClient) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	start := time.Now()
	
	result, err := r.client.GetSet(ctx, key, value).Result()
	duration := time.Since(start)
	
	if err != nil && err != redis.Nil {
		r.logger.WithError(err).WithField("key", key).WithField("duration", duration).Error("Redis GETSET failed")
		return "", fmt.Errorf("redis getset failed: %w", err)
	}
	
	r.logger.LogCacheOperation("GETSET", key, true, duration)
	return result, nil
}

// MGet gets multiple keys at once
func (r *RedisClient) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	start := time.Now()
	
	result, err := r.client.MGet(ctx, keys...).Result()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("keys", keys).WithField("duration", duration).Error("Redis MGET failed")
		return nil, fmt.Errorf("redis mget failed: %w", err)
	}
	
	r.logger.LogCacheOperation("MGET", fmt.Sprintf("%v", keys), true, duration)
	return result, nil
}

// MSet sets multiple keys at once
func (r *RedisClient) MSet(ctx context.Context, pairs ...interface{}) error {
	start := time.Now()
	
	err := r.client.MSet(ctx, pairs...).Err()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("pairs", pairs).WithField("duration", duration).Error("Redis MSET failed")
		return fmt.Errorf("redis mset failed: %w", err)
	}
	
	r.logger.LogCacheOperation("MSET", fmt.Sprintf("%v", pairs), true, duration)
	return nil
}

// FlushDB clears all keys in the current database
func (r *RedisClient) FlushDB(ctx context.Context) error {
	start := time.Now()
	
	err := r.client.FlushDB(ctx).Err()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("duration", duration).Error("Redis FLUSHDB failed")
		return fmt.Errorf("redis flushdb failed: %w", err)
	}
	
	r.logger.LogCacheOperation("FLUSHDB", "all", true, duration)
	return nil
}

// Ping tests the connection to Redis
func (r *RedisClient) Ping(ctx context.Context) error {
	start := time.Now()
	
	err := r.client.Ping(ctx).Err()
	duration := time.Since(start)
	
	if err != nil {
		r.logger.WithError(err).WithField("duration", duration).Error("Redis PING failed")
		return fmt.Errorf("redis ping failed: %w", err)
	}
	
	r.logger.LogCacheOperation("PING", "server", true, duration)
	return nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.WithError(err).Error("Failed to close Redis connection")
			return fmt.Errorf("failed to close redis connection: %w", err)
		}
		r.logger.Info("Redis connection closed")
	}
	return nil
}

// GetClient returns the underlying Redis client (for advanced operations)
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Stats returns Redis connection statistics
func (r *RedisClient) Stats() interface{} {
	return r.client.PoolStats()
}