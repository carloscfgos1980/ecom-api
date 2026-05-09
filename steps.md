# STEPS

## 1. Packages

* init de api
go mod init github.com/carloscfgos1980/ecom-api

* install the driver to communicate with postgresql
go get github.com/lib/pq

* install gin framework
go get github.com/gin-gonic/gin

* install packag to handle .env
go get github.com/joho/godotenv

* install package for google uuid
go get github.com/google/uuid

* install argon to hash password
go get github.com/alexedwards/argon2id

* install JWT package
"github.com/golang-jwt/jwt/v5"

## 2. Set up

### 1. Config /internal/config/config.go

1. Define custom error variables for missing configuration values
2. Config struct to hold application configuration values

3. LoadConfig loads configuration values from environment variables and returns a Config struct
3.1 Try common .env locations (project root and cmd/ execution path).
3.2 Load required configuration values from environment variables
3.3 Return the configuration struct with the loaded values

### 2. Set up server and health route

1. Load configuration from environment variables
2. Connect to the database
3. Create a new database queries instance
4. Initialize the Gin router
5. Set trusted proxies to nil to avoid warnings in Gin 1.7+
6. Define a simple health check route
7. Start the server on the specified port

### 3. Make first commit to github

```bash
git init
git add .
git commit -m "first commit"
git remote add origin https://github.com/carloscfgos1980/ecom-api.git
git checkout -b gin_framework
git push orgin gin_framework
```

## 3. Write and Read (excel)

Note: I just as Copilot how to do it and make a few adjustments:

1. database will be reset everytime we add new products
2. products id will match the id from xls file
3. add flags to pass arguments from the CLI

```bash
go run cmd/seed/main.go -mode import
go run cmd/seed/main.go -mode export
```

## 4.Register customer

1. types to handle customer request and response internal/handlers/customer_habdler.go
1.1 structs and handler for creating a new customer in the system
1.2 CustomerRequest is the struct for the request body when creating a new customer

2. CreateCustomerHandler is the handler for creating a new customer internal/handlers/customer_habdler.go
2.1 Bind the JSON request body to the CustomerRequest struct and validate it
2.2 Validate email format
2.3 Validate the password strength
2.4 Hash the password before storing it in the database
2.5 Create the customer in the database using the provided configuration and request data
2.6 Prepare the response with the created customer's information, excluding the password
2.7 Return the created customer information in the response with a 201 Created status

3. Register customer-related routes cmd/main.go
 router.POST("/auth/register", handlers.CreateCustomerHandler(cfg))

## Login Customer

1. LoginCustomerHandler is the handler for logging in a customer and generating a JWT token
1.1 Define a response struct that includes the customer information and the generated token
1.2 Return a handler function that can be used in the Gin router
1.3 Bind the JSON request body to the CustomerRequest struct
1.4 Validate email format
1.5 Retrieve the customer from the database using the provided email
1.6 Check if the provided password matches the stored hashed password
1.7 Generate a JWT token for the authenticated customer
1.8 Prepare the response with the authenticated customer's information and the generated token
1.9 Send the response back to the client with a 200 OK status

2. Register customer-related routes
 router.POST("/auth/login", handlers.LoginCustomerHandler(cfg))

## 5. Get products

1. Product is the struct representing a product in the system /internal/habdlers/products_handler.go

2. GetProductsHandler is the handler for retrieving a list of products
2.1 Return a handler function that can be used in the Gin router
2.2 Retrieve the list of products from the database using the provided configuration
2.3 Prepare the response by converting the products from the database format to the API response format
2.4 Send the list of products back to the client with a 200 OK status

3. Product routes
 router.GET("/products", handlers.GetProductsHandler(cfg))

## 6. Get product by id

1. GetProductByIDHandler is the handler for retrieving a single product by its ID /internal/handlers/products_handler.go
1.1 Get the product ID from the URL parameters
1.2 convert id to int64
1.3 Retrieve the product from the database using the provided configuration and product ID
1.4 Prepare the response by converting the product from the database format to the API response format
1.5 Send the product back to the client with a 200 OK status
2. Product routes
 router.GET("/products/:id", handlers.GetProductByIDHandler(cfg))

