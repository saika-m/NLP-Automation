package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Server Configuration
	Port    string
	GinMode string

	// OpenAI Configuration
	OpenAIAPIKey string

	// Database Configuration
	DBPath string

	// File Storage
	UploadDir string
	OutputDir string
	TempDir   string

	// Security
	JWTSecret   string
	CORSOrigins []string

	// Script Generation Settings
	MaxScriptSize      int64
	AllowedExtensions  []string
	CompilationTimeout int

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   int
}

func Load() *Config {
	return &Config{
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "debug"),

		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		DBPath: getEnv("DB_PATH", "./nlp-automation.db"),

		UploadDir: getEnv("UPLOAD_DIR", "./uploads"),
		OutputDir: getEnv("OUTPUT_DIR", "./outputs"),
		TempDir:   getEnv("TEMP_DIR", "./temp"),

		JWTSecret:   getEnv("JWT_SECRET", "default-secret-change-in-production"),
		CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ","),

		MaxScriptSize:      getEnvInt64("MAX_SCRIPT_SIZE", 10485760), // 10MB
		AllowedExtensions:  strings.Split(getEnv("ALLOWED_EXTENSIONS", ".ps1,.sh,.py,.exe,.bat"), ","),
		CompilationTimeout: getEnvInt("COMPILATION_TIMEOUT", 300), // 5 minutes

		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvInt("RATE_LIMIT_WINDOW", 3600), // 1 hour
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
