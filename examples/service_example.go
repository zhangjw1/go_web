package main

import (
	"context"
	"fmt"
	"log"

	"go-web-starter/internal/config"
	"go-web-starter/internal/domain/service"
	"go-web-starter/internal/infrastructure/cache"
	"go-web-starter/internal/infrastructure/logger"
	"go-web-starter/internal/infrastructure/messaging"
	serviceImpl "go-web-starter/internal/infrastructure/service"
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

	// Create cache manager (optional - can be nil)
	var cacheManager *cache.Manager
	redisConfig := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 0,
	}
	
	cacheManager, err = cache.NewManager(redisConfig, appLogger)
	if err != nil {
		fmt.Printf("Warning: Failed to create cache manager: %v\n", err)
		cacheManager = nil // Continue without cache
	}
	defer func() {
		if cacheManager != nil {
			cacheManager.Close()
		}
	}()

	// Create messaging manager (optional - can be nil)
	var messagingManager *messaging.Manager
	kafkaConfig := &config.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "go-web-starter-example",
	}
	
	messagingManager, err = messaging.NewManager(kafkaConfig, appLogger)
	if err != nil {
		fmt.Printf("Warning: Failed to create messaging manager: %v\n", err)
		messagingManager = nil // Continue without messaging
	}
	defer func() {
		if messagingManager != nil {
			messagingManager.Close()
		}
	}()

	// Note: In a real application, you would also create a user repository
	// For this example, we'll create a mock or skip repository-dependent operations
	fmt.Println("=== Service Layer Example ===")
	fmt.Println("Note: This example demonstrates service interfaces and validation")
	fmt.Println("Repository operations are skipped as they require database setup")

	ctx := context.Background()

	// Test service validation functions
	fmt.Println("\n=== Testing Service Validation ===")
	testServiceValidation()

	// Test message service (if messaging is available)
	if messagingManager != nil {
		fmt.Println("\n=== Testing Message Service ===")
		testMessageService(ctx, messagingManager, cacheManager, appLogger)
	}

	// Test service helper functions
	fmt.Println("\n=== Testing Service Helpers ===")
	testServiceHelpers()

	fmt.Println("\nService example completed!")
}

func testServiceValidation() {
	// Test CreateUserRequest validation
	fmt.Println("Testing CreateUserRequest validation...")

	validRequest := &service.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		FullName: "Test User",
	}

	err := service.ValidateCreateUserRequest(validRequest)
	if err != nil {
		fmt.Printf("❌ Valid request failed validation: %v\n", err)
	} else {
		fmt.Println("✅ Valid request passed validation")
	}

	invalidRequest := &service.CreateUserRequest{
		Username: "", // Empty username
		Email:    "test@example.com",
		Password: "password123",
	}

	err = service.ValidateCreateUserRequest(invalidRequest)
	if err != nil {
		fmt.Printf("✅ Invalid request correctly failed validation: %v\n", err)
	} else {
		fmt.Println("❌ Invalid request incorrectly passed validation")
	}

	// Test UpdateUserRequest validation
	fmt.Println("\nTesting UpdateUserRequest validation...")

	validUpdateRequest := &service.UpdateUserRequest{
		FullName: stringPtr("Updated Name"),
	}

	err = service.ValidateUpdateUserRequest(validUpdateRequest)
	if err != nil {
		fmt.Printf("❌ Valid update request failed validation: %v\n", err)
	} else {
		fmt.Println("✅ Valid update request passed validation")
	}

	emptyUpdateRequest := &service.UpdateUserRequest{}

	err = service.ValidateUpdateUserRequest(emptyUpdateRequest)
	if err != nil {
		fmt.Printf("✅ Empty update request correctly failed validation: %v\n", err)
	} else {
		fmt.Println("❌ Empty update request incorrectly passed validation")
	}

	// Test ListUsersRequest validation
	fmt.Println("\nTesting ListUsersRequest validation...")

	listRequest := &service.ListUsersRequest{
		Page:    0,    // Invalid page
		PerPage: 200,  // Too large per page
	}

	err = service.ValidateListUsersRequest(listRequest)
	if err != nil {
		fmt.Printf("❌ List request validation failed: %v\n", err)
	} else {
		fmt.Printf("✅ List request validation passed (corrected values: page=%d, per_page=%d)\n", 
			listRequest.Page, listRequest.PerPage)
	}
}

