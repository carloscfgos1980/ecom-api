package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/carloscfgos1980/ecom-api/internal/auth"
	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/google/uuid"
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

// handlerOrderCreate handles the placement of an order by a customer in the system
func (cfg *apiConfig) handlerOrderCreate(w http.ResponseWriter, r *http.Request) {
	// Extract the JWT token from the Authorization header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	// Validate JWT and get customer ID
	customerID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	// Check if the customer exists in the database
	_, err = cfg.db.GetCustomerByID(r.Context(), customerID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Customer not found", err)
		return
	}
	// Define the expected parameters for placing an order and the response structure
	var params []orderItem
	// Decode the JSON request body into the parameters struct
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// Validate the provided parameters (e.g., check if items are valid, quantities are positive, etc.)
	if len(params) == 0 {
		respondWithError(w, http.StatusBadRequest, "No items provided", nil)
		return
	}
	// Create a new order in the database using the provided parameters and the authenticated customer's ID
	order, err := cfg.db.CreateOrder(context.Background(), customerID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create order", err)
		return
	}
	// Initialize total price for the order and prepare the response items for the order response
	total := Decimal2(0)
	responseItems := make([]itemsResponse, 0, len(params))
	// Create order items in the database for each item in the order
	for _, item := range params {
		// check if the quantity is greater than zero
		if item.Quantity <= 0 {
			respondWithError(w, http.StatusBadRequest, "Quantity must be greater than zero", nil)
			return
		}
		//check if the product exists in the database
		product, err := cfg.db.GetProductByID(r.Context(), item.ProductID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Product with ID %d not found", item.ProductID), err)
			return
		}
		//check if the product has enough quantity in stock
		if product.Quantity < item.Quantity {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Not enough quantity in stock for product with ID %d", item.ProductID), nil)
			return
		}
		//create the order item in the database
		err = cfg.db.CreateOrderItem(context.Background(), database.CreateOrderItemParams{
			OrderID:   order.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create order item", err)
			return
		}
		//update the product quantity in the database
		err = cfg.db.UpdateProductStock(context.Background(), database.UpdateProductStockParams{
			ID:       item.ProductID,
			Quantity: product.Quantity - item.Quantity,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update product:%s", product.Name), err)
			return
		}
		// convert price and subtotal to Decimal2 for consistent JSON formatting
		price, err := strconv.ParseFloat(product.Price, 64)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Invalid product price", err)
			return
		}
		// calculate the subtotal for the order item and add it to the total price for the order
		subtotal := Decimal2(float64(item.Quantity) * price)
		total += subtotal
		// prepare the response items for the order response
		responseItems = append(responseItems, itemsResponse{
			ProductID:   item.ProductID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			Price:       Decimal2(price),
			Subtotal:    subtotal,
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

	// Respond with the created order's information in JSON format
	respondWithJSON(w, http.StatusCreated, response)

}
