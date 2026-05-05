package orders

import (
	"fmt"
	"log"
	"math/big"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/carloscfgos1980/ecom-api/internal/json"
	"github.com/carloscfgos1980/ecom-api/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
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
	// get the customer ID from the context
	customerID := r.Context().Value("customerID")
	if customerID == nil {
		log.Println("customerID not found in context")
		http.Error(w, "customerID not found in context", http.StatusInternalServerError)
		return
	}
	//check if the customer is registered in the database
	_, err := h.service.GetCustomerByID(r.Context(), customerID.(pgtype.UUID))
	if err != nil {
		log.Println(err)
		http.Error(w, "you must be a registered customer to place an order", http.StatusNotFound)
		return
	}
	// read the request body and unmarshal it into a slice of orderItems
	var items []orderItem
	// validate the request body and return a 400 Bad Request if it's invalid
	if err := json.ReadJSON(r, &items); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// call the service to place the order and return a 201 Created with the created order in the response body
	createdOrder, err := h.service.PlaceOrder(r.Context(), customerID.(pgtype.UUID), items)
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
	// format the response with total as decimal with 2 places
	orderItems, err := h.service.GetOrderItemsByOrderID(r.Context(), createdOrder.OrderID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// calculate total from items
	var total pgtype.Numeric
	total.Valid = true
	total.Int = new(big.Int)
	for _, item := range orderItems {
		if item.Subtotal.Valid {
			total.Int.Add(total.Int, item.Subtotal.Int)
		}
	}
	response := OrderResponse{
		ID:         createdOrder.OrderID,
		CustomerID: createdOrder.CustomerID,
		CreatedAt:  utils.FormatTimestamp(createdOrder.CreatedAt),
		Total:      utils.FormatTotal(total),
		Items:      make([]itemsResponse, len(orderItems)),
	}
	// map order items to response
	for i, item := range orderItems {
		response.Items[i] = itemsResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       utils.FormatNumeric(item.Price),
			Subtotal:    utils.FormatNumeric(item.Subtotal),
		}
	}
	// return the created order in the response body
	json.WriteJSON(w, http.StatusCreated, response)
}

// GetOrders handles the GET /orders endpoint to get all orders for the authenticated customer
func (h *handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	// get the customer ID from the context
	customerID := r.Context().Value("customerID")
	if customerID == nil {
		log.Println("customerID not found in context")
		http.Error(w, "customerID not found in context", http.StatusInternalServerError)
		return
	}
	//check if the customer is registered in the database
	_, err := h.service.GetCustomerByID(r.Context(), customerID.(pgtype.UUID))
	if err != nil {
		log.Println(err)
		http.Error(w, "you must be a registered customer to place an order", http.StatusNotFound)
		return
	}
	// call the service to get the orders
	orders, err := h.service.GetOrders(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// format the response with total as decimal with 2 places
	ordersResponse := make([]OrderResponse, len(orders))
	// map orders to response
	for i, order := range orders {
		// get order items for each order
		order, err := h.service.GetOrderByID(r.Context(), fmt.Sprintf("%d", order.OrderID))
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// get items for each order
		orderItems, err := h.service.GetOrderItemsByOrderID(r.Context(), order.OrderID)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// calculate total from items
		var total pgtype.Numeric
		total.Valid = true
		total.Int = new(big.Int)
		// map order items to response
		var items []itemsResponse = make([]itemsResponse, len(orderItems))
		for j, item := range orderItems {
			items[j] = itemsResponse{
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				Price:       utils.FormatNumeric(item.Price),
				Subtotal:    utils.FormatNumeric(item.Subtotal),
			}
			// add subtotal to total
			if item.Subtotal.Valid {
				total.Int.Add(total.Int, item.Subtotal.Int)
			}
		}
		// map order to response
		ordersResponse[i] = OrderResponse{
			ID:         order.OrderID,
			CustomerID: order.CustomerID,
			CreatedAt:  utils.FormatTimestamp(order.CreatedAt),
			Total:      utils.FormatTotal(total),
			Items:      items,
		}
	}
	// return the orders in the response body
	json.WriteJSON(w, http.StatusOK, ordersResponse)
}

// GetOrderByID handles the GET /orders/{id} endpoint to get an order by ID for the authenticated customer
func (h *handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	// get the customer ID from the context
	customerID := r.Context().Value("customerID")
	if customerID == nil {
		log.Println("customerID not found in context")
		http.Error(w, "customerID not found in context", http.StatusInternalServerError)
		return
	}
	//check if the customer is registered in the database
	_, err := h.service.GetCustomerByID(r.Context(), customerID.(pgtype.UUID))
	if err != nil {
		log.Println(err)
		http.Error(w, "you must be a registered customer to place an order", http.StatusNotFound)
		return
	}
	// get the order ID from the URL parameter
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}
	// call the service to get the order by ID
	order, err := h.service.GetOrderByID(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// get order items for the order
	orderItems, err := h.service.GetOrderItemsByOrderID(r.Context(), order.OrderID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// calculate total from items
	var total pgtype.Numeric
	total.Valid = true
	total.Int = new(big.Int)
	// map order items to response
	var items []itemsResponse = make([]itemsResponse, len(orderItems))
	for j, item := range orderItems {
		items[j] = itemsResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       utils.FormatNumeric(item.Price),
			Subtotal:    utils.FormatNumeric(item.Subtotal),
		}
		// add subtotal to total
		if item.Subtotal.Valid {
			total.Int.Add(total.Int, item.Subtotal.Int)
		}
	}
	// map order to response
	response := OrderResponse{
		ID:         order.OrderID,
		CustomerID: order.CustomerID,
		CreatedAt:  utils.FormatTimestamp(order.CreatedAt),
		Total:      utils.FormatTotal(total),
		Items:      items,
	}
	// return the order in the response body
	json.WriteJSON(w, http.StatusOK, response)
}
