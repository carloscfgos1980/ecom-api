package products

import (
	"context"

	repo "github.com/carloscfgos1980/ecom-api/internal/database"
)

type Service interface {
	GetProducts(ctx context.Context) ([]repo.Product, error)
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
	products, error := s.repo.GetProducts(ctx)
	if error != nil {
		return nil, error
	}
	return products, nil
}
