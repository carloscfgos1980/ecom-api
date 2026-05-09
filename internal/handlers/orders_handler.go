package handlers

import (
	"context"
	"database/sql"
	"errors"
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

// GetOrdersHandler is the handler for getting orders
func GetOrdersHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Get the customer ID from the Gin context (set by the authentication middleware)
		customerID, exists := c.Get("customerID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "customer ID not found in context"})
			return
		}
		// Check if the customer is registered
		_, err := cfg.DB.GetCustomerByID(context.Background(), customerID.(uuid.UUID))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "customer not found"})
			return
		}
		// Get the role query parameter to determine if the user is an admin or a customer
		role := c.Query("role")
		// If the role is admin, return all orders. If the role is customer, return only the orders for the authenticated customer. If the role is not provided or is invalid, return a bad request error.
		switch role {
		case "admin":
			// get all orders from the database
			orders, err := cfg.DB.GetOrders(context.Background())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get orders"})
				return
			}
			ordersResponse := make([]OrderResponse, 0, len(orders))
			// Loop through the orders and get the order items and product details for each order to prepare the response
			for _, order := range orders {
				// get order items for each order
				orderItems, err := cfg.DB.GetOrderItemsByOrderID(context.Background(), order.OrderID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get order items for order ID: %d", order.OrderID)})
					return
				}
				// format the response with total as decimal with 2 places
				total := 0.00
				responseItems := make([]itemsResponse, 0, len(orderItems))
				// Loop through the order items and get the product details for each item to prepare the response
				for _, item := range orderItems {
					// get product details for each order item
					product, err := cfg.DB.GetProductByID(context.Background(), item.ProductID)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get product details for product ID: %d", item.ProductID)})
						return
					}
					// calculate subtotal for the item
					unitPrice, err := strconv.ParseFloat(item.Price, 64)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("invalid price format for product ID: %d", item.ProductID)})
						return
					}
					subtotal := unitPrice * float64(item.Quantity)
					// Item response with product details and subtotal
					responseItems = append(responseItems, itemsResponse{
						ProductID:   item.ProductID,
						ProductName: product.Name,
						Quantity:    item.Quantity,
						Price:       Decimal2(unitPrice),
						Subtotal:    Decimal2(subtotal),
					})
					// calculate total for the order
					total += subtotal
				}
				// Only include orders that have items in the response
				if len(responseItems) > 0 {
					ordersResponse = append(ordersResponse, OrderResponse{
						ID:         order.OrderID,
						CustomerID: order.CustomerID,
						CreatedAt:  order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
						Total:      Decimal2(total),
						Items:      responseItems,
					})
				}
			}
			// send the orders response back to the client with a 200 OK status
			c.JSON(http.StatusOK, ordersResponse)
		// If the role is customer, return only the orders for the authenticated customer
		case "customer":
			// get orders for the authenticated customer from the database
			orders, err := cfg.DB.GetOrdersByCustomerID(context.Background(), customerID.(uuid.UUID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get orders"})
				return
			}
			// format the response with total as decimal with 2 places
			ordersResponse := make([]OrderResponse, 0, len(orders))
			for _, order := range orders {
				// get order items for each order
				orderItems, err := cfg.DB.GetOrderItemsByOrderID(context.Background(), order.OrderID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get order items for order ID: %d", order.OrderID)})
					return
				}
				total := 0.00
				responseItems := make([]itemsResponse, 0, len(orderItems))
				for _, item := range orderItems {
					// get product details for each order item
					product, err := cfg.DB.GetProductByID(context.Background(), item.ProductID)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get product details for product ID: %d", item.ProductID)})
						return
					}
					unitPrice, err := strconv.ParseFloat(item.Price, 64)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("invalid price format for product ID: %d", item.ProductID)})
						return
					}
					subtotal := unitPrice * float64(item.Quantity)
					responseItems = append(responseItems, itemsResponse{
						ProductID:   item.ProductID,
						ProductName: product.Name,
						Quantity:    item.Quantity,
						Price:       Decimal2(unitPrice),
						Subtotal:    Decimal2(subtotal),
					})
					total += subtotal
				}
				if len(responseItems) > 0 {
					ordersResponse = append(ordersResponse, OrderResponse{
						ID:         order.OrderID,
						CustomerID: order.CustomerID,
						CreatedAt:  order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
						Total:      Decimal2(total),
						Items:      responseItems,
					})
				}
			}
			c.JSON(http.StatusOK, ordersResponse)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role query parameter"})
			return
		}
	}
}

