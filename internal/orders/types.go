package orders

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// orderItem represents an item in an order
type orderItem struct {
	ProductID int64 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

// OrderResponse represents the response for an order
type OrderResponse struct {
	ID         int64           `json:"id"`
	CustomerID pgtype.UUID     `json:"customerId"`
	CreatedAt  string          `json:"createdAt"`
	Total      string          `json:"total"`
	Items      []itemsResponse `json:"items"`
}

// itemsResponse represents the response for an order item
type itemsResponse struct {
	ProductID   int64  `json:"productId"`
	ProductName string `json:"productName"`
	Quantity    int32  `json:"quantity"`
	Price       string `json:"price"`
	Subtotal    string `json:"subtotal"`
}
