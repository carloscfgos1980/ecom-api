package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/carloscfgos1980/ecom-api/internal/auth"
	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/google/uuid"
)

// structs and handler for creating a new customer in the system
type Customer struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
}

// handlerUsersCreate handles the creation of a new customer in the system
func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	// Define the expected parameters for creating a new customer and the response structure
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	// Define the response structure for a single customer
	type response struct {
		Customer
	}
	// Decode the JSON request body into the parameters struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// Validate the provided parameters (e.g., check if email is valid, password meets criteria, etc.)
	err = auth.IsValidEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// strong password validation can be added here before hashing the password and creating the customer in the database
	err = auth.IsStrongPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	// Check if a customer with the provided email already exists in the database
	_, err = cfg.db.GetCustomerByEmail(r.Context(), params.Email)
	if err == nil {
		respondWithError(w, http.StatusConflict, "Email already exists", nil)
		return
	}
	// If the error is not sql.ErrNoRows, it means there was an issue querying the database
	if err != sql.ErrNoRows {
		respondWithError(w, http.StatusInternalServerError, "Couldn't verify email", err)
		return
	}

	// Hash the customer's password before storing it in the database
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}
	// Create a new customer in the database using the provided parameters and the hashed password
	customer, err := cfg.db.CreateCustomer(r.Context(), database.CreateCustomerParams{
		Email:    params.Email,
		Password: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create customer", err)
		return
	}
	// Respond with the created customer's information (excluding the password)
	respondWithJSON(w, http.StatusCreated, response{
		Customer: Customer{
			ID:        customer.ID,
			CreatedAt: customer.CreatedAt,
			UpdatedAt: customer.UpdatedAt,
			Email:     customer.Email,
		},
	})
}
