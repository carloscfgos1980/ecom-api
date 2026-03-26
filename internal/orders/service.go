package orders

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	repo "github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/jackc/pgx/v5"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrProductNoStock  = errors.New("product has not enough stock")
)

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) PlaceOrder(ctx context.Context, tempOrder createOrderParams) (repo.Order, error) {
	// validate payload
	if tempOrder.CustomerID == 0 {
		return repo.Order{}, fmt.Errorf("customer ID is required")
	}
	if len(tempOrder.Items) == 0 {
		return repo.Order{}, fmt.Errorf("at least one item is required")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return repo.Order{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)

	// create an order
	order, err := qtx.CreateOrder(ctx, tempOrder.CustomerID)
	if err != nil {
		return repo.Order{}, err
	}

	// look for the product if exists
	for _, item := range tempOrder.Items {
		product, err := qtx.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return repo.Order{}, ErrProductNotFound
		}

		if product.Quantity < item.Quantity {
			return repo.Order{}, ErrProductNoStock
		}

		// create order item
		_, err = qtx.CreateOrderItem(ctx, repo.CreateOrderItemParams{
			OrderID:      order.ID,
			ProductID:    item.ProductID,
			Quantity:     item.Quantity,
			PriceInCents: product.PriceInCents,
		})
		if err != nil {
			return repo.Order{}, err
		}

		// Challenge: Update the product stock quantity
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

func (s *svc) GetOrders(ctx context.Context) ([]repo.Order, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)
	orders, err := qtx.GetOrders(ctx)
	if err != nil {
		return nil, err
	}

	tx.Commit(ctx)
	return orders, nil
}

func (s *svc) GetOrderByID(ctx context.Context, id string) (*repo.Order, error) {

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("missing order id")
	}

	// Convert id from string to int64

	orderID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %v", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)
	order, err := qtx.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	tx.Commit(ctx)
	return &order, nil
}

func (s *svc) GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]repo.OrderItem, error) {
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

	items := make([]repo.OrderItem, len(itemsRow))
	for i, row := range itemsRow {
		items[i] = repo.OrderItem{
			ID:           row.ID,
			OrderID:      row.OrderID,
			ProductID:    row.ProductID,
			Quantity:     row.Quantity,
			PriceInCents: row.PriceInCents,
			ProductName:  row.ProductName,
		}
	}

	return items, nil
}
