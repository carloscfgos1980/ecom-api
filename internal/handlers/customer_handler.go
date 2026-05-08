package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/carloscfgos1980/ecom-api/internal/config"
	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/carloscfgos1980/ecom-api/internal/utils"
	"github.com/google/uuid"
)

// structs and handler for creating a new customer in the system
type Customer struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
}

// CustomerRequest is the struct for the request body when creating a new customer
type CustomerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CreateCustomerHandler is the handler for creating a new customer
func CreateCustomerHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Bind the JSON request body to the CustomerRequest struct and validate it
		var req CustomerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		// Validate email format
		err := utils.IsValidEmail(req.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Validate the password strength
		err = utils.IsStrongPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Hash the password before storing it in the database
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Create the customer in the database using the provided configuration and request data
		customer, err := cfg.DB.CreateCustomer(c, database.CreateCustomerParams{
			Email:    req.Email,
			Password: hashedPassword,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Prepare the response with the created customer's information, excluding the password
		response := Customer{
			ID:        customer.ID,
			CreatedAt: customer.CreatedAt,
			UpdatedAt: customer.UpdatedAt,
			Email:     customer.Email,
		}
		// Return the created customer information in the response with a 201 Created status
		c.JSON(http.StatusCreated, response)
	}
}

// LoginCustomerHandler is the handler for logging in a customer and generating a JWT token
func LoginCustomerHandler(cfg *config.Config) gin.HandlerFunc {
	// Define a response struct that includes the customer information and the generated token
	type response struct {
		Customer
		Token string `json:"token"`
	}
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Bind the JSON request body to the CustomerRequest struct
		var req CustomerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Validate email format
		if err := utils.IsValidEmail(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Retrieve the customer from the database using the provided email
		customer, err := cfg.DB.GetCustomerByEmail(c, req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		// Check if the provided password matches the stored hashed password
		match, err := utils.CheckPasswordHash(req.Password, customer.Password)
		if err != nil || !match {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		// Generate a JWT token for the authenticated customer
		token, err := utils.MakeJWT(
			customer.ID,
			cfg.JWTSecret,
			24*7*time.Hour,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}
		// Prepare the response with the authenticated customer's information and the generated token
		response := response{
			Customer: Customer{
				ID:        customer.ID,
				CreatedAt: customer.CreatedAt,
				UpdatedAt: customer.UpdatedAt,
				Email:     customer.Email,
			},
			Token: token,
		}
		// Send the response back to the client with a 200 OK status
		c.JSON(http.StatusOK, response)
	}
}
