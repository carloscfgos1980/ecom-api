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
	DB                  *database.Queries
	DB_URL              string
	Port                string
	JWTSecret           string
	XLS_FILE_PATH_READ  string
	XLS_FILE_PATH_WRITE string
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

	XLS_FILE_PATH_READ := os.Getenv("XLS_FILE_PATH_READ")
	if XLS_FILE_PATH_READ == "" {
		XLS_FILE_PATH_READ = "data-exel/products_start.xls"
	}

	XLS_FILE_PATH_WRITE := os.Getenv("XLS_FILE_PATH_WRITE")
	if XLS_FILE_PATH_WRITE == "" {
		XLS_FILE_PATH_WRITE = "data-exel/products_export.xlsx"
	}
	// Return the configuration struct with the loaded values
	return &Config{
		DB_URL:              DB_URL,
		Port:                Port,
		JWTSecret:           JWTSecret,
		XLS_FILE_PATH_READ:  XLS_FILE_PATH_READ,
		XLS_FILE_PATH_WRITE: XLS_FILE_PATH_WRITE,
	}, nil
}
