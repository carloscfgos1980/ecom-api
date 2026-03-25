package products

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	repo "github.com/carloscfgos1980/ecom-api/internal/database"
)

type Service interface {
	GetProducts(ctx context.Context) ([]repo.Product, error)
	GetProductByID(ctx context.Context, id string) (*repo.Product, error)
}

type svc struct {
	repo repo.Queries
}

func NewService(repo repo.Queries) Service {
	return &svc{
		repo: repo,
	}
}

func (s *svc) GetProducts(ctx context.Context) ([]repo.Product, error) {
	products, err := s.repo.GetProducts(ctx)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *svc) GetProductByID(ctx context.Context, id string) (*repo.Product, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("missing product id")
	}

	// Convert id from string to int64
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}
	product, err := s.repo.GetProductByID(ctx, idInt)
	if err != nil {
		return nil, err
	}
	return &product, nil
}
