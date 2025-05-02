package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"starter-app/pkg/env"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
	Cache    CacheConfig    `json:"cache"` // Add cache config
}

// ServerConfig contains server related configuration
type ServerConfig struct {
	Port string `json:"port"`
	Env  string `json:"env"`
}

// DatabaseConfig contains database related configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// AuthConfig contains authentication related configuration
type AuthConfig struct {
	JWTSecret     string        `json:"jwt_secret"`
	TokenExpiry   time.Duration `json:"token_expiry"`
	RefreshExpiry time.Duration `json:"refresh_expiry"`
}

// CacheConfig contains cache related configuration
type CacheConfig struct {
	Driver string        `json:"driver"`
	TTL    time.Duration `json:"ttl"`
}

// Load loads the configuration from environment variables or config file
func Load() (*Config, error) {
	// Load .env file
	rootDir := findRootDir()
	envPath := filepath.Join(rootDir, ".env")

	if err := godotenv.Load(envPath); err != nil {
		// If .env file doesn't exist or can't be read, log the error but continue
		fmt.Printf("Warning: .env file not found or cannot be read: %v\n", err)
		fmt.Println("Using default values or environment variables if set")
	} else {
		fmt.Println("Environment variables loaded from .env file")
	}

	// Parse token expiry durations with error handling
	tokenExpiry := env.GetDuration("TOKEN_EXPIRY", 24*time.Hour)
	refreshExpiry := env.GetDuration("REFRESH_EXPIRY", 7*24*time.Hour)
	cacheTTL := env.GetDuration("CACHE_TTL", 5*time.Minute)

	// Default configuration
	config := &Config{
		Server: ServerConfig{
			Port: env.Get("PORT", "8080"),
			Env:  env.Get("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     env.Get("DB_HOST", "localhost"),
			Port:     env.Get("DB_PORT", "5432"),
			User:     env.Get("DB_USER", "postgres"),
			Password: env.Get("DB_PASSWORD", "postgres"),
			Name:     env.Get("DB_NAME", "starter_db"),
		},
		Auth: AuthConfig{
			JWTSecret:     env.Get("JWT_SECRET", "your-secret-key-change-in-production"),
			TokenExpiry:   tokenExpiry,
			RefreshExpiry: refreshExpiry,
		},
		Cache: CacheConfig{
			Driver: env.Get("CACHE_DRIVER", "memory"),
			TTL:    cacheTTL,
		},
	}

	// Try to load from config.json if it exists
	configJsonPath := filepath.Join(rootDir, "config.json")
	if configFile, err := os.Open(configJsonPath); err == nil {
		defer configFile.Close()
		if err = json.NewDecoder(configFile).Decode(config); err != nil {
			return nil, fmt.Errorf("error parsing config.json: %w", err)
		}
		fmt.Println("Configuration loaded from config.json")
	}

	return config, nil
}

// findRootDir attempts to find the root directory of the project
func findRootDir() string {
	// Try to get from environment variable first
	if rootDir := os.Getenv("APP_ROOT_DIR"); rootDir != "" {
		return rootDir
	}

	// Default to current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Warning: Could not determine current directory: %v\n", err)
		return "."
	}

	// Check if we're running from the root or from a subdirectory
	// If common project files/directories exist in the current directory, assume it's the root
	commonFiles := []string{"go.mod", ".env", "cmd"}
	for _, file := range commonFiles {
		if _, err := os.Stat(filepath.Join(cwd, file)); err == nil {
			return cwd
		}
	}

	// If not found, try to go up one directory
	parentDir := filepath.Dir(cwd)
	for _, file := range commonFiles {
		if _, err := os.Stat(filepath.Join(parentDir, file)); err == nil {
			return parentDir
		}
	}

	// If still not found, just return current directory
	return cwd
}

// parseDuration parses a duration string with a fallback value if parsing fails
func parseDuration(durationStr string, fallback time.Duration) (time.Duration, error) {
	if durationStr == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fallback, err
	}

	return duration, nil
}
