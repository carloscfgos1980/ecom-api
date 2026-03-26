package orders

import (
	"log"
	"net/http"

	"github.com/carloscfgos1980/ecom-api/internal/json"
	"github.com/go-chi/chi/v5"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var tempOrder createOrderParams
	if err := json.Read(r, &tempOrder); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdOrder, err := h.service.PlaceOrder(r.Context(), tempOrder)
	if err != nil {
		log.Println(err)

		if err == ErrProductNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.WriteJSON(w, http.StatusCreated, createdOrder)
}

func (h *handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.GetOrders(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.WriteJSON(w, http.StatusOK, orders)
}

func (h *handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}
	order, err := h.service.GetOrderByID(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	orderItems, err := h.service.GetOrderItemsByOrderID(r.Context(), order.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var createdAtStr string
	if order.CreatedAt.Valid {
		createdAtStr = order.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	response := OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		CreatedAt:  createdAtStr,
	}

	for _, item := range orderItems {
		response.Items = append(response.Items, struct {
			ProductID       int64  `json:"productId"`
			ProductName     string `json:"name"`
			Quantity        int32  `json:"quantity"`
			PriceInCents    int32  `json:"priceInCents"`
			SubtotalInCents int32  `json:"subtotalInCents"`
		}{
			ProductID:       item.ProductID,
			ProductName:     item.ProductName,
			Quantity:        item.Quantity,
			PriceInCents:    item.PriceInCents,
			SubtotalInCents: item.SubtotalInCents,
		})
		response.TotalInCents += item.SubtotalInCents
	}

	json.WriteJSON(w, http.StatusOK, response)
}
