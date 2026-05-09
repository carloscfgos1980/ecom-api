package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/carloscfgos1980/ecom-api/internal/config"
	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/gin-gonic/gin"
)

// Decimal2 marshals to JSON number with exactly two decimal places.
type Decimal2 float64

// MarshalJSON implements the json.Marshaler interface for Decimal2.
func (d Decimal2) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.2f", d)), nil
}

// orderItem represents an item in an order
type orderItem struct {
	ProductID int64 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

// OrderResponse represents the response for an order
type OrderResponse struct {
	ID         int64           `json:"id"`
	CustomerID uuid.UUID       `json:"customerId"`
	CreatedAt  string          `json:"createdAt"`
	Total      Decimal2        `json:"total"`
	Items      []itemsResponse `json:"items"`
}

// itemsResponse represents the response for an order item
type itemsResponse struct {
	ProductID   int64    `json:"productId"`
	ProductName string   `json:"productName"`
	Quantity    int32    `json:"quantity"`
	Price       Decimal2 `json:"price"`
	Subtotal    Decimal2 `json:"subtotal"`
}

// PlaceOrderHandler is the handler for placing an order
func PlaceOrderHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Get the customer ID from the Gin context (set by the authentication middleware)
		customerID, exists := c.Get("customerID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "customer ID not found in context"})
			return
		}
		// Check if the customer is resgister
		_, err := cfg.DB.GetCustomerByID(context.Background(), customerID.(uuid.UUID))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "customer not found"})
			return
		}
		// create a new order in the database with the customer ID and the current timestamp
		order, err := cfg.DB.CreateOrder(context.Background(), customerID.(uuid.UUID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
			return
		}
		// Bind the JSON body to a slice of orderItem structs
		var items []orderItem
		if err := c.ShouldBindJSON(&items); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if len(items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order must include at least one item"})
			return
		}
		// Loop through the order items, check if the product exists and has enough stock, create order items in the database, and calculate the total price of the order
		total := 0.00
		responseItems := make([]itemsResponse, 0, len(items))

		// look for the product if exists
		for _, item := range items {
			// check if the product exists
			product, err := cfg.DB.GetProductByID(context.Background(), item.ProductID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("product: %s not found", product.Name)})
				return
			}
			// check if the product has enough stock
			if product.Quantity < item.Quantity {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("not enough stock for product: %s", product.Name)})
				return
			}

			// create order item
			err = cfg.DB.CreateOrderItem(context.Background(), database.CreateOrderItemParams{
				OrderID:   order.OrderID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     product.Price,
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create order item for product: %s", product.Name)})
				return
			}
			// calculate subtotal for the item
			unitPrice, err := strconv.ParseFloat(product.Price, 64)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("invalid price format for product: %s", product.Name)})
				return
			}
			subtotal := unitPrice * float64(item.Quantity)
			// calculate total
			total += subtotal

			// update product stock
			err = cfg.DB.UpdateProductStock(context.Background(), database.UpdateProductStockParams{
				ID:       item.ProductID,
				Quantity: product.Quantity - item.Quantity,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update product stock for product: %s", product.Name)})
				return
			}
			// prepare the response item
			responseItems = append(responseItems, itemsResponse{
				ProductID:   item.ProductID,
				ProductName: product.Name,
				Quantity:    item.Quantity,
				Price:       Decimal2(unitPrice),
				Subtotal:    Decimal2(subtotal),
			})

		}
		// prepare the order response
		response := OrderResponse{
			ID:         order.OrderID,
			CustomerID: order.CustomerID,
			CreatedAt:  order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			Total:      Decimal2(total),
			Items:      responseItems,
		}
		// send the order response back to the client with a 200 OK status
		c.JSON(http.StatusOK, response)

	}
}
