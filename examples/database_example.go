package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"go-web-starter/internal/config"
	"go-web-starter/internal/domain/model"
	"go-web-starter/internal/infrastructure/database"
	"go-web-starter/internal/infrastructure/logger"
)

// ExampleUser demonstrates how to use BaseModel
type ExampleUser struct {
	model.BaseModel
	Name     string `json:"name" gorm:"not null;size:100"`
	Email    string `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Status   string `json:"status" gorm:"default:'active';size:20"`
	Age      int    `json:"age" gorm:"default:0"`
	Metadata string `json:"metadata" gorm:"type:json"`
}

// TableName specifies the table name for ExampleUser
func (ExampleUser) TableName() string {
	return "example_users"
}

// BeforeCreate hook example
func (u *ExampleUser) BeforeCreate(tx *gorm.DB) error {
	// Call parent BeforeCreate
	if err := u.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}

	// Custom logic
	if u.Status == "" {
		u.Status = "active"
	}

	return nil
}

// Validate implements custom validation
func (u *ExampleUser) Validate() error {
	if u.Name == "" {
		return fmt.Errorf("name is required")
	}
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	if u.Age < 0 {
		return fmt.Errorf("age must be non-negative")
	}
	return nil
}

func main() {
	// Example 1: Database connection setup
	fmt.Println("=== Database Connection Example ===")

	// Create logger
	loggerConfig := &config.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	appLogger, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}

	// Database configuration
	dbConfig := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Username:        "root",
		Password:        "password",
		Database:        "example_db",
		LogLevel:        "info",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 60,
	}

	// Note: This example will fail without a real database
	// In a real application, you would have a MySQL database running
	fmt.Println("Attempting to connect to database...")
	db, err := database.New(dbConfig, appLogger)
	if err != nil {
		fmt.Printf("Failed to connect to database (expected in this example): %v\n", err)
		fmt.Println("To run this example with a real database:")
		fmt.Println("1. Start MySQL server")
		fmt.Println("2. Create database 'example_db'")
		fmt.Println("3. Update connection credentials")
		return
	}
	defer db.Close()

	// Example 2: Database migration
	fmt.Println("\n=== Database Migration Example ===")
	migrator := database.NewMigrator(db.DB, appLogger)

	// Auto migrate models
	if err := migrator.AutoMigrate(&ExampleUser{}); err != nil {
		log.Fatal("Failed to migrate:", err)
	}

	// Create custom indexes
	if err := migrator.CreateIndexes(); err != nil {
		log.Fatal("Failed to create indexes:", err)
	}

	// Example 3: Basic CRUD operations
	fmt.Println("\n=== CRUD Operations Example ===")

	// Create
	user := &ExampleUser{
		Name:     "John Doe",
		Email:    "john@example.com",
		Status:   "active",
		Age:      30,
		Metadata: `{"preferences": {"theme": "dark"}}`,
	}

	// Validate before creating
	if err := user.Validate(); err != nil {
		log.Fatal("Validation failed:", err)
	}

	result := db.Create(user)
	if result.Error != nil {
		log.Fatal("Failed to create user:", result.Error)
	}
	fmt.Printf("Created user with ID: %d\n", user.ID)

	// Read
	var foundUser ExampleUser
	result = db.First(&foundUser, user.ID)
	if result.Error != nil {
		log.Fatal("Failed to find user:", result.Error)
	}
	fmt.Printf("Found user: %+v\n", foundUser)

	// Update
	result = db.Model(&foundUser).Update("age", 31)
	if result.Error != nil {
		log.Fatal("Failed to update user:", result.Error)
	}
	fmt.Printf("Updated user age to: %d\n", foundUser.Age)

	// Soft Delete
	result = db.Delete(&foundUser)
	if result.Error != nil {
		log.Fatal("Failed to delete user:", result.Error)
	}
	fmt.Println("User soft deleted")

	// Try to find deleted user (should fail)
	var deletedUser ExampleUser
	result = db.First(&deletedUser, foundUser.ID)
	if result.Error != nil {
		fmt.Println("Deleted user not found (expected)")
	}

	// Find with deleted records
	result = db.Unscoped().First(&deletedUser, foundUser.ID)
	if result.Error == nil {
		fmt.Printf("Found deleted user: %+v\n", deletedUser)
		fmt.Printf("Deleted at: %v\n", deletedUser.GetDeletedAt())
	}

	// Example 4: Pagination
	fmt.Println("\n=== Pagination Example ===")

	// Create some test users
	testUsers := []ExampleUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
		{Name: "Diana", Email: "diana@example.com", Age: 28},
		{Name: "Eve", Email: "eve@example.com", Age: 32},
	}

	for _, testUser := range testUsers {
		if err := testUser.Validate(); err != nil {
			log.Fatal("Validation failed:", err)
		}
		db.Create(&testUser)
	}

	// Paginated query
	params := &model.PaginationParams{Page: 1, PageSize: 3}
	params.SetDefaults()

	var users []ExampleUser
	var total int64

	// Count total records
	db.Model(&ExampleUser{}).Count(&total)

	// Get paginated results
	result = db.Offset(params.GetOffset()).Limit(params.GetLimit()).Find(&users)
	if result.Error != nil {
		log.Fatal("Failed to get paginated users:", result.Error)
	}

	paginationResult := model.NewPaginationResult(params, total, users)
	fmt.Printf("Pagination result: %+v\n", paginationResult)

	// Example 5: Sorting and filtering
	fmt.Println("\n=== Sorting and Filtering Example ===")

	sortParams := &model.SortParams{SortBy: "age", SortOrder: "desc"}
	sortParams.SetDefaults()

	var sortedUsers []ExampleUser
	result = db.Order(sortParams.GetOrderBy()).Find(&sortedUsers)
	if result.Error != nil {
		log.Fatal("Failed to get sorted users:", result.Error)
	}

	fmt.Printf("Sorted users by %s:\n", sortParams.GetOrderBy())
	for _, u := range sortedUsers {
		fmt.Printf("- %s (age: %d)\n", u.Name, u.Age)
	}

	// Example 6: Transactions
	fmt.Println("\n=== Transaction Example ===")

	err = db.Transaction(func(tx *gorm.DB) error {
		// Create multiple users in a transaction
		user1 := &ExampleUser{Name: "Trans User 1", Email: "trans1@example.com", Age: 25}
		user2 := &ExampleUser{Name: "Trans User 2", Email: "trans2@example.com", Age: 30}

		if err := tx.Create(user1).Error; err != nil {
			return err
		}

		if err := tx.Create(user2).Error; err != nil {
			return err
		}

		fmt.Println("Transaction completed successfully")
		return nil
	})

	if err != nil {
		log.Fatal("Transaction failed:", err)
	}

	// Example 7: Database health check
	fmt.Println("\n=== Health Check Example ===")

	if err := db.Health(); err != nil {
		fmt.Printf("Database health check failed: %v\n", err)
	} else {
		fmt.Println("Database is healthy")
	}

	// Example 8: Database statistics
	fmt.Println("\n=== Database Statistics Example ===")

	stats := db.GetStats()
	fmt.Printf("Database stats: %+v\n", stats)

	// Example 9: Migration status
	fmt.Println("\n=== Migration Status Example ===")

	status := migrator.GetMigrationStatus(&ExampleUser{})
	fmt.Printf("Migration status: %+v\n", status)

	// Example 10: Query logging
	fmt.Println("\n=== Query Logging Example ===")

	start := time.Now()
	var count int64
	db.Model(&ExampleUser{}).Count(&count)
	duration := time.Since(start)

	db.LogQuery("SELECT COUNT(*) FROM example_users", duration, count)

	fmt.Println("Database examples completed successfully!")
}