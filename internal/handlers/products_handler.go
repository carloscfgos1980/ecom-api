package handlers

import (
	"context"
	"net/http"
	"strconv"

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

// GetProductByIDHandler is the handler for retrieving a single product by its ID
func GetProductByIDHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the product ID from the URL parameters
		id := c.Param("productID")
		//convert id to int64
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
			return
		}
		// Retrieve the product from the database using the provided configuration and product ID
		product, err := cfg.DB.GetProductByID(context.Background(), idInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve product"})
			return
		}

		// Prepare the response by converting the product from the database format to the API response format
		response := Product{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		}
		// Send the product back to the client with a 200 OK status
		c.JSON(http.StatusOK, response)
	}
}
