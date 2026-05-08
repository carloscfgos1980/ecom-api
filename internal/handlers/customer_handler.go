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
