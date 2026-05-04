package orders

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	repo "github.com/carloscfgos1980/ecom-api/internal/database"
)

type orderItem struct {
	ProductID int64 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

type createOrderParams struct {
	CustomerID uuid.UUID   `json:"customerId"`
	Items      []orderItem `json:"items"`
}

type OrderResponse struct {
	ID         int64     `json:"id"`
	CustomerID uuid.UUID `json:"customerId"`
	CreatedAt  string    `json:"createdAt"`
	Total      int32     `json:"totalInCents"`
	Items      []struct {
		ProductID   int64          `json:"productId"`
		ProductName string         `json:"productName"`
		Quantity    int32          `json:"quantity"`
		Price       pgtype.Numeric `json:"price"`
		Subtotal    pgtype.Numeric `json:"subtotal"`
	} `json:"items"`
}

type Service interface {
	PlaceOrder(ctx context.Context, tempOrder createOrderParams) (repo.Order, error)
	GetOrders(ctx context.Context) ([]repo.Order, error)
	GetOrderByID(ctx context.Context, id string) (*repo.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]repo.GetOrderItemsByOrderIDRow, error)
}
