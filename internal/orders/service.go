package orders

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	ordersRow, err := s.repo.GetOrders(ctx)
	if err != nil {
		return nil, err
	}
	orders := make([]repo.Order, len(ordersRow))
	for i, row := range ordersRow {
		orders[i] = repo.Order{
			ID:         row.ID,
			CustomerID: row.CustomerID,
			CreatedAt:  row.CreatedAt,
		}
	}
	return orders, nil
}

func (s *svc) GetOrderByID(ctx context.Context, id string) (*repo.Order, error) {
	type response struct {
		ID         int64
		CustomerID int64
		CreatedAt  string
		Items      []struct {
			ID           int64
			OrderID      int64
			ProductID    int64
			Quantity     int64
			PriceInCents int64
		}
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("missing order id")
	}

	// Convert id from string to int64

	orderID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %v", err)
	}

	orderRow, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	log.Printf("orderRow: %+v\n", orderRow)
	order := &repo.Order{
		ID:         orderRow.ID,
		CustomerID: orderRow.CustomerID,
		CreatedAt:  orderRow.CreatedAt,
	}

	return order, nil
}
