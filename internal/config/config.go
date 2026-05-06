package config

import (
	"errors"
	"os"

	"github.com/carloscfgos1980/ecom-api/internal/database"

	"github.com/joho/godotenv"
)

// Define custom error variables for missing configuration values
var (
	ErrMissingDatabaseURL = errors.New("missing database URL")
	ErrMissingPort        = errors.New("missing port")
	ErrMissingJWT         = errors.New("missing JWT secret")
)

// Config struct to hold application configuration values
type Config struct {
	DB        *database.Queries
	DB_URL    string
	Port      string
	JWTSecret string
}

// LoadConfig loads configuration values from environment variables and returns a Config struct
func LoadConfig() (*Config, error) {
	// Try common .env locations (project root and cmd/ execution path).
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load("../.env.local")

	// Load required configuration values from environment variables
	DB_URL := os.Getenv("DB_URL")
	if DB_URL == "" {
		return nil, ErrMissingDatabaseURL
	}

	Port := os.Getenv("PORT")
	if Port == "" {
		return nil, ErrMissingPort
	}

	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		return nil, ErrMissingJWT
	}

	// Return the configuration struct with the loaded values
	return &Config{
		DB_URL:    DB_URL,
		Port:      Port,
		JWTSecret: JWTSecret,
	}, nil
}
