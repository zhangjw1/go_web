package main

import (
	"errors"
	"time"

	"go-web-starter/internal/config"
	"go-web-starter/internal/infrastructure/logger"
)

func main() {
	// Example 1: Basic logger setup
	println("=== Basic Logger Setup ===")
	loggerConfig := &config.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	log, err := logger.New(loggerConfig)
	if err != nil {
		panic(err)
	}

	// Basic logging
	log.Info("Application started")
	log.Warn("This is a warning")
	log.Error("This is an error")

	// Example 2: Structured logging with context
	println("\n=== Structured Logging ===")
	log.WithRequestID("req-123").
		WithField("user_id", "user-456").
		WithField("duration", time.Millisecond*150).
		Info("User request processed")

	// Example 3: HTTP request logging
	println("\n=== HTTP Request Logging ===")
	log.LogHTTPRequest(
		"GET",
		"/api/users/123",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"192.168.1.100",
		200,
		time.Millisecond*250,
	)

	// Example 4: Database query logging
	println("\n=== Database Query Logging ===")
	log.LogDatabaseQuery(
		"SELECT id, name, email FROM users WHERE active = ? LIMIT ?",
		time.Millisecond*45,
		25,
	)

	// Example 5: Cache operation logging
	println("\n=== Cache Operation Logging ===")
	log.LogCacheOperation("GET", "user:123:profile", true, time.Microsecond*500)
	log.LogCacheOperation("SET", "user:456:session", false, time.Millisecond*2)

	// Example 6: Kafka message logging
	println("\n=== Kafka Message Logging ===")
	log.LogKafkaMessage("user-events", "produce", 1024, time.Millisecond*15)
	log.LogKafkaMessage("notification-events", "consume", 512, time.Millisecond*8)

	// Example 7: Business event logging
	println("\n=== Business Event Logging ===")
	userDetails := map[string]interface{}{
		"email":      "john.doe@example.com",
		"role":       "admin",
		"department": "engineering",
	}
	log.LogBusinessEvent("user_created", "user", 123, userDetails)

	orderDetails := map[string]interface{}{
		"total_amount": 99.99,
		"currency":     "USD",
		"items_count":  3,
	}
	log.LogBusinessEvent("order_completed", "order", "order-789", orderDetails)

	// Example 8: System event logging
	println("\n=== System Event Logging ===")
	systemDetails := map[string]interface{}{
		"version":     "1.0.0",
		"build_time":  "2024-01-15T10:30:00Z",
		"environment": "production",
	}
	log.LogSystemEvent("application_started", systemDetails)

	// Example 9: Error logging with context
	println("\n=== Error Logging ===")
	err = errors.New("database connection failed")
	log.WithError(err).
		WithFields(map[string]interface{}{
			"host":        "db.example.com",
			"port":        3306,
			"database":    "myapp",
			"retry_count": 3,
		}).
		Error("Failed to connect to database after retries")

	// Example 10: Text format logger
	println("\n=== Text Format Logger ===")
	textLoggerConfig := &config.LoggerConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}

	textLog, err := logger.New(textLoggerConfig)
	if err != nil {
		panic(err)
	}

	textLog.WithFields(map[string]interface{}{
		"component": "auth",
		"action":    "login",
		"user_id":   "user-789",
	}).Info("User login successful")

	println("\n=== Logger Examples Complete ===")
}
