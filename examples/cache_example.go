//go:build examples
// +build examples

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-web-starter/internal/config"
	"go-web-starter/internal/infrastructure/cache"
	"go-web-starter/internal/infrastructure/logger"
)

func main() {
	// Create logger
	loggerConfig := &config.LoggerConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}

	appLogger, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}

	// Create Redis configuration
	redisConfig := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 0,
	}

	// Create cache manager
	cacheManager, err := cache.NewManager(redisConfig, appLogger)
	if err != nil {
		log.Fatal("Failed to create cache manager:", err)
	}
	defer cacheManager.Close()

	ctx := context.Background()

	// Test basic cache operations
	fmt.Println("=== Basic Cache Operations ===")
	testBasicOperations(ctx, cacheManager)

	// Test user cache operations
	fmt.Println("\n=== User Cache Operations ===")
	testUserOperations(ctx, cacheManager)

	// Test session cache operations
	fmt.Println("\n=== Session Cache Operations ===")
	testSessionOperations(ctx, cacheManager)

	// Test rate limiting
	fmt.Println("\n=== Rate Limiting ===")
	testRateLimiting(ctx, cacheManager)

	// Test distributed locks
	fmt.Println("\n=== Distributed Locks ===")
	testDistributedLocks(ctx, cacheManager)

	// Test counters
	fmt.Println("\n=== Counters ===")
	testCounters(ctx, cacheManager)

	fmt.Println("\nCache example completed successfully!")
}

func testBasicOperations(ctx context.Context, manager *cache.Manager) {
	service := manager.GetService()

	// Set a value
	err := service.Set(ctx, "test:key", "Hello, Redis!", 5*time.Minute)
	if err != nil {
		fmt.Printf("Error setting value: %v\n", err)
		return
	}
	fmt.Println("✓ Set value successfully")

	// Get the value
	value, err := service.Get(ctx, "test:key")
	if err != nil {
		fmt.Printf("Error getting value: %v\n", err)
		return
	}
	fmt.Printf("✓ Got value: %s\n", value)

	// Check if key exists
	exists, err := service.Exists(ctx, "test:key")
	if err != nil {
		fmt.Printf("Error checking existence: %v\n", err)
		return
	}
	fmt.Printf("✓ Key exists: %d\n", exists)

	// Get TTL
	ttl, err := service.TTL(ctx, "test:key")
	if err != nil {
		fmt.Printf("Error getting TTL: %v\n", err)
		return
	}
	fmt.Printf("✓ TTL: %v\n", ttl)

	// Delete the key
	err = service.Delete(ctx, "test:key")
	if err != nil {
		fmt.Printf("Error deleting key: %v\n", err)
		return
	}
	fmt.Println("✓ Deleted key successfully")
}

func testUserOperations(ctx context.Context, manager *cache.Manager) {
	// Create user data
	user := map[string]interface{}{
		"id":       123,
		"username": "john_doe",
		"email":    "john@example.com",
		"name":     "John Doe",
		"active":   true,
	}

	// Set user in cache
	err := manager.SetUser(ctx, 123, user)
	if err != nil {
		fmt.Printf("Error setting user: %v\n", err)
		return
	}
	fmt.Println("✓ Set user successfully")

	// Get user from cache
	cachedUser, err := manager.GetUser(ctx, 123)
	if err != nil {
		fmt.Printf("Error getting user: %v\n", err)
		return
	}
	fmt.Printf("✓ Got user: %+v\n", cachedUser)

	// Set user by username
	err = manager.SetUserByUsername(ctx, "john_doe", user)
	if err != nil {
		fmt.Printf("Error setting user by username: %v\n", err)
		return
	}
	fmt.Println("✓ Set user by username successfully")

	// Get user by username
	userByUsername, err := manager.GetUserByUsername(ctx, "john_doe")
	if err != nil {
		fmt.Printf("Error getting user by username: %v\n", err)
		return
	}
	fmt.Printf("✓ Got user by username: %+v\n", userByUsername)

	// Delete user
	err = manager.DeleteUser(ctx, 123)
	if err != nil {
		fmt.Printf("Error deleting user: %v\n", err)
		return
	}
	fmt.Println("✓ Deleted user successfully")
}