func GetOrderByIDHandler(cfg *config.Config) gin.HandlerFunc {
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
		// Get the order ID from the URL parameter
		orderIDParam := c.Param("orderID")
		orderID, err := strconv.ParseInt(orderIDParam, 10, 64)
		if err != nil || orderID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
			return
		}
		// Get the role query parameter to determine if the user is an admin or a customer
		role := c.Query("role")
		switch role {
		case "admin":
			// get the order by ID from the database
			order, err := cfg.DB.GetOrderByID(context.Background(), orderID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
				return
			}
			// get order items for the order
			orderItems, err := cfg.DB.GetOrderItemsByOrderID(context.Background(), order.OrderID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get order items for order ID: %d", order.OrderID)})
				return
			}
			total := 0.00
			responseItems := make([]itemsResponse, 0, len(orderItems))
			for _, item := range orderItems {
				// get product details for each order item
				product, err := cfg.DB.GetProductByID(context.Background(), item.ProductID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get product details for product ID: %d", item.ProductID)})
					return
				}
				unitPrice, err := strconv.ParseFloat(item.Price, 64)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("invalid price format for product ID: %d", item.ProductID)})
					return
				}
				subtotal := unitPrice * float64(item.Quantity)
				responseItems = append(responseItems, itemsResponse{
					ProductID:   item.ProductID,
					ProductName: product.Name,
					Quantity:    item.Quantity,
					Price:       Decimal2(unitPrice),
					Subtotal:    Decimal2(subtotal),
				})
				total += subtotal
			}
			response := OrderResponse{
				ID:         order.OrderID,
				CustomerID: order.CustomerID,
				CreatedAt:  order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
				Total:      Decimal2(total),
				Items:      responseItems,
			}
			c.JSON(http.StatusOK, response)
		case "customer":
			// get the order by ID from the database
			order, err := cfg.DB.GetOrderByID(context.Background(), orderID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
				return
			}
			// check if the order belongs to the authenticated customer
			if order.CustomerID != customerID.(uuid.UUID) {
				c.JSON(http.StatusForbidden, gin.H{"error": "you do not have permission to view this order"})
				return
			}
			// get order items for the order
			orderItems, err := cfg.DB.GetOrderItemsByOrderID(context.Background(), order.OrderID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get order items for order ID: %d", order.OrderID)})
				return
			}
			total := 0.00
			responseItems := make([]itemsResponse, 0, len(orderItems))
			for _, item := range orderItems {
				// get product details for each order item
				product, err := cfg.DB.GetProductByID(context.Background(), item.ProductID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get product details for product ID: %d", item.ProductID)})
					return
				}
				unitPrice, err := strconv.ParseFloat(item.Price, 64)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("invalid price format for product ID: %d", item.ProductID)})
					return
				}
				subtotal := unitPrice * float64(item.Quantity)
				responseItems = append(responseItems, itemsResponse{
					ProductID:   item.ProductID,
					ProductName: product.Name,
					Quantity:    item.Quantity,
					Price:       Decimal2(unitPrice),
					Subtotal:    Decimal2(subtotal),
				})
				total += subtotal
			}
			response := OrderResponse{
				ID:         order.OrderID,
				CustomerID: order.CustomerID,
				CreatedAt:  order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
				Total:      Decimal2(total),
				Items:      responseItems,
			}
			c.JSON(http.StatusOK, response)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role query parameter"})
			return
		}
	}

}
