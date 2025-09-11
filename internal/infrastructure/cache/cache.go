package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Common cache errors
var (
	ErrCacheMiss     = errors.New("cache miss")
	ErrCacheNotFound = errors.New("cache not found")
	ErrInvalidKey    = errors.New("invalid cache key")
	ErrInvalidValue  = errors.New("invalid cache value")
)

// CacheService defines the interface for cache operations
type CacheService interface {
	// Basic operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	
	// JSON operations
	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	
	// Advanced operations
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Increment(ctx context.Context, key string) (int64, error)
	IncrementBy(ctx context.Context, key string, value int64) (int64, error)
	
	// Atomic operations
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	GetSet(ctx context.Context, key string, value interface{}) (string, error)
	
	// Batch operations
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	MSet(ctx context.Context, pairs ...interface{}) error
	
	// Utility operations
	FlushDB(ctx context.Context) error
	Ping(ctx context.Context) error
	Close() error
}

// RedisClientInterface defines the interface for Redis client operations
type RedisClientInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Increment(ctx context.Context, key string) (int64, error)
	IncrementBy(ctx context.Context, key string, value int64) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	GetSet(ctx context.Context, key string, value interface{}) (string, error)
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	MSet(ctx context.Context, pairs ...interface{}) error
	FlushDB(ctx context.Context) error
	Ping(ctx context.Context) error
	Close() error
	Stats() interface{}
}

// RedisCacheService implements CacheService using Redis
type RedisCacheService struct {
	client RedisClientInterface
}

// NewRedisCacheService creates a new Redis-based cache service
func NewRedisCacheService(client RedisClientInterface) CacheService {
	return &RedisCacheService{
		client: client,
	}
}

// Get retrieves a value from cache
func (r *RedisCacheService) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}
	return r.client.Get(ctx, key)
}

// Set stores a value in cache with expiration
func (r *RedisCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}
	if value == nil {
		return ErrInvalidValue
	}
	return r.client.Set(ctx, key, value, expiration)
}

// Delete removes keys from cache
func (r *RedisCacheService) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return ErrInvalidKey
	}
	for _, key := range keys {
		if key == "" {
			return ErrInvalidKey
		}
	}
	return r.client.Delete(ctx, keys...)
}

// Exists checks if keys exist in cache
func (r *RedisCacheService) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, ErrInvalidKey
	}
	for _, key := range keys {
		if key == "" {
			return 0, ErrInvalidKey
		}
	}
	return r.client.Exists(ctx, keys...)
}

// GetJSON retrieves and unmarshals JSON from cache
func (r *RedisCacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if key == "" {
		return ErrInvalidKey
	}
	if dest == nil {
		return ErrInvalidValue
	}
	
	data, err := r.client.Get(ctx, key)
	if err != nil {
		return err
	}
	
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	return nil
}

// SetJSON marshals and stores JSON in cache
func (r *RedisCacheService) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}
	if value == nil {
		return ErrInvalidValue
	}
	
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return r.client.Set(ctx, key, data, expiration)
}

// Expire sets expiration time for a key
func (r *RedisCacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}
	return r.client.Expire(ctx, key, expiration)
}

// TTL returns the time to live for a key
func (r *RedisCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	if key == "" {
		return 0, ErrInvalidKey
	}
	return r.client.TTL(ctx, key)
}

// Increment increments a key's value
func (r *RedisCacheService) Increment(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return 0, ErrInvalidKey
	}
	return r.client.Increment(ctx, key)
}

// IncrementBy increments a key's value by a specific amount
func (r *RedisCacheService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	if key == "" {
		return 0, ErrInvalidKey
	}
	return r.client.IncrementBy(ctx, key, value)
}

// SetNX sets a key only if it doesn't exist
func (r *RedisCacheService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}
	if value == nil {
		return false, ErrInvalidValue
	}
	return r.client.SetNX(ctx, key, value, expiration)
}

// GetSet atomically sets a key and returns its old value
func (r *RedisCacheService) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}
	if value == nil {
		return "", ErrInvalidValue
	}
	return r.client.GetSet(ctx, key, value)
}

// MGet gets multiple keys at once
func (r *RedisCacheService) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	if len(keys) == 0 {
		return nil, ErrInvalidKey
	}
	for _, key := range keys {
		if key == "" {
			return nil, ErrInvalidKey
		}
	}
	return r.client.MGet(ctx, keys...)
}

// MSet sets multiple keys at once
func (r *RedisCacheService) MSet(ctx context.Context, pairs ...interface{}) error {
	if len(pairs) == 0 || len(pairs)%2 != 0 {
		return ErrInvalidValue
	}
	return r.client.MSet(ctx, pairs...)
}

// FlushDB clears all keys in the current database
func (r *RedisCacheService) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx)
}

// Ping tests the connection to cache
func (r *RedisCacheService) Ping(ctx context.Context) error {
	return r.client.Ping(ctx)
}

// Close closes the cache connection
func (r *RedisCacheService) Close() error {
	return r.client.Close()
}