package orders

import (
	"fmt"
	"log"
	"net/http"

	"github.com/carloscfgos1980/ecom-api/internal/json"
	"github.com/go-chi/chi/v5"
)

// handler is the HTTP handler for orders endpoints
type handler struct {
	service Service
}

// NewHandler creates a new handler for orders endpoints
func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

// PlaceOrder handles the POST /orders endpoint to create a new order
func (h *handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	// read the request body and unmarshal it into a createOrderParams struct
	var tempOrder createOrderParams
	// validate the request body and return a 400 Bad Request if it's invalid
	if err := json.Read(r, &tempOrder); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// call the service to place the order and return a 201 Created with the created order in the response body
	createdOrder, err := h.service.PlaceOrder(r.Context(), tempOrder)
	if err != nil {
		log.Println(err)
		if err == ErrProductNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == ErrProductNoStock {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// return the created order in the response body
	json.WriteJSON(w, http.StatusCreated, createdOrder)
}

func (h *handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.GetOrders(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ordersResponse := make([]OrderResponse, len(orders))
	for i, order := range orders {
		order, err := h.service.GetOrderByID(r.Context(), fmt.Sprintf("%d", order.ID))
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orderItems, err := h.service.GetOrderItemsByOrderID(r.Context(), order.ID)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		items := make([]struct {
			ProductID       int64  `json:"productId"`
			ProductName     string `json:"productName"`
			Quantity        int32  `json:"quantity"`
			PriceInCents    int32  `json:"priceInCents"`
			SubtotalInCents int32  `json:"subtotalInCents"`
		}, len(orderItems))
		var totalInCents int32
		for j, item := range orderItems {
			items[j] = struct {
				ProductID       int64  `json:"productId"`
				ProductName     string `json:"productName"`
				Quantity        int32  `json:"quantity"`
				PriceInCents    int32  `json:"priceInCents"`
				SubtotalInCents int32  `json:"subtotalInCents"`
			}{
				ProductID:       item.ProductID,
				ProductName:     item.ProductName,
				Quantity:        item.Quantity,
				PriceInCents:    item.PriceInCents,
				SubtotalInCents: item.SubtotalInCents,
			}
			totalInCents += item.SubtotalInCents
		}
		var createdAtStr string
		if order.CreatedAt.Valid {
			createdAtStr = order.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		ordersResponse[i] = OrderResponse{
			ID:           order.ID,
			CustomerID:   order.CustomerID,
			CreatedAt:    createdAtStr,
			TotalInCents: totalInCents,
			Items:        items,
		}
		ordersResponse[i] = OrderResponse{
			ID:           order.ID,
			CustomerID:   order.CustomerID,
			CreatedAt:    createdAtStr,
			TotalInCents: totalInCents,
			Items:        items,
		}
	}
	json.WriteJSON(w, http.StatusOK, ordersResponse)
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
			ProductName     string `json:"productName"`
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
