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

## 3. Get products

1. Product is the struct representing a product in the system /internal/habdlers/products_handler.go

2. GetProductsHandler is the handler for retrieving a list of products
2.1 Return a handler function that can be used in the Gin router
2.2 Retrieve the list of products from the database using the provided configuration
2.3 Prepare the response by converting the products from the database format to the API response format
2.4 Send the list of products back to the client with a 200 OK status

3. Product routes
 router.GET("/products", handlers.GetProductsHandler(cfg))
