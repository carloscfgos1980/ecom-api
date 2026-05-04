package customers

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// structs and handler for creating a new customer in the system
type Customer struct {
	ID        pgtype.UUID `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Email     string      `json:"email"`
	Password  string      `json:"password"`
}

// CustomerRequest is the struct for the request body when creating a new customer
type CustomerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the response body when logging in a customer.
type LoginResponse struct {
	Customer
	Token string `json:"token"`
}
