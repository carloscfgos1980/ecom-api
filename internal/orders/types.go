package orders

import (
	"context"

	repo "github.com/carloscfgos1980/ecom-api/internal/database"
)

type orderItem struct {
	ProductID int64 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

type createOrderParams struct {
	CustomerID int64       `json:"customerId"`
	Items      []orderItem `json:"items"`
}

type Service interface {
	PlaceOrder(ctx context.Context, tempOrder createOrderParams) (repo.Order, error)
	GetOrders(ctx context.Context) ([]repo.Order, error)
	GetOrderByID(ctx context.Context, id string) (*repo.Order, error)
}
