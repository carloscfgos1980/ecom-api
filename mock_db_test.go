package main

import (
	"context"

	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/google/uuid"
)

// mockDB is a test-only implementation of the DB interface.
// Each field is a function that can be overridden per test.
type mockDB struct {
	createCustomerFn         func(ctx context.Context, arg database.CreateCustomerParams) (database.Customer, error)
	getCustomerByEmailFn     func(ctx context.Context, email string) (database.Customer, error)
	getCustomerByIDFn        func(ctx context.Context, id uuid.UUID) (database.Customer, error)
	getProductsFn            func(ctx context.Context) ([]database.Product, error)
	getProductByIDFn         func(ctx context.Context, id int64) (database.Product, error)
	updateProductStockFn     func(ctx context.Context, arg database.UpdateProductStockParams) error
	createOrderFn            func(ctx context.Context, customerID uuid.UUID) (database.Order, error)
	createOrderItemFn        func(ctx context.Context, arg database.CreateOrderItemParams) error
	getOrdersFn              func(ctx context.Context) ([]database.Order, error)
	getOrdersByCustomerIDFn  func(ctx context.Context, customerID uuid.UUID) ([]database.Order, error)
	getOrderItemsByOrderIDFn func(ctx context.Context, orderID int64) ([]database.GetOrderItemsByOrderIDRow, error)
	getOrderByIDFn           func(ctx context.Context, orderID int64) (database.GetOrderByIDRow, error)
}

func (m *mockDB) CreateCustomer(ctx context.Context, arg database.CreateCustomerParams) (database.Customer, error) {
	return m.createCustomerFn(ctx, arg)
}
func (m *mockDB) GetCustomerByEmail(ctx context.Context, email string) (database.Customer, error) {
	return m.getCustomerByEmailFn(ctx, email)
}
func (m *mockDB) GetCustomerByID(ctx context.Context, id uuid.UUID) (database.Customer, error) {
	return m.getCustomerByIDFn(ctx, id)
}
func (m *mockDB) GetProducts(ctx context.Context) ([]database.Product, error) {
	return m.getProductsFn(ctx)
}
func (m *mockDB) GetProductByID(ctx context.Context, id int64) (database.Product, error) {
	return m.getProductByIDFn(ctx, id)
}
func (m *mockDB) UpdateProductStock(ctx context.Context, arg database.UpdateProductStockParams) error {
	return m.updateProductStockFn(ctx, arg)
}
func (m *mockDB) CreateOrder(ctx context.Context, customerID uuid.UUID) (database.Order, error) {
	return m.createOrderFn(ctx, customerID)
}
func (m *mockDB) CreateOrderItem(ctx context.Context, arg database.CreateOrderItemParams) error {
	return m.createOrderItemFn(ctx, arg)
}
func (m *mockDB) GetOrders(ctx context.Context) ([]database.Order, error) {
	return m.getOrdersFn(ctx)
}
func (m *mockDB) GetOrdersByCustomerID(ctx context.Context, customerID uuid.UUID) ([]database.Order, error) {
	return m.getOrdersByCustomerIDFn(ctx, customerID)
}
func (m *mockDB) GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]database.GetOrderItemsByOrderIDRow, error) {
	return m.getOrderItemsByOrderIDFn(ctx, orderID)
}
func (m *mockDB) GetOrderByID(ctx context.Context, orderID int64) (database.GetOrderByIDRow, error) {
	return m.getOrderByIDFn(ctx, orderID)
}
