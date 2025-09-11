package main

import (
	"log"
	"os"

	"go-web-starter/internal/app"
	"go-web-starter/internal/config"
)

// @title Go Web Starter API
// @version 1.0
// @description 这是 Go Web Starter 项目的 API 文档。
// @contact.name API Support
// @contact.url https://example.com
// @contact.email support@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @schemes http

func main() {
	log.Println("Starting Go Web Starter application...")

	// Load configuration
	cfg, err := config.Load("configs/config.dev.yaml")
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create application instance
	application, err := app.New(cfg)
	if err != nil {
		log.Fatal("Failed to create application:", err)
	}

	// Run application
	if err := application.Run(); err != nil {
		log.Fatal("Application failed:", err)
	}

	// Exit gracefully
	os.Exit(0)
}
