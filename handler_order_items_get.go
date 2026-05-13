package main

import (
	"net/http"
	"strconv"

	"github.com/carloscfgos1980/ecom-api/internal/auth"
)

// handlerOrdersItemsGet handles of all orders and their items for a customer in the system
func (cfg *apiConfig) handlerOrdersItemsGet(w http.ResponseWriter, r *http.Request) {
	// Extract the JWT token from the Authorization header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	// Validate JWT and get customer ID
	customerID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	// Check if the customer exists in the database
	_, err = cfg.db.GetCustomerByID(r.Context(), customerID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Customer not found", err)
		return
	}
	// Get the role query parameter to determine if the user is an admin or a customer
	role := r.URL.Query().Get("role")
	// Based on the role, retrieve the appropriate orders and their items for the authenticated customer or all customers if the user is an admin
	switch role {
	// If the role is admin, get all orders and their items for all customers
	case "admin":
		// get all orders and their items for all customers
		orders, err := cfg.db.GetOrders(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get orders and items", err)
			return
		}
		if len(orders) == 0 {
			respondWithError(w, http.StatusNotFound, "no orders found", nil)
			return
		}
		// ordersResponse will hold the list of orders and their items to be returned in the response
		ordersResponse := make([]OrderResponse, 0, len(orders))
		// Loop through the orders and get the order items and product details for each order to prepare the response
		for _, order := range orders {
			orderItems, err := cfg.db.GetOrderItemsByOrderID(r.Context(), order.OrderID)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't get order items", err)
				return
			}
			// Calculate the total and prepare the response items for the order
			total := Decimal2(0)
			responseItems := make([]itemsResponse, 0, len(orderItems))
			// Loop through the order items to get the product details and calculate the subtotal for each item, as well as the total for the order
			for _, item := range orderItems {
				// Get the product details for the current order item to include in the response and calculate the subtotal and total for the order
				product, err := cfg.db.GetProductByID(r.Context(), item.ProductID)
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Couldn't get product details", err)
					return
				}
				// Convert the product price from string to float64 to calculate the subtotal and total for the order
				price, err := strconv.ParseFloat(product.Price, 64)
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Invalid product price", err)
					return
				}
				// Calculate the subtotal for the current order item and add it to the total for the order, as well as prepare the response item with the product details, quantity, price, and subtotal
				subtotal := Decimal2(float64(item.Quantity) * price)
				total += subtotal
				// Append the current order item with the product details, quantity, price, and subtotal to the response items for the order
				responseItems = append(responseItems, itemsResponse{
					ProductID:   item.ProductID,
					ProductName: product.Name,
					Quantity:    item.Quantity,
					Price:       Decimal2(price),
					Subtotal:    subtotal,
				})
			}
			// Append the current order with its details, total, and response items to the ordersResponse to be returned in the response
			ordersResponse = append(ordersResponse, OrderResponse{
				ID:         order.OrderID,
				CustomerID: order.CustomerID,
				CreatedAt:  order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
				Total:      total,
				Items:      responseItems,
			})
		}
		// Respond with the list of orders and their items in JSON format
		respondWithJSON(w, http.StatusOK, ordersResponse)
	// If the role is customer, get all orders and their items for the authenticated customer
	case "customer":
		// If the role is customer, get all orders and their items for the authenticated customer
		orders, err := cfg.db.GetOrdersByCustomerID(r.Context(), customerID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get orders and items", err)
			return
		}
		if len(orders) == 0 {
			respondWithError(w, http.StatusNotFound, "no orders found for this customer", nil)
			return
		}
		ordersResponse := make([]OrderResponse, 0, len(orders))
		// Loop through the orders and get the order items and product details for each order to prepare the response
		for _, order := range orders {
			orderItems, err := cfg.db.GetOrderItemsByOrderID(r.Context(), order.OrderID)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't get order items", err)
				return
			}
			total := Decimal2(0)
			responseItems := make([]itemsResponse, 0, len(orderItems))
			for _, item := range orderItems {
				product, err := cfg.db.GetProductByID(r.Context(), item.ProductID)
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Couldn't get product details", err)
					return
				}
				price, err := strconv.ParseFloat(product.Price, 64)
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Invalid product price", err)
					return
				}
				subtotal := Decimal2(float64(item.Quantity) * price)
				total += subtotal
				responseItems = append(responseItems, itemsResponse{
					ProductID:   item.ProductID,
					ProductName: product.Name,
					Quantity:    item.Quantity,
					Price:       Decimal2(price),
					Subtotal:    subtotal,
				})
			}
			ordersResponse = append(ordersResponse, OrderResponse{
				ID:         order.OrderID,
				CustomerID: order.CustomerID,
				CreatedAt:  order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
				Total:      total,
				Items:      responseItems,
			})
		}
		// Respond with the list of orders and their items in JSON format
		respondWithJSON(w, http.StatusOK, ordersResponse)
	default:
		// If the role query parameter is missing or invalid, respond with a bad request error
		respondWithError(w, http.StatusBadRequest, "Invalid role query parameter", nil)
		return
	}
}

