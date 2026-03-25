package products

import (
	"log"
	"net/http"

	json "github.com/carloscfgos1980/ecom-api/internal"
	"github.com/go-chi/chi/v5"
)

type handler struct {
	service Service
}

func NewHandler(s Service) *handler {
	return &handler{
		service: s,
	}
}

func (h *handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.GetProducts(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.WriteJSON(w, http.StatusOK, products)

}

func (h *handler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}
	product, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.WriteJSON(w, http.StatusOK, product)
}
