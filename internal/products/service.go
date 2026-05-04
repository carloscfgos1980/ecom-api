package products

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	repo "github.com/carloscfgos1980/ecom-api/internal/database"
)

// Service defines the interface for the products service
type Service interface {
	GetProducts(ctx context.Context) ([]repo.Product, error)
	GetProductByID(ctx context.Context, id string) (*repo.Product, error)
}

// svc is the implementation of the Service interface
type svc struct {
	repo *repo.Queries
}

// NewService creates a new service for products
func NewService(repo *repo.Queries) Service {
	return &svc{
		repo: repo,
	}
}

// GetProducts retrieves all products from the database
func (s *svc) GetProducts(ctx context.Context) ([]repo.Product, error) {
	products, err := s.repo.GetProducts(ctx)
	if err != nil {
		return nil, err
	}
	return products, nil
}

// GetProductByID retrieves a product by its ID from the database
func (s *svc) GetProductByID(ctx context.Context, id string) (*repo.Product, error) {
	id = strings.TrimSpace(id)
	// Convert id from string to int64
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}
	// Look for the product if exists
	product, err := s.repo.GetProductByID(ctx, idInt)
	if err != nil {
		return nil, err
	}
	return &product, nil
}