// handlerOrderItemsGetByOrderID handles the retrieval of a single order and its items by the order ID for a customer in the system
func (cfg *apiConfig) handlerOrderItemsGetByOrderID(w http.ResponseWriter, r *http.Request) {
	// Extract the JWT token from the Authorization header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	// Validate JWT and get customer ID
	customerID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	// Check if the customer exists in the database
	_, err = cfg.db.GetCustomerByID(r.Context(), customerID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Customer not found", err)
		return
	}
	// Extract the order ID from the URL path or query parameters
	stringID := r.PathValue("orderID")
	if stringID == "" {
		stringID = r.URL.Query().Get("orderID")
	}
	// Validate the provided order ID (e.g., check if it's a valid integer, etc.)
	orderID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid order ID", err)
		return
	}
	// Get the role query parameter to determine if the user is an admin or a customer
	role := r.URL.Query().Get("role")
	// Based on the role, retrieve the appropriate order and its items for the authenticated customer or any order if the user is an admin
	switch role {
	// If the role is admin, get the order and its items for any order, otherwise if the role is customer, get the order and its items for the authenticated customer
	case "admin":
		// If the role is admin, get the order and its items for any order
		order, err := cfg.db.GetOrderByID(r.Context(), orderID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get order", err)
			return
		}
		// Get the order items and product details for the order to prepare the response
		orderItems, err := cfg.db.GetOrderItemsByOrderID(r.Context(), order.OrderID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get order items", err)
			return
		}
		// Calculate the total and prepare the response items for the order
		total := Decimal2(0)
		responseItems := make([]itemsResponse, 0, len(orderItems))
		// Loop through the order items to get the product details and calculate the subtotal for each item, as well as the total for the order
		for _, item := range orderItems {
			// Get the product details for the current order item to include in the response and calculate the subtotal and total for the order
			product, err := cfg.db.GetProductByID(r.Context(), item.ProductID)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't get product details", err)
				return
			}
			// Convert the product price from string to float64 to calculate the subtotal and total for the order
			price, err := strconv.ParseFloat(product.Price, 64)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Invalid product price", err)
				return
			}
			// Calculate the subtotal for the current order item and add it to the total for the order, as well as prepare the response item with the product details, quantity, price, and subtotal
			subtotal := Decimal2(float64(item.Quantity) * price)
			total += subtotal
			// Append the current order item with the product details, quantity, price, and subtotal to the response items for the order
			responseItems = append(responseItems, itemsResponse{
				ProductID:   item.ProductID,
				ProductName: product.Name,
				Quantity:    item.Quantity,
				Price:       Decimal2(price),
				Subtotal:    subtotal,
			})
		}
		// Prepare the response with the order details, total, and response items to be returned in the response
		response := OrderResponse{
			ID:         order.OrderID,
			CustomerID: order.CustomerID,
			CreatedAt:  order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			Total:      total,
			Items:      responseItems,
		}
		// Respond with the order and its items in JSON format
		respondWithJSON(w, http.StatusOK, response)
	// If the role is customer, get the order and its items for the authenticated customer
	case "customer":
		// If the role is customer, get the order and its items for the authenticated customer
		order, err := cfg.db.GetOrderByID(r.Context(), orderID)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				respondWithError(w, http.StatusNotFound, "order not found", nil)
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Couldn't get order", err)
			return
		}
		// check if the order belongs to the authenticated customer
		if order.CustomerID != customerID {
			respondWithError(w, http.StatusForbidden, "You don't have access to this order", nil)
			return
		}
		// get order items for the order
		orderItems, err := cfg.db.GetOrderItemsByOrderID(r.Context(), order.OrderID)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				respondWithError(w, http.StatusNotFound, "order items not found", nil)
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Couldn't get order items", err)
			return
		}
		// Calculate the total and prepare the response items for the order
		total := Decimal2(0)
		responseItems := make([]itemsResponse, 0, len(orderItems))
		// Loop through the order items to get the product details and calculate the subtotal for each item, as well as the total for the order
		for _, item := range orderItems {
			product, err := cfg.db.GetProductByID(r.Context(), item.ProductID)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't get product details", err)
				return
			}
			// Convert the product price from string to float64 to calculate the subtotal and total for the order
			price, err := strconv.ParseFloat(product.Price, 64)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Invalid product price", err)
				return
			}
			// Calculate the subtotal for the current order item and add it to the total for the order, as well as prepare the response item with the product details, quantity, price, and subtotal
			subtotal := Decimal2(float64(item.Quantity) * price)
			total += subtotal
			// Append the current order item with the product details, quantity, price, and subtotal to the response items for the order
			responseItems = append(responseItems, itemsResponse{
				ProductID:   item.ProductID,
				ProductName: product.Name,
				Quantity:    item.Quantity,
				Price:       Decimal2(price),
				Subtotal:    subtotal,
			})
		}
		// Prepare the response with the order details, total, and response items to be returned in the response
		response := OrderResponse{
			ID:         order.OrderID,
			CustomerID: order.CustomerID,
			CreatedAt:  order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			Total:      total,
			Items:      responseItems,
		}
		// Respond with the order and its items in JSON format
		respondWithJSON(w, http.StatusOK, response)
	default:
		// If the role query parameter is missing or invalid, respond with a bad request error
		respondWithError(w, http.StatusBadRequest, "Invalid role query parameter", nil)
		return
	}
}
