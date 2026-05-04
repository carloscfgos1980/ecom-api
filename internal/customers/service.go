package customers

import (
	"context"

	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/jackc/pgx/v5"
)

// Service defines the interface for the customers service
type Service interface {
	CreateCustomer(ctx context.Context, customer CustomerRequest) (database.Customer, error)
}

// svc defines the struct for the customers service
type svc struct {
	repo *database.Queries
	db   *pgx.Conn
}

// NewService creates a new service for the customers package
func NewService(repo *database.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

// CreateCustomer creates a new customer in the database
func (s *svc) CreateCustomer(ctx context.Context, customer CustomerRequest) (database.Customer, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.Customer{}, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)

	// create the customer
	createdCustomer, err := qtx.CreateCustomer(ctx, database.CreateCustomerParams{
		Email:    customer.Email,
		Password: customer.Password,
	})
	if err != nil {
		return database.Customer{}, err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return database.Customer{}, err
	}
	// return the created customer
	return createdCustomer, nil
}
