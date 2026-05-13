package main

import (
	"context"

	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/google/uuid"
)

// DB is an interface over *database.Queries so handlers can be tested with a mock.
type DB interface {
	CreateCustomer(ctx context.Context, arg database.CreateCustomerParams) (database.Customer, error)
	GetCustomerByEmail(ctx context.Context, email string) (database.Customer, error)
	GetCustomerByID(ctx context.Context, id uuid.UUID) (database.Customer, error)

	GetProducts(ctx context.Context) ([]database.Product, error)
	GetProductByID(ctx context.Context, id int64) (database.Product, error)
	UpdateProductStock(ctx context.Context, arg database.UpdateProductStockParams) error

	CreateOrder(ctx context.Context, customerID uuid.UUID) (database.Order, error)
	CreateOrderItem(ctx context.Context, arg database.CreateOrderItemParams) error
	GetOrders(ctx context.Context) ([]database.Order, error)
	GetOrdersByCustomerID(ctx context.Context, customerID uuid.UUID) ([]database.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]database.GetOrderItemsByOrderIDRow, error)
	GetOrderByID(ctx context.Context, orderID int64) (database.GetOrderByIDRow, error)
}
