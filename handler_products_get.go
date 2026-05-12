package main

import (
	"net/http"
	"strconv"
)

// ProductResponse defines the structure of the response for a single product
type ProductResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Quantity    int32  `json:"quantity"`
}

// handlerProductsGet handles the retrieval of all products in the system
func (cfg *apiConfig) handlerProductsGet(w http.ResponseWriter, r *http.Request) {
	// Retrieve all products from the database
	products, err := cfg.db.GetProducts(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve products", err)
		return
	}
	// Define the response structure for a list of products
	response := []ProductResponse{}
	// Convert the retrieved products to the response format
	for _, product := range products {
		response = append(response, ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Quantity:    product.Quantity,
		})
	}
	// Respond with the list of products in JSON format
	respondWithJSON(w, http.StatusOK, response)
}

// handlerProductsGetByID handles the retrieval of a single product by its ID
func (cfg *apiConfig) handlerProductsGetByID(w http.ResponseWriter, r *http.Request) {
	// Extract the product ID from the URL path or query parameters
	stringID := r.PathValue("productID")
	if stringID == "" {
		stringID = r.URL.Query().Get("productID")
	}
	// Validate the provided product ID (e.g., check if it's a valid integer, etc.)
	productID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID", err)
		return
	}
	// Retrieve the product from the database using the provided ID
	product, err := cfg.db.GetProductByID(r.Context(), productID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve product", err)
		return
	}
	// Define the response structure for a single product
	response := ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Quantity:    product.Quantity,
	}
	// Respond with the product information in JSON format
	respondWithJSON(w, http.StatusOK, response)
}
