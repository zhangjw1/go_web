package cache

import (
	"fmt"
	"strings"
	"time"
)

// KeyManager manages cache key generation and validation
type KeyManager struct {
	prefix    string
	separator string
}

// NewKeyManager creates a new key manager
func NewKeyManager(prefix string) *KeyManager {
	return &KeyManager{
		prefix:    prefix,
		separator: ":",
	}
}

// UserKey generates a cache key for user data
func (km *KeyManager) UserKey(userID interface{}) string {
	return km.buildKey("user", userID)
}

// UserByUsernameKey generates a cache key for user lookup by username
func (km *KeyManager) UserByUsernameKey(username string) string {
	return km.buildKey("user", "username", username)
}

// UserByEmailKey generates a cache key for user lookup by email
func (km *KeyManager) UserByEmailKey(email string) string {
	return km.buildKey("user", "email", email)
}

// UserSessionKey generates a cache key for user session
func (km *KeyManager) UserSessionKey(sessionID string) string {
	return km.buildKey("session", sessionID)
}

// UserTokenKey generates a cache key for user token
func (km *KeyManager) UserTokenKey(tokenID string) string {
	return km.buildKey("token", tokenID)
}

// UserListKey generates a cache key for user list with pagination
func (km *KeyManager) UserListKey(page, perPage int, filters map[string]interface{}) string {
	parts := []interface{}{"users", "list", fmt.Sprintf("page_%d", page), fmt.Sprintf("per_page_%d", perPage)}

	// Add filters to key
	if len(filters) > 0 {
		var filterParts []string
		for k, v := range filters {
			filterParts = append(filterParts, fmt.Sprintf("%s_%v", k, v))
		}
		parts = append(parts, strings.Join(filterParts, "_"))
	}

	return km.buildKey(parts...)
}

// RateLimitKey generates a cache key for rate limiting
func (km *KeyManager) RateLimitKey(identifier, action string) string {
	return km.buildKey("rate_limit", action, identifier)
}

// LockKey generates a cache key for distributed locks
func (km *KeyManager) LockKey(resource string) string {
	return km.buildKey("lock", resource)
}

// CounterKey generates a cache key for counters
func (km *KeyManager) CounterKey(name string) string {
	return km.buildKey("counter", name)
}

// StatsKey generates a cache key for statistics
func (km *KeyManager) StatsKey(metric string, period time.Duration) string {
	periodStr := km.formatPeriod(period)
	return km.buildKey("stats", metric, periodStr)
}

// ConfigKey generates a cache key for configuration
func (km *KeyManager) ConfigKey(configName string) string {
	return km.buildKey("config", configName)
}

// HealthCheckKey generates a cache key for health check results
func (km *KeyManager) HealthCheckKey(service string) string {
	return km.buildKey("health", service)
}

// buildKey constructs a cache key from parts
func (km *KeyManager) buildKey(parts ...interface{}) string {
	var keyParts []string

	// Add prefix if set
	if km.prefix != "" {
		keyParts = append(keyParts, km.prefix)
	}

	// Add all parts
	for _, part := range parts {
		keyParts = append(keyParts, fmt.Sprintf("%v", part))
	}

	return strings.Join(keyParts, km.separator)
}

// formatPeriod formats a time duration for use in cache keys
func (km *KeyManager) formatPeriod(period time.Duration) string {
	switch {
	case period >= 24*time.Hour:
		days := int(period.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	case period >= time.Hour:
		hours := int(period.Hours())
		return fmt.Sprintf("%dh", hours)
	case period >= time.Minute:
		minutes := int(period.Minutes())
		return fmt.Sprintf("%dm", minutes)
	default:
		seconds := int(period.Seconds())
		return fmt.Sprintf("%ds", seconds)
	}
}

// ValidateKey validates a cache key
func (km *KeyManager) ValidateKey(key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	// Check for invalid characters
	invalidChars := []string{" ", "\t", "\n", "\r"}
	for _, char := range invalidChars {
		if strings.Contains(key, char) {
			return fmt.Errorf("key contains invalid character: %s", char)
		}
	}

	// Check key length (Redis has a limit of 512MB, but we'll use a reasonable limit)
	if len(key) > 250 {
		return fmt.Errorf("key too long: %d characters (max 250)", len(key))
	}

	return nil
}

// ParseKey parses a cache key and returns its parts
func (km *KeyManager) ParseKey(key string) []string {
	return strings.Split(key, km.separator)
}

// IsUserKey checks if a key is a user-related key
func (km *KeyManager) IsUserKey(key string) bool {
	parts := km.ParseKey(key)
	if len(parts) < 2 {
		return false
	}

	// Skip prefix if present
	startIndex := 0
	if km.prefix != "" && len(parts) > 0 && parts[0] == km.prefix {
		startIndex = 1
	}

	return len(parts) > startIndex && parts[startIndex] == "user"
}

// IsSessionKey checks if a key is a session-related key
func (km *KeyManager) IsSessionKey(key string) bool {
	parts := km.ParseKey(key)
	if len(parts) < 2 {
		return false
	}

	// Skip prefix if present
	startIndex := 0
	if km.prefix != "" && len(parts) > 0 && parts[0] == km.prefix {
		startIndex = 1
	}

	return len(parts) > startIndex && parts[startIndex] == "session"
}

// GetKeyTTL returns the recommended TTL for different key types
func (km *KeyManager) GetKeyTTL(keyType string) time.Duration {
	switch keyType {
	case "user":
		return 1 * time.Hour
	case "session":
		return 24 * time.Hour
	case "token":
		return 15 * time.Minute
	case "rate_limit":
		return 1 * time.Minute
	case "lock":
		return 30 * time.Second
	case "counter":
		return 24 * time.Hour
	case "stats":
		return 1 * time.Hour
	case "config":
		return 30 * time.Minute
	case "health":
		return 5 * time.Minute
	default:
		return 30 * time.Minute
	}
}

// GetKeyTypeFromKey extracts the key type from a cache key
func (km *KeyManager) GetKeyTypeFromKey(key string) string {
	parts := km.ParseKey(key)
	if len(parts) < 2 {
		return "unknown"
	}

	// Skip prefix if present
	startIndex := 0
	if km.prefix != "" && len(parts) > 0 && parts[0] == km.prefix {
		startIndex = 1
	}

	if len(parts) > startIndex {
		return parts[startIndex]
	}

	return "unknown"
}
