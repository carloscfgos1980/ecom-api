package orders

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	repo "github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/jackc/pgx/v5"
)

// errors that can be returned by the service
var (
	ErrProductNotFound = errors.New("product not found")
	ErrProductNoStock  = errors.New("product has not enough stock")
)

type Service interface {
	PlaceOrder(ctx context.Context, customerID pgtype.UUID, items []orderItem) (repo.Order, error)
	GetOrders(ctx context.Context) ([]repo.Order, error)
	GetOrderByID(ctx context.Context, id string) (*repo.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]repo.GetOrderItemsByOrderIDRow, error)
}

// svc is the implementation of the Service interface
type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

// NewService creates a new service for orders
func NewService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

// PlaceOrder creates a new order with the given parameters
func (s *svc) PlaceOrder(ctx context.Context, customerID pgtype.UUID, items []orderItem) (repo.Order, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return repo.Order{}, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)

	// create the order
	order, err := qtx.CreateOrder(ctx, customerID)
	if err != nil {
		return repo.Order{}, err
	}

	// look for the product if exists
	for _, item := range items {
		product, err := qtx.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return repo.Order{}, ErrProductNotFound
		}

		if product.Quantity < item.Quantity {
			return repo.Order{}, ErrProductNoStock
		}

		// create order item
		_, err = qtx.CreateOrderItem(ctx, repo.CreateOrderItemParams{
			OrderID:   order.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		})
		if err != nil {
			return repo.Order{}, err
		}

		// update product stock
		newQuantity := product.Quantity - item.Quantity
		err = qtx.UpdateProductStock(ctx, repo.UpdateProductStockParams{
			ID:       product.ID,
			Quantity: newQuantity,
		})
		if err != nil {
			return repo.Order{}, err
		}
	}

	tx.Commit(ctx)

	return order, nil
}

// GetOrders returns all orders
func (s *svc) GetOrders(ctx context.Context) ([]repo.Order, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)
	// get all orders
	orders, err := qtx.GetOrders(ctx)
	if err != nil {
		return nil, err
	}
	// commit the transaction
	tx.Commit(ctx)

	return orders, nil
}

// GetOrderByID returns an order by its ID
func (s *svc) GetOrderByID(ctx context.Context, id string) (*repo.Order, error) {
	// validate the order ID and return a 400 Bad Request if it's invalid
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("missing order id")
	}

	// Convert id from string to int64
	orderID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %v", err)
	}
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)
	orderRow, err := qtx.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	tx.Commit(ctx)

	order := &repo.Order{
		OrderID:    orderRow.OrderID,
		CustomerID: orderRow.CustomerID,
		CreatedAt:  orderRow.CreatedAt,
	}

	return order, nil
}

func (s *svc) GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]repo.GetOrderItemsByOrderIDRow, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)
	itemsRow, err := qtx.GetOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	tx.Commit(ctx)

	items := make([]repo.GetOrderItemsByOrderIDRow, len(itemsRow))
	for i, row := range itemsRow {
		items[i] = repo.GetOrderItemsByOrderIDRow{
			ID:          row.ID,
			OrderID:     row.OrderID,
			ProductID:   row.ProductID,
			ProductName: row.ProductName,
			Quantity:    row.Quantity,
			Price:       row.Price,
			Subtotal:    row.Subtotal,
		}
	}

	return items, nil
}
