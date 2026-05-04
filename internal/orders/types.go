package orders

import (
	"github.com/carloscfgos1980/ecom-api/internal/products"

	"github.com/jackc/pgx/v5/pgtype"
)

type orderItem struct {
	ProductID int64 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

type createOrderParams struct {
	CustomerID pgtype.UUID `json:"customerId"`
	Items      []orderItem `json:"items"`
}

type OrderResponse struct {
	ID         int64                      `json:"id"`
	CustomerID pgtype.UUID                `json:"customerId"`
	CreatedAt  string                     `json:"createdAt"`
	Total      int32                      `json:"totalInCents"`
	Items      []products.ProductResponse `json:"items"`
}