## 7. Auth Middleware

1. AuthMiddleware is a Gin middleware function that validates JWT tokens in incoming requests to protect routes that require authentication. It checks for the presence of a valid JWT token in the Authorization header of the request, verifies the token using the secret key from the configuration, and sets the user ID in the Gin context for use in subsequent handlers if the token is valid. If the token is missing or invalid, it returns a 401 Unauthorized response and aborts further processing of the request.
1.1 Return a handler function that can be used in the Gin router as middleware for routes that require authentication.
1.2 If there is an error extracting the token (e.g., missing or malformed header), return a 401 Unauthorized response with an appropriate error message and abort the request processing.
1.3 Validate the extracted token using the secret key from the configuration. If the token is invalid (e.g., expired, malformed, or signature mismatch), return a 401 Unauthorized response with an appropriate error message and abort the request processing.
1.4 If the token is valid, set the customer ID in the Gin context (e.g., using c.Set("customerID", customerID)) for use in subsequent handlers that require authentication.

2. Register order-related routes with authentication middleware
 ordersGroup := router.Group("/orders")
 ordersGroup.Use(middleware.AuthMiddleware(cfg))

## 8. Place orders

1. Create a type to return a number with 2 decimals
1.1 Decimal2 marshals to JSON number with exactly two decimal places
1.2 MarshalJSON implements the json.Marshaler interface for Decimal2.

2. Types /internal/handlers/orders_handler.go
2.1 orderItem represents an item in an order
2.2 OrderResponse represents the response for an order
2.3 itemsResponse represents the response for an order item

3. PlaceOrderHandler is the handler for placing an order
3.1 Return a handler function that can be used in the Gin router
3.2 Get the customer ID from the Gin context (set by the authentication middleware)
3.3 Check if the customer is resgister
3.4 create a new order in the database with the customer ID and the current timestamp
3.5 Bind the JSON body to a slice of orderItem structs
3.6 Loop through the order items, check if the product exists and has enough stock, create order items in the database, and calculate the total price of the order
3.7 look for the product if exists
3.7.1 check if the product exists
3.7.2 check if the product has enough stock
3.7.3 create order item
3.7.4 calculate subtotal for the item
3.7.5 calculate total
3.7.6 update product stock
3.7.7 prepare the response item
3.8 prepare the order response
3.9 send the order response back to the client with a 200 OK status

4. Register order-related routes with authentication middleware
 apiGroup := router.Group("/api")
 apiGroup.Use(middleware.AuthMiddleware(cfg))
 {
  apiGroup.POST("/orders", handlers.PlaceOrderHandler(cfg))

 }

## 9.Get orders

1. GetOrdersHandler is the handler for getting orders /internal/nadlers/order_handler.go
1.1 Return a handler function that can be used in the Gin router
1.2 Get the customer ID from the Gin context (set by the authentication middleware)
1.3 Check if the customer is registered
1.4 Get the role query parameter to determine if the user is an admin or a customer
1.5 If the role is admin, return all orders. If the role is customer, return only the orders for the authenticated customer. If the role is not provided or is invalid, return a bad request error.
1.5.1 get all orders from the database
1.5.2 Loop through the orders and get the order items and product details for each order to prepare the response
1.5.3 format the response with total as decimal with 2 places
1.5.4 Loop through the order items and get the product details for each item to prepare the response
1.5.4 Item response with product details and subtotal
1.5.5 calculate total for the order
1.5.6 Only include orders that have items in the response

1.6 If the role is customer, return only the orders for the authenticated customer
1.6.1 get orders for the authenticated customer from the database
Note: The rest of the steps are the same for adming, the ony different is that the slice of orders will match customerID
2. Register order-related routes with authentication middleware
 apiGroup := router.Group("/api")
 apiGroup.Use(middleware.AuthMiddleware(cfg))
 {
  apiGroup.GET("/orders", handlers.GetOrdersHandler(cfg))

 }
