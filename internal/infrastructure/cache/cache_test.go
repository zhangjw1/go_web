package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCacheService is a mock implementation of CacheService for testing
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Delete(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockCacheService) Exists(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheService) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCacheService) Increment(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	args := m.Called(ctx, key, value)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	args := m.Called(ctx, key, value, expiration)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	args := m.Called(ctx, key, value)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockCacheService) MSet(ctx context.Context, pairs ...interface{}) error {
	args := m.Called(ctx, pairs)
	return args.Error(0)
}

func (m *MockCacheService) FlushDB(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheService) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheService) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRedisCacheService_Get(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
		errType error
	}{
		{
			name:    "empty key should return error",
			key:     "",
			wantErr: true,
			errType: ErrInvalidKey,
		},
		{
			name:    "valid key should work",
			key:     "test:key",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockRedisClient{}
			service := NewRedisCacheService(mockClient)

			if !tt.wantErr {
				mockClient.On("Get", mock.Anything, tt.key).Return("test-value", nil)
			}

			_, err := service.Get(context.Background(), tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestRedisCacheService_Set(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   interface{}
		wantErr bool
		errType error
	}{
		{
			name:    "empty key should return error",
			key:     "",
			value:   "test",
			wantErr: true,
			errType: ErrInvalidKey,
		},
		{
			name:    "nil value should return error",
			key:     "test:key",
			value:   nil,
			wantErr: true,
			errType: ErrInvalidValue,
		},
		{
			name:    "valid key and value should work",
			key:     "test:key",
			value:   "test-value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockRedisClient{}
			service := NewRedisCacheService(mockClient)

			if !tt.wantErr {
				mockClient.On("Set", mock.Anything, tt.key, tt.value, mock.AnythingOfType("time.Duration")).Return(nil)
			}

			err := service.Set(context.Background(), tt.key, tt.value, time.Minute)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestKeyManager_UserKey(t *testing.T) {
	km := NewKeyManager("test-app")

	tests := []struct {
		name   string
		userID interface{}
		want   string
	}{
		{
			name:   "integer user ID",
			userID: 123,
			want:   "test-app:user:123",
		},
		{
			name:   "string user ID",
			userID: "user-456",
			want:   "test-app:user:user-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := km.UserKey(tt.userID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKeyManager_ValidateKey(t *testing.T) {
	km := NewKeyManager("test-app")

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "empty key should return error",
			key:     "",
			wantErr: true,
		},
		{
			name:    "key with space should return error",
			key:     "test key",
			wantErr: true,
		},
		{
			name:    "key with tab should return error",
			key:     "test\tkey",
			wantErr: true,
		},
		{
			name:    "valid key should pass",
			key:     "test:key:123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := km.ValidateKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKeyManager_GetKeyTTL(t *testing.T) {
	km := NewKeyManager("test-app")

	tests := []struct {
		name    string
		keyType string
		want    time.Duration
	}{
		{
			name:    "user key TTL",
			keyType: "user",
			want:    1 * time.Hour,
		},
		{
			name:    "session key TTL",
			keyType: "session",
			want:    24 * time.Hour,
		},
		{
			name:    "unknown key TTL",
			keyType: "unknown",
			want:    30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := km.GetKeyTTL(tt.keyType)
			assert.Equal(t, tt.want, got)
		})
	}
}

// MockRedisClient is a mock implementation of RedisClient for testing
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) Delete(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockRedisClient) Increment(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	args := m.Called(ctx, key, value)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	args := m.Called(ctx, key, value, expiration)
	return args.Bool(0), args.Error(1)
}

func (m *MockRedisClient) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	args := m.Called(ctx, key, value)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockRedisClient) MSet(ctx context.Context, pairs ...interface{}) error {
	args := m.Called(ctx, pairs)
	return args.Error(0)
}

func (m *MockRedisClient) FlushDB(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) Stats() interface{} {
	args := m.Called()
	return args.Get(0)
}