func testSessionOperations(ctx context.Context, manager *cache.Manager) {
	sessionID := "session_123456"
	sessionData := map[string]interface{}{
		"user_id":    123,
		"username":   "john_doe",
		"created_at": time.Now().Unix(),
		"expires_at": time.Now().Add(24 * time.Hour).Unix(),
	}

	// Set session
	err := manager.SetSession(ctx, sessionID, sessionData)
	if err != nil {
		fmt.Printf("Error setting session: %v\n", err)
		return
	}
	fmt.Println("✓ Set session successfully")

	// Get session
	cachedSession, err := manager.GetSession(ctx, sessionID)
	if err != nil {
		fmt.Printf("Error getting session: %v\n", err)
		return
	}
	fmt.Printf("✓ Got session: %+v\n", cachedSession)

	// Delete session
	err = manager.DeleteSession(ctx, sessionID)
	if err != nil {
		fmt.Printf("Error deleting session: %v\n", err)
		return
	}
	fmt.Println("✓ Deleted session successfully")
}

func testRateLimiting(ctx context.Context, manager *cache.Manager) {
	identifier := "user:123"
	action := "api_call"
	limit := int64(5)
	window := 1 * time.Minute

	fmt.Printf("Testing rate limit: %d requests per %v\n", limit, window)

	// Test multiple requests
	for i := 1; i <= 7; i++ {
		exceeded, remaining, err := manager.CheckRateLimit(ctx, identifier, action, limit, window)
		if err != nil {
			fmt.Printf("Error checking rate limit: %v\n", err)
			return
		}

		status := "✓ Allowed"
		if exceeded {
			status = "✗ Rate limited"
		}

		fmt.Printf("Request %d: %s (remaining: %d)\n", i, status, remaining)

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	// Reset rate limit
	err := manager.ResetRateLimit(ctx, identifier, action)
	if err != nil {
		fmt.Printf("Error resetting rate limit: %v\n", err)
		return
	}
	fmt.Println("✓ Reset rate limit successfully")
}

func testDistributedLocks(ctx context.Context, manager *cache.Manager) {
	resource := "critical_section"
	lockTTL := 30 * time.Second

	// Acquire lock
	acquired, err := manager.AcquireLock(ctx, resource, lockTTL)
	if err != nil {
		fmt.Printf("Error acquiring lock: %v\n", err)
		return
	}

	if acquired {
		fmt.Println("✓ Lock acquired successfully")

		// Simulate work
		fmt.Println("  Doing critical work...")
		time.Sleep(1 * time.Second)

		// Release lock
		err = manager.ReleaseLock(ctx, resource)
		if err != nil {
			fmt.Printf("Error releasing lock: %v\n", err)
			return
		}
		fmt.Println("✓ Lock released successfully")
	} else {
		fmt.Println("✗ Failed to acquire lock (already held)")
	}

	// Try to acquire the same lock again (should succeed now)
	acquired, err = manager.AcquireLock(ctx, resource, lockTTL)
	if err != nil {
		fmt.Printf("Error acquiring lock (second attempt): %v\n", err)
		return
	}

	if acquired {
		fmt.Println("✓ Lock acquired successfully (second attempt)")
		manager.ReleaseLock(ctx, resource)
	}
}

func testCounters(ctx context.Context, manager *cache.Manager) {
	counterName := "page_views"

	// Increment counter multiple times
	for i := 1; i <= 5; i++ {
		count, err := manager.IncrementCounter(ctx, counterName)
		if err != nil {
			fmt.Printf("Error incrementing counter: %v\n", err)
			return
		}
		fmt.Printf("✓ Counter incremented to: %d\n", count)
	}

	// Increment by specific amount
	count, err := manager.IncrementCounterBy(ctx, counterName, 10)
	if err != nil {
		fmt.Printf("Error incrementing counter by 10: %v\n", err)
		return
	}
	fmt.Printf("✓ Counter incremented by 10 to: %d\n", count)

	// Get current counter value
	currentCount, err := manager.GetCounter(ctx, counterName)
	if err != nil {
		fmt.Printf("Error getting counter: %v\n", err)
		return
	}
	fmt.Printf("✓ Current counter value: %d\n", currentCount)

	// Reset counter
	err = manager.ResetCounter(ctx, counterName)
	if err != nil {
		fmt.Printf("Error resetting counter: %v\n", err)
		return
	}
	fmt.Println("✓ Counter reset successfully")

	// Verify counter is reset
	resetCount, err := manager.GetCounter(ctx, counterName)
	if err != nil {
		fmt.Printf("Error getting reset counter: %v\n", err)
		return
	}
	fmt.Printf("✓ Counter after reset: %d\n", resetCount)
}
