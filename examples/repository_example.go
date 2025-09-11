package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-web-starter/internal/config"
	"go-web-starter/internal/domain/model"
	"go-web-starter/internal/infrastructure/database"
	"go-web-starter/internal/infrastructure/logger"
	"go-web-starter/internal/infrastructure/repository"
)

func main() {
	fmt.Println("=== Repository Layer Example ===")

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

	// Run migrations
	migrationManager := database.NewMigrationManager(db.DB, appLogger)
	err = migrationManager.MigrateAll()
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	// Create repository manager
	repoManager := repository.NewRepositoryManager(db.DB, appLogger)
	defer repoManager.Close()

	ctx := context.Background()

	// Example 1: User Repository Operations
	fmt.Println("\n=== User Repository Operations ===")

	userRepo := repoManager.User()

	// Create a new user
	newUser := &model.User{
		Username:  "repouser",
		Email:     "repo@example.com",
		Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		FirstName: "Repository",
		LastName:  "User",
		Status:    model.UserStatusActive,
	}

	// Validate and create user
	if err := newUser.Validate(); err != nil {
		log.Fatal("User validation failed:", err)
	}

	err = userRepo.Create(ctx, newUser)
	if err != nil {
		log.Fatal("Failed to create user:", err)
	}
	fmt.Printf("✅ Created user with ID: %d\n", newUser.ID)

	// Get user by ID
	foundUser, err := userRepo.GetByID(ctx, newUser.ID)
	if err != nil {
		log.Fatal("Failed to get user by ID:", err)
	}
	fmt.Printf("✅ Found user by ID: %s (%s)\n", foundUser.Username, foundUser.Email)

	// Get user by username
	foundUser, err = userRepo.GetByUsername(ctx, newUser.Username)
	if err != nil {
		log.Fatal("Failed to get user by username:", err)
	}
	fmt.Printf("✅ Found user by username: %s\n", foundUser.GetFullName())

	// Get user by email
	foundUser, err = userRepo.GetByEmail(ctx, newUser.Email)
	if err != nil {
		log.Fatal("Failed to get user by email:", err)
	}
	fmt.Printf("✅ Found user by email: %s\n", foundUser.Email)

	// Update user
	foundUser.FirstName = "Updated"
	foundUser.LastName = "Name"
	err = userRepo.Update(ctx, foundUser)
	if err != nil {
		log.Fatal("Failed to update user:", err)
	}
	fmt.Printf("✅ Updated user: %s\n", foundUser.GetFullName())

	// Update last login
	err = userRepo.UpdateLastLogin(ctx, foundUser.ID)
	if err != nil {
		log.Fatal("Failed to update last login:", err)
	}
	fmt.Println("✅ Updated user last login")

	// Update email verification
	err = userRepo.UpdateEmailVerification(ctx, foundUser.ID, true)
	if err != nil {
		log.Fatal("Failed to update email verification:", err)
	}
	fmt.Println("✅ Updated user email verification")

	// Update status
	err = userRepo.UpdateStatus(ctx, foundUser.ID, model.UserStatusSuspended)
	if err != nil {
		log.Fatal("Failed to update user status:", err)
	}
	fmt.Println("✅ Updated user status to suspended")

	// Example 2: Profile Repository Operations
	fmt.Println("\n=== Profile Repository Operations ===")

	profileRepo := repoManager.Profile()

	// Create a profile for the user
	newProfile := &model.Profile{
		UserID:   foundUser.ID,
		Bio:      "A user created via repository example",
		Phone:    "+1234567890",
		Company:  "Repository Corp",
		JobTitle: "Repository Manager",
		Country:  "USA",
		City:     "New York",
		Website:  "https://example.com",
	}

	// Validate and create profile
	if err := newProfile.Validate(); err != nil {
		log.Fatal("Profile validation failed:", err)
	}

	err = profileRepo.Create(ctx, newProfile)
	if err != nil {
		log.Fatal("Failed to create profile:", err)
	}
	fmt.Printf("✅ Created profile with ID: %d\n", newProfile.ID)

	// Get profile by user ID
	foundProfile, err := profileRepo.GetByUserID(ctx, foundUser.ID)
	if err != nil {
		log.Fatal("Failed to get profile by user ID:", err)
	}
	fmt.Printf("✅ Found profile by user ID: %s at %s\n", foundProfile.JobTitle, foundProfile.Company)

	// Get profile with user
	profileWithUser, err := profileRepo.GetWithUser(ctx, foundProfile.ID)
	if err != nil {
		log.Fatal("Failed to get profile with user:", err)
	}
	fmt.Printf("✅ Found profile with user: %s\n", profileWithUser.User.GetFullName())

	// Update avatar
	err = profileRepo.UpdateAvatar(ctx, foundUser.ID, "https://example.com/avatar.jpg")
	if err != nil {
		log.Fatal("Failed to update avatar:", err)
	}
	fmt.Println("✅ Updated profile avatar")

	// Example 3: User with Profile Operations
	fmt.Println("\n=== User with Profile Operations ===")

	// Get user with profile
	userWithProfile, err := userRepo.GetWithProfile(ctx, foundUser.ID)
	if err != nil {
		log.Fatal("Failed to get user with profile:", err)
	}
	fmt.Printf("✅ Found user with profile: %s\n", userWithProfile.GetFullName())
	if userWithProfile.Profile != nil {
		fmt.Printf("   Profile: %s at %s\n", userWithProfile.Profile.JobTitle, userWithProfile.Profile.Company)
		fmt.Printf("   Location: %s, %s\n", userWithProfile.Profile.City, userWithProfile.Profile.Country)
	}

	// Example 4: List Operations with Pagination
	fmt.Println("\n=== List Operations with Pagination ===")

	// List users with pagination
	userParams := &model.UserQueryParams{
		PaginationParams: model.PaginationParams{Page: 1, PageSize: 5},
		UserSort:         model.UserSort{},
		UserFilter:       model.UserFilter{Status: model.UserStatusActive},
	}
	userParams.SetDefaults()

	userResult, err := userRepo.ListWithProfiles(ctx, userParams)
	if err != nil {
		log.Fatal("Failed to list users with profiles:", err)
	}
	fmt.Printf("✅ Listed users: Page %d of %d, Total: %d\n",
		userResult.Page, userResult.TotalPages, userResult.Total)

	if users, ok := userResult.Data.([]*model.User); ok {
		for _, user := range users {
			fmt.Printf("   - %s (%s) - Status: %s\n", user.GetFullName(), user.Email, user.Status)
		}
	}

	// List profiles with pagination
	profileParams := &model.QueryParams{
		PaginationParams: model.PaginationParams{Page: 1, PageSize: 5},
		SortParams:       model.SortParams{SortBy: "created_at", SortOrder: "desc"},
	}
	profileParams.SetDefaults()

	profileResult, err := profileRepo.List(ctx, profileParams)
	if err != nil {
		log.Fatal("Failed to list profiles:", err)
	}
	fmt.Printf("✅ Listed profiles: Page %d of %d, Total: %d\n",
		profileResult.Page, profileResult.TotalPages, profileResult.Total)

	// Example 5: Search Operations
	fmt.Println("\n=== Search Operations ===")

	// Search users
	searchParams := &model.PaginationParams{Page: 1, PageSize: 10}
	searchParams.SetDefaults()

	searchResult, err := userRepo.SearchUsers(ctx, "repo", searchParams)
	if err != nil {
		log.Fatal("Failed to search users:", err)
	}
	fmt.Printf("✅ Search results: %d users found\n", searchResult.Total)

	// Search profiles by location
	locationResult, err := profileRepo.SearchByLocation(ctx, "USA", "New York", searchParams)
	if err != nil {
		log.Fatal("Failed to search profiles by location:", err)
	}
	fmt.Printf("✅ Location search results: %d profiles found in New York, USA\n", locationResult.Total)

	// Search profiles by company
	companyResult, err := profileRepo.GetByCompany(ctx, "Repository", searchParams)
	if err != nil {
		log.Fatal("Failed to search profiles by company:", err)
	}
	fmt.Printf("✅ Company search results: %d profiles found for 'Repository'\n", companyResult.Total)

	// Example 6: Transaction Operations
	fmt.Println("\n=== Transaction Operations ===")

	err = repoManager.Transaction(ctx, func(txRepoManager repository.RepositoryManager) error {
		// Create user and profile in a transaction
		txUser := &model.User{
			Username:  "txuser",
			Email:     "tx@example.com",
			Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			FirstName: "Transaction",
			LastName:  "User",
			Status:    model.UserStatusActive,
		}

		if err := txUser.Validate(); err != nil {
			return err
		}

		// Create user within transaction
		if err := txRepoManager.User().Create(ctx, txUser); err != nil {
			return err
		}

		// Create profile within transaction
		txProfile := &model.Profile{
			UserID:   txUser.ID,
			Bio:      "Created in transaction",
			Company:  "Transaction Corp",
			JobTitle: "Transaction Manager",
		}

		if err := txProfile.Validate(); err != nil {
			return err
		}

		if err := txRepoManager.Profile().Create(ctx, txProfile); err != nil {
			return err
		}

		fmt.Printf("✅ Created user and profile in transaction: %s\n", txUser.GetFullName())
		return nil
	})

	if err != nil {
		log.Fatal("Transaction failed:", err)
	}
	fmt.Println("✅ Transaction completed successfully")

	// Example 7: Count and Exists Operations
	fmt.Println("\n=== Count and Exists Operations ===")

	// Count users
	userCount, err := userRepo.Count(ctx, map[string]interface{}{"status": model.UserStatusActive})
	if err != nil {
		log.Fatal("Failed to count users:", err)
	}
	fmt.Printf("✅ Active users count: %d\n", userCount)

	// Check if user exists
	exists, err := userRepo.Exists(ctx, foundUser.ID)
	if err != nil {
		log.Fatal("Failed to check user existence:", err)
	}
	fmt.Printf("✅ User exists: %t\n", exists)

	// Example 8: Get Active Users
	fmt.Println("\n=== Get Active Users ===")

	activeParams := &model.PaginationParams{Page: 1, PageSize: 10}
	activeParams.SetDefaults()

	activeResult, err := userRepo.GetActiveUsers(ctx, activeParams)
	if err != nil {
		log.Fatal("Failed to get active users:", err)
	}
	fmt.Printf("✅ Active users: %d found\n", activeResult.Total)

	// Example 9: Soft Delete and Hard Delete
	fmt.Println("\n=== Delete Operations ===")

	// Create a user to delete
	deleteUser := &model.User{
		Username: "deleteuser",
		Email:    "delete@example.com",
		Password: "password123",
		Status:   model.UserStatusActive,
	}

	if err := deleteUser.Validate(); err != nil {
		log.Fatal("Delete user validation failed:", err)
	}

	err = userRepo.Create(ctx, deleteUser)
	if err != nil {
		log.Fatal("Failed to create delete user:", err)
	}

	// Soft delete
	err = userRepo.Delete(ctx, deleteUser.ID)
	if err != nil {
		log.Fatal("Failed to soft delete user:", err)
	}
	fmt.Printf("✅ Soft deleted user with ID: %d\n", deleteUser.ID)

	// Try to get deleted user (should fail)
	_, err = userRepo.GetByID(ctx, deleteUser.ID)
	if err != nil {
		fmt.Println("✅ Soft deleted user not found (expected)")
	}

	// Hard delete
	err = userRepo.HardDelete(ctx, deleteUser.ID)
	if err != nil {
		log.Fatal("Failed to hard delete user:", err)
	}
	fmt.Printf("✅ Hard deleted user with ID: %d\n", deleteUser.ID)

	fmt.Println("\n✅ Repository example completed successfully!")

	// Example 10: Error Handling
	fmt.Println("\n=== Error Handling Examples ===")

	// Try to get non-existent user
	_, err = userRepo.GetByID(ctx, 99999)
	if err != nil {
		fmt.Printf("✅ Expected error for non-existent user: %v\n", err)
		if repository.IsNotFoundError(err) {
			fmt.Println("✅ Correctly identified as not found error")
		}
	}

	// Try to create duplicate user
	duplicateUser := &model.User{
		Username: foundUser.Username, // Same username
		Email:    "different@example.com",
		Password: "password123",
		Status:   model.UserStatusActive,
	}

	err = userRepo.Create(ctx, duplicateUser)
	if err != nil {
		fmt.Printf("✅ Expected error for duplicate username: %v\n", err)
		if repository.IsAlreadyExistsError(err) {
			fmt.Println("✅ Correctly identified as already exists error")
		}
	}

	fmt.Println("\n🎉 All repository examples completed successfully!")
}