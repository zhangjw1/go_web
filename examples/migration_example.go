package main

import (
	"fmt"
	"log"

	"go-web-starter/internal/config"
	"go-web-starter/internal/domain/model"
	"go-web-starter/internal/infrastructure/database"
	"go-web-starter/internal/infrastructure/logger"
)

func main() {
	fmt.Println("=== Database Migration Example ===")

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
		Database:        "go_web_starter",
		LogLevel:        "info",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 60,
	}

	// Note: This example will fail without a real database
	fmt.Println("Attempting to connect to database...")
	db, err := database.New(dbConfig, appLogger)
	if err != nil {
		fmt.Printf("Failed to connect to database (expected in this example): %v\n", err)
		fmt.Println("To run this example with a real database:")
		fmt.Println("1. Start MySQL server")
		fmt.Println("2. Create database 'go_web_starter'")
		fmt.Println("3. Update connection credentials in the code")
		return
	}
	defer db.Close()

	// Create migration manager
	migrationManager := database.NewMigrationManager(db.DB, appLogger)

	// Example 1: Get all models
	fmt.Println("\n=== Available Models ===")
	models := migrationManager.GetAllModels()
	for i, model := range models {
		fmt.Printf("%d. %T\n", i+1, model)
	}

	// Example 2: Run all migrations
	fmt.Println("\n=== Running Migrations ===")
	err = migrationManager.MigrateAll()
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	fmt.Println("✅ All migrations completed successfully")

	// Example 3: Get migration status
	fmt.Println("\n=== Migration Status ===")
	status := migrationManager.GetMigrationStatus()
	for modelName, info := range status {
		fmt.Printf("Model: %s\n", modelName)
		if infoMap, ok := info.(map[string]interface{}); ok {
			for key, value := range infoMap {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
		fmt.Println()
	}

	// Example 4: Validate schema
	fmt.Println("\n=== Schema Validation ===")
	err = migrationManager.ValidateSchema()
	if err != nil {
		fmt.Printf("❌ Schema validation failed: %v\n", err)
	} else {
		fmt.Println("✅ Schema validation passed")
	}

	// Example 5: Seed initial data
	fmt.Println("\n=== Seeding Data ===")
	err = migrationManager.SeedData()
	if err != nil {
		log.Fatal("Data seeding failed:", err)
	}
	fmt.Println("✅ Data seeding completed")

	// Example 6: Verify seeded data
	fmt.Println("\n=== Verifying Seeded Data ===")
	var users []model.User
	result := db.Find(&users)
	if result.Error != nil {
		log.Fatal("Failed to query users:", result.Error)
	}

	fmt.Printf("Found %d users:\n", len(users))
	for _, user := range users {
		fmt.Printf("- ID: %d, Username: %s, Email: %s, Status: %s\n",
			user.ID, user.Username, user.Email, user.Status)
	}

	// Example 7: Query users with profiles
	fmt.Println("\n=== Users with Profiles ===")
	var usersWithProfiles []model.User
	result = db.Preload("Profile").Find(&usersWithProfiles)
	if result.Error != nil {
		log.Fatal("Failed to query users with profiles:", result.Error)
	}

	for _, user := range usersWithProfiles {
		fmt.Printf("User: %s (%s)\n", user.GetFullName(), user.Email)
		if user.Profile != nil {
			fmt.Printf("  Profile: %s at %s\n", user.Profile.JobTitle, user.Profile.Company)
			if user.Profile.Bio != "" {
				fmt.Printf("  Bio: %s\n", user.Profile.Bio)
			}
		} else {
			fmt.Println("  No profile")
		}
		fmt.Println()
	}

	// Example 8: Create a new user with profile
	fmt.Println("\n=== Creating New User ===")
	newUser := &model.User{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		FirstName: "New",
		LastName:  "User",
		Status:    model.UserStatusActive,
	}

	// Validate before creating
	if err := newUser.Validate(); err != nil {
		log.Fatal("User validation failed:", err)
	}

	// Create user
	result = db.Create(newUser)
	if result.Error != nil {
		log.Fatal("Failed to create user:", result.Error)
	}
	fmt.Printf("✅ Created user with ID: %d\n", newUser.ID)

	// Create profile for the new user
	newProfile := &model.Profile{
		UserID:   newUser.ID,
		Bio:      "A new user created via migration example",
		JobTitle: "Example User",
		Company:  "Example Corp",
		Country:  "USA",
		City:     "New York",
	}

	// Validate before creating
	if err := newProfile.Validate(); err != nil {
		log.Fatal("Profile validation failed:", err)
	}

	// Create profile
	result = db.Create(newProfile)
	if result.Error != nil {
		log.Fatal("Failed to create profile:", result.Error)
	}
	fmt.Printf("✅ Created profile with ID: %d\n", newProfile.ID)

	// Example 9: Test user methods
	fmt.Println("\n=== Testing User Methods ===")
	fmt.Printf("User full name: %s\n", newUser.GetFullName())
	fmt.Printf("User is active: %t\n", newUser.IsActive())
	fmt.Printf("User is pending: %t\n", newUser.IsPending())

	// Mark email as verified
	newUser.MarkEmailAsVerified()
	fmt.Printf("Email verified: %t\n", newUser.EmailVerified)

	// Record login
	newUser.RecordLogin()
	fmt.Printf("Login count: %d\n", newUser.LoginCount)

	// Update user in database
	result = db.Save(newUser)
	if result.Error != nil {
		log.Fatal("Failed to update user:", result.Error)
	}
	fmt.Println("✅ User updated successfully")

	// Example 10: Test profile methods
	fmt.Println("\n=== Testing Profile Methods ===")
	fmt.Printf("Profile display name: %s\n", newProfile.GetDisplayName())
	
	// Set date of birth and calculate age
	// birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	// newProfile.DateOfBirth = &birthDate
	// fmt.Printf("Profile age: %d\n", newProfile.GetAge())

	// Example 11: Pagination example
	fmt.Println("\n=== Pagination Example ===")
	params := &model.UserQueryParams{
		PaginationParams: model.PaginationParams{Page: 1, PageSize: 2},
		UserSort:         model.UserSort{},
		UserFilter:       model.UserFilter{Status: model.UserStatusActive},
	}
	params.SetDefaults()

	var paginatedUsers []model.User
	var total int64

	// Count total active users
	db.Model(&model.User{}).Where("status = ?", model.UserStatusActive).Count(&total)

	// Get paginated results
	query := db.Model(&model.User{}).Where("status = ?", model.UserStatusActive)
	result = query.Order(params.UserSort.GetOrderBy()).
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&paginatedUsers)

	if result.Error != nil {
		log.Fatal("Failed to get paginated users:", result.Error)
	}

	paginationResult := model.NewPaginationResult(&params.PaginationParams, total, paginatedUsers)
	fmt.Printf("Pagination result: Page %d of %d, Total: %d\n",
		paginationResult.Page, paginationResult.TotalPages, paginationResult.Total)

	for _, user := range paginatedUsers {
		fmt.Printf("- %s (%s)\n", user.GetFullName(), user.Email)
	}

	fmt.Println("\n✅ Migration example completed successfully!")

	// Note: In a real application, you might want to add a flag to reset the database
	// fmt.Println("\n=== Database Reset (Uncomment to test) ===")
	// err = migrationManager.ResetDatabase()
	// if err != nil {
	//     log.Fatal("Database reset failed:", err)
	// }
	// fmt.Println("✅ Database reset completed")
}