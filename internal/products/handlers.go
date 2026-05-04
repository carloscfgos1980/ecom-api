package products

import (
	"log"
	"net/http"

	"github.com/carloscfgos1980/ecom-api/internal/json"
	"github.com/go-chi/chi/v5"
)

// handler is the HTTP handler for products endpoints
type handler struct {
	service Service
}

// NewHandler creates a new handler for products endpoints
func NewHandler(s Service) *handler {
	return &handler{
		service: s,
	}
}

// GetProducts handles the GET /products endpoint to retrieve all products
func (h *handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	// call the service to get all products
	products, err := h.service.GetProducts(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Convert products to ProductResponse and write JSON response
	response := make([]ProductResponse, len(products))
	for i, p := range products {
		response[i] = ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Quantity:    p.Quantity,
			Price:       p.Price,
			Description: p.Description,
		}
	}
	// Write the JSON response with a 200 OK status
	json.WriteJSON(w, http.StatusOK, response)

}

// GetProductByID handles the GET /products/{id} endpoint to retrieve a product by its ID
func (h *handler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	// get the product id from the URL parameters
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}
	// call the service to get the product by its ID and return a 200 OK with the product in the response body
	product, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Convert the product to ProductResponse
	response := ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Quantity:    product.Quantity,
		Price:       product.Price,
		Description: product.Description,
	}
	// Write the JSON response with a 200 OK status
	json.WriteJSON(w, http.StatusOK, response)
}
