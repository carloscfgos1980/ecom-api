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
	db        DB
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
	mux.HandleFunc("POST /auth/register", apiCfg.handlerUsersCreate)
	// Register the handler for logging in a customer
	mux.HandleFunc("POST /auth/login", apiCfg.handlerLogin)
	// Register the handler for retrieving all products
	mux.HandleFunc("GET /products", apiCfg.handlerProductsGet)
	// Register the handler for retrieving a product by ID
	mux.HandleFunc("GET /products/{productID}", apiCfg.handlerProductsGetByID)
	// Register the handler for placing an order
	mux.HandleFunc("POST /api/orders", apiCfg.handlerOrderCreate)
	// Register the handler for retrieving all orders and their items for a customer
	mux.HandleFunc("GET /api/orders", apiCfg.handlerOrdersItemsGet)
	// Register the handler for retrieving a specific order and its items by order ID
	mux.HandleFunc("GET /api/orders/{orderID}", apiCfg.handlerOrderItemsGetByOrderID)
	// Log the server start and the URL it's running on

	log.Printf("Server is running http://localhost:%s", apiCfg.port)
	// Listen and serve
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
