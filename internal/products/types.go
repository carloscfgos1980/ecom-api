package products

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type ProductResponse struct {
	ID          int64          `json:"productId"`
	Name        string         `json:"productName"`
	Description string         `json:"description"`
	Quantity    int32          `json:"quantity"`
	Price       pgtype.Numeric `json:"price"`
}
