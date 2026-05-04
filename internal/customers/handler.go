package customers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/carloscfgos1980/ecom-api/internal/json"
	"github.com/carloscfgos1980/ecom-api/internal/utils"
	"github.com/jackc/pgx/v5/pgconn"
)

// handler is the HTTP handler for customers endpoints
type handler struct {
	service   Service
	jwtSecret string
}

// NewHandler creates a new handler for users endpoints
func NewHandler(service Service, jwtSecret string) *handler {
	return &handler{
		service:   service,
		jwtSecret: jwtSecret,
	}
}

// CreateCustomer handles the HTTP request for creating a new customer
func (h *handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body into a CustomerRequest struct
	var customerReq CustomerRequest
	if err := json.ReadJSON(r, &customerReq); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Check if any field is empty
	if customerReq.Email == "" || customerReq.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}
	// Validate email format
	err := utils.IsValidEmail(customerReq.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Validate the password strength
	err = utils.IsStrongPassword(customerReq.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Hash the password before storing it in the database
	hashedPassword, err := utils.HashPassword(customerReq.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Update the customer request with the hashed password
	customerReq.Password = hashedPassword

	// Call the service to create the customer
	customer, err := h.service.CreateCustomer(r.Context(), customerReq)
	if err != nil {
		log.Println(err)
		// Check if the error is a unique constraint violation (duplicate email)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// PostgreSQL unique violation error code
			if strings.Contains(pgErr.Message, "email") {
				http.Error(w, "Email already exists", http.StatusConflict)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create a response struct to send back to the client, excluding the password
	response := Customer{
		ID:        customer.ID,
		CreatedAt: customer.CreatedAt.Time,
		UpdatedAt: customer.UpdatedAt.Time,
		Email:     customer.Email,
	}
	// Write the response as JSON with a 201 Created status code
	if err := json.WriteJSON(w, http.StatusCreated, response); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// LoginCustomer handles the HTTP request for logging in a customer
func (h *handler) LoginCustomer(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body into a CustomerRequest struct
	var customerReq CustomerRequest
	if err := json.ReadJSON(r, &customerReq); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Check if email and password are provided
	if customerReq.Email == "" || customerReq.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}
	// Get the customer by email from the database
	customer, err := h.service.GetCustomerByEmail(r.Context(), customerReq.Email)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	// Check if the provided password matches the stored hashed password
	match, err := utils.CheckPasswordHash(customerReq.Password, customer.Password)
	if err != nil || !match {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	// Generate a JWT token for the authenticated user
	token, err := utils.MakeJWT(
		customer.ID,
		h.jwtSecret,
		24*7*time.Hour,
	)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Create a response struct to send back to the client with the access token
	response := LoginResponse{
		Customer: Customer{
			ID:        customer.ID,
			CreatedAt: customer.CreatedAt.Time,
			UpdatedAt: customer.UpdatedAt.Time,
			Email:     customer.Email,
		},
		Token: token,
	}
	// Write the response as JSON with a 200 OK status code
	if err := json.WriteJSON(w, http.StatusOK, response); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
