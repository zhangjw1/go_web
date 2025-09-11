//go:build examples
// +build examples

package main

import (
	"fmt"
	"log"
	"os"

	"go-web-starter/internal/config"
)

func main() {
	// Example 1: Load default config
	fmt.Println("=== Loading default config ===")
	cfg, err := config.Load("")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	fmt.Printf("Server Port: %s\n", cfg.Server.Port)
	fmt.Printf("Database Host: %s\n", cfg.Database.Host)
	fmt.Printf("Redis Host: %s\n", cfg.Redis.Host)

	// Example 2: Load specific config file
	fmt.Println("\n=== Loading dev config ===")
	devCfg, err := config.Load("configs/config.dev.yaml")
	if err != nil {
		log.Fatal("Failed to load dev config:", err)
	}
	fmt.Printf("Server Mode: %s\n", devCfg.Server.Mode)
	fmt.Printf("Logger Level: %s\n", devCfg.Logger.Level)

	// Example 3: Override with environment variables
	fmt.Println("\n=== Testing environment variable override ===")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DATABASE_HOST", "remote-db")

	envCfg, err := config.Load("")
	if err != nil {
		log.Fatal("Failed to load config with env vars:", err)
	}
	fmt.Printf("Server Port (from env): %s\n", envCfg.Server.Port)
	fmt.Printf("Database Host (from env): %s\n", envCfg.Database.Host)

	// Clean up
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DATABASE_HOST")
}
