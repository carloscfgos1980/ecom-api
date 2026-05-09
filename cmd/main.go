package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/carloscfgos1980/ecom-api/internal/config"
	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/carloscfgos1980/ecom-api/internal/handlers"
	"github.com/carloscfgos1980/ecom-api/internal/middleware"

	_ "github.com/lib/pq"
)

func main() {
	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Connect to the database
	dbConn, err := sql.Open("postgres", cfg.DB_URL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer dbConn.Close()

	// Create a new database queries instance
	db := database.New(dbConn)

	cfg.DB = db

	// Initialize the Gin router
	var router *gin.Engine = gin.Default()

	// Set trusted proxies to nil to avoid warnings in Gin 1.7+
	router.SetTrustedProxies(nil)

	// Define a simple health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":  "Ecom API is running",
			"status":   "success",
			"database": "connected",
		})
	})
	// Register customer-related routes
	router.POST("/auth/register", handlers.CreateCustomerHandler(cfg))
	router.POST("/auth/login", handlers.LoginCustomerHandler(cfg))

	// Product routes
	router.GET("/products", handlers.GetProductsHandler(cfg))
	router.GET("/products/:productID", handlers.GetProductByIDHandler(cfg))

	// Register order-related routes with authentication middleware
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.AuthMiddleware(cfg))
	{
		apiGroup.POST("/orders", handlers.PlaceOrderHandler(cfg))
		apiGroup.GET("/orders", handlers.GetOrdersHandler(cfg))
		apiGroup.GET("/orders/:orderID", handlers.GetOrderByIDHandler(cfg))
	}

	// Start the server on the specified port
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
