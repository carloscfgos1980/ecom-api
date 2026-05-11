package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// apiConfig holds the dependencies for the API handlers.
type apiConfig struct {
	db        *database.Queries
	jwtSecret string
	port      string
}

func main() {
	// Load environment variables from .env file
	godotenv.Load()
	// Get configuration from environment variables
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	// Get the port from environment variables, default to 8080 if not set
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}
	// Get the JWT secret from environment variables
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}
	// Connect to the database
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer dbConn.Close()

	// database queries variable
	dbQueries := database.New(dbConn)
	// variable for the apiConfig struct
	apiCfg := apiConfig{
		db:        dbQueries,
		port:      port,
		jwtSecret: jwtSecret,
	}
	// Set up the HTTP server and routes
	mux := http.NewServeMux()
	// Start the HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	//health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
		"message":  "Ecom API is running",
		"status":   "success",
		"database": "connected",
		"version":  "1.4.0"
		}`))
	})
	// Register the handler for creating a new customer
	mux.HandleFunc("/auth/register", apiCfg.handlerUsersCreate)
	// Register the handler for logging in a customer
	mux.HandleFunc("/auth/login", apiCfg.handlerLogin)

	log.Printf("Server is running http://localhost:%s", apiCfg.port)
	// Listen and serve
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
