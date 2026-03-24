package products

import (
	"log"
	"net/http"

	json "github.com/carloscfgos1980/ecom-api/internal"
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