func testMessageService(ctx context.Context, messagingManager *messaging.Manager, cacheManager *cache.Manager, appLogger *logger.Logger) {
	// Create message service
	messageService := serviceImpl.NewMessageService(messagingManager, cacheManager, appLogger)

	// Test user event handling
	fmt.Println("Testing user event handling...")

	userData := map[string]interface{}{
		"username":  "testuser",
		"email":     "test@example.com",
		"full_name": "Test User",
	}

	err := messageService.HandleUserCreated(ctx, 123, userData)
	if err != nil {
		fmt.Printf("❌ HandleUserCreated failed: %v\n", err)
	} else {
		fmt.Println("✅ HandleUserCreated succeeded")
	}

	changes := map[string]interface{}{
		"full_name": map[string]interface{}{
			"old": "Test User",
			"new": "Updated Test User",
		},
	}

	err = messageService.HandleUserUpdated(ctx, 123, changes)
	if err != nil {
		fmt.Printf("❌ HandleUserUpdated failed: %v\n", err)
	} else {
		fmt.Println("✅ HandleUserUpdated succeeded")
	}

	err = messageService.HandleUserDeleted(ctx, 123)
	if err != nil {
		fmt.Printf("❌ HandleUserDeleted failed: %v\n", err)
	} else {
		fmt.Println("✅ HandleUserDeleted succeeded")
	}

	// Test system event handling
	fmt.Println("\nTesting system event handling...")

	serviceData := map[string]interface{}{
		"version":     "1.0.0",
		"environment": "development",
	}

	err = messageService.HandleSystemStartup(ctx, serviceData)
	if err != nil {
		fmt.Printf("❌ HandleSystemStartup failed: %v\n", err)
	} else {
		fmt.Println("✅ HandleSystemStartup succeeded")
	}

	healthData := map[string]interface{}{
		"status": "healthy",
		"checks": map[string]interface{}{
			"database": "ok",
			"cache":    "ok",
		},
	}

	err = messageService.HandleHealthCheck(ctx, healthData)
	if err != nil {
		fmt.Printf("❌ HandleHealthCheck failed: %v\n", err)
	} else {
		fmt.Println("✅ HandleHealthCheck succeeded")
	}

	// Test notification handling
	fmt.Println("\nTesting notification handling...")

	notificationData := map[string]interface{}{
		"template": "welcome",
	}

	err = messageService.HandleEmailNotification(ctx, "test@example.com", "Welcome", "Welcome to our service!", notificationData)
	if err != nil {
		fmt.Printf("❌ HandleEmailNotification failed: %v\n", err)
	} else {
		fmt.Println("✅ HandleEmailNotification succeeded")
	}

	err = messageService.HandleSMSNotification(ctx, "+1234567890", "Welcome to our service!", notificationData)
	if err != nil {
		fmt.Printf("❌ HandleSMSNotification failed: %v\n", err)
	} else {
		fmt.Println("✅ HandleSMSNotification succeeded")
	}

	// Test audit event handling
	fmt.Println("\nTesting audit event handling...")

	auditDetails := map[string]interface{}{
		"action":   "CREATE",
		"resource": "user",
	}

	err = messageService.HandleAuditEvent(ctx, 123, "CREATE", "user", auditDetails, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		fmt.Printf("❌ HandleAuditEvent failed: %v\n", err)
	} else {
		fmt.Println("✅ HandleAuditEvent succeeded")
	}
}

func testServiceHelpers() {
	// Test CalculateTotalPages
	fmt.Println("Testing CalculateTotalPages...")

	tests := []struct {
		total   int64
		perPage int
		want    int
	}{
		{100, 10, 10},
		{105, 10, 11},
		{0, 10, 0},
		{5, 10, 1},
	}

	for _, test := range tests {
		result := service.CalculateTotalPages(test.total, test.perPage)
		if result == test.want {
			fmt.Printf("✅ CalculateTotalPages(%d, %d) = %d\n", test.total, test.perPage, result)
		} else {
			fmt.Printf("❌ CalculateTotalPages(%d, %d) = %d, want %d\n", test.total, test.perPage, result, test.want)
		}
	}

	// Test IsValidSortField
	fmt.Println("\nTesting IsValidSortField...")

	validFields := []string{"id", "username", "email", "created_at", "updated_at"}
	invalidFields := []string{"invalid", "password", "secret"}

	for _, field := range validFields {
		if service.IsValidSortField(field) {
			fmt.Printf("✅ IsValidSortField(%s) = true\n", field)
		} else {
			fmt.Printf("❌ IsValidSortField(%s) = false, want true\n", field)
		}
	}

	for _, field := range invalidFields {
		if !service.IsValidSortField(field) {
			fmt.Printf("✅ IsValidSortField(%s) = false\n", field)
		} else {
			fmt.Printf("❌ IsValidSortField(%s) = true, want false\n", field)
		}
	}

	// Test SanitizeSearchQuery
	fmt.Println("\nTesting SanitizeSearchQuery...")

	queries := []struct {
		input string
		want  string
	}{
		{"  test query  ", "test query"},
		{"normal query", "normal query"},
		{"", ""},
	}

	for _, query := range queries {
		result := service.SanitizeSearchQuery(query.input)
		if result == query.want {
			fmt.Printf("✅ SanitizeSearchQuery(%q) = %q\n", query.input, result)
		} else {
			fmt.Printf("❌ SanitizeSearchQuery(%q) = %q, want %q\n", query.input, result, query.want)
		}
	}

	// Test service errors
	fmt.Println("\nTesting service errors...")

	err := service.NewServiceError("TEST_ERROR", "This is a test error", map[string]interface{}{
		"field": "test_field",
		"value": "test_value",
	})

	fmt.Printf("✅ Created service error: %s\n", err.Error())
	fmt.Printf("   Code: %s\n", err.Code)
	fmt.Printf("   Details: %+v\n", err.Details)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}