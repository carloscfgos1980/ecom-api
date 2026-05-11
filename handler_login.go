package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/carloscfgos1980/ecom-api/internal/auth"
)

// handlerLogin handles the login of a customer in the system
func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Define the expected parameters for user login and the response structure
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	// Define the response structure for a successful login, including the customer's information and the generated JWT token
	type response struct {
		Customer
		Token string `json:"token"`
	}
	// Decode the JSON request body into the parameters struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// Retrieve the customer from the database using the provided email address
	customer, err := cfg.db.GetCustomerByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	// Check if the provided password matches the hashed password stored in the database for the retrieved customer
	match, err := auth.CheckPasswordHash(params.Password, customer.Password)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	// If the password is correct, generate a JWT token for the customer to authenticate future requests
	token, err := auth.MakeJWT(
		customer.ID,
		cfg.jwtSecret,
		24*7*time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT token", err)
		return
	}

	// Respond with the customer's information (excluding the password) and the generated JWT token
	respondWithJSON(w, http.StatusOK, response{
		Customer: Customer{
			ID:        customer.ID,
			Email:     customer.Email,
			CreatedAt: customer.CreatedAt,
			UpdatedAt: customer.UpdatedAt,
		},
		Token: token,
	})
}
