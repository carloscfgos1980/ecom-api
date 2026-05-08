package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/carloscfgos1980/ecom-api/internal/config"
)

// Product is the struct representing a product in the system
type Product struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
}

// GetProductsHandler is the handler for retrieving a list of products
func GetProductsHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Retrieve the list of products from the database using the provided configuration
		products, err := cfg.DB.GetProducts(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve products"})
			return
		}
		// Prepare the response by converting the products from the database format to the API response format
		response := []Product{}
		for _, p := range products {
			response = append(response, Product{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
		// Send the list of products back to the client with a 200 OK status
		c.JSON(http.StatusOK, response)
	}
}
