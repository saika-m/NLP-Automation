package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"tashi-backend/internal/api"
	"tashi-backend/internal/config"
	"tashi-backend/internal/database"
	"tashi-backend/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Create required directories
	createDirectories(cfg)

	// Initialize database
	db, err := database.Initialize(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize core services
	aiService := services.NewAIService(cfg.OpenAIAPIKey)
	systemService := services.NewSystemService()
	confirmationService := services.NewConfirmationService()

	// Gather system information on startup
	ctx := context.Background()
	if _, err := systemService.GatherSystemInfo(ctx); err != nil {
		log.Printf("Warning: Failed to gather system info: %v", err)
	}

	// Initialize specialized services
	scriptService := services.NewScriptService(cfg)
	compilerService := services.NewCompilerService(cfg)
	commandService := services.NewCommandService(aiService)
	officeService := services.NewOfficeService(systemService, aiService)
	taskService := services.NewTaskService(db, scriptService, aiService, compilerService)

	// Initialize and start server
	server := api.NewServer(cfg, taskService, commandService, systemService, officeService, confirmationService)

	log.Printf("Starting Tashi Office Super Tool backend server on port %s", cfg.Port)
	log.Printf("System: %s %s", systemService.GetCachedSystemInfo().OS.Name, systemService.GetCachedSystemInfo().OS.Version)
	log.Printf("Office capabilities detected...")

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createDirectories(cfg *config.Config) {
	dirs := []string{cfg.UploadDir, cfg.OutputDir, cfg.TempDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Also ensure the database directory exists
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory %s: %v", dbDir, err)
	}
}
