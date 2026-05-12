# STEPS

## 1.Packages

* Start mod init
go mod init github.com/carloscfgos1980/ecom-api

* Install package to load .env
go get github.com/joho/godotenv

* install package for postgresql driver
go get github.com/lib/pq

* install uuid package
go get github.com/google/uuid

* install package to hash password
go get github.com/alexedwards/argon2id

* install package to create and verify jwt token
go get github.com/golang-jwt/jwt/v5

## 2. Setup /main.go

1. Load environment variables from .env file
2. Get configuration from environment variables
3. Get the port from environment variables, default to 8080 if not set
4. Get the JWT secret from environment variables
5. Connect to the database
6. database queries variable
7. variable for the apiConfig struct
8. Set up the HTTP server and routes
9. health check endpoint
10. Listen and serve

## 3. Read and write from exel /seed/main.go

Note: Just as copilot to build the file. commands

```bash
go run cmd/seed/main.go -mode import
go run cmd/seed/main.go -mode export
```

## 4. First commit to Github

```bash
git init
git add .
git commit -m "first commit"
git remote add origin https://github.com/carloscfgos1980/ecom-api.git
git checkout -b no-framework
git push origin no_framework
```

## 5. Register customer

1. auxiliary functions to response with JSON /json.go
1.1 Helper functions for responding with JSON and error messages in the API handlers
1.2 respondWithJSON is a helper function to send JSON responses with a given HTTP status code and payload

2. structs and handler for creating a new customer in the system /handler_users_create.go
3. handlerUsersCreate handles the creation of a new customer in the system
3.1 Define the expected parameters for creating a new customer and the response structure
3.2 Define the response structure for a single customer
3.3 Decode the JSON request body into the parameters struct
3.4 Validate the provided parameters (e.g., check if email is valid, password meets criteria, etc.)
3.5 strong password validation can be added here before hashing the password and creating the customer in the database
3.6 Check if a customer with the provided email already exists in the database
3.7 If the error is not sql.ErrNoRows, it means there was an issue querying the database
3.8 Hash the customer's password before storing it in the database
3.9 Create a new customer in the database using the provided parameters and the hashed password
3.10 Respond with the created customer's information (excluding the password)

4. Register the handler for creating a new customer /main.go
 mux.HandleFunc("/auth/register", apiCfg.handlerUsersCreate)

## 6. Login

1. handlerLogin handles the login of a customer in the system
1.1 Define the expected parameters for user login and the response structure
1.2 Define the response structure for a successful login, including the customer's information and the generated JWT token
1.3 Decode the JSON request body into the parameters struct
1.4 Retrieve the customer from the database using the provided email address
1.5 Check if the provided password matches the hashed password stored in the database for the retrieved customer
1.6 If the password is correct, generate a JWT token for the customer to authenticate future requests
1.7 Respond with the customer's information (excluding the password) and the generated JWT token

2. Register the handler for logging in a customer
 mux.HandleFunc("/auth/login", apiCfg.handlerLogin)

## 7. Get products

1. ProductResponse defines the structure of the response for a single product /handler_products_get.go
2. handlerProductsGet handles the retrieval of all products in the system
2.1 Retrieve all products from the database
2.2 Define the response structure for a list of products
2.3 Convert the retrieved products to the response format
2.4 Respond with the list of products in JSON format
3. Register the handler for retrieving all products
 mux.HandleFunc("GET /products", apiCfg.handlerProductsGet)

## 8. Get product by Id 

1.handlerProductsGetByID handles the retrieval of a single product by its ID /handler_products_get.go
1.1 Extract the product ID from the URL path or query parameters
1.2 Validate the provided product ID (e.g., check if it's a valid integer, etc.)
1.3 Retrieve the product from the database using the provided ID
1.4 Define the response structure for a single product
1.5 Respond with the product information in JSON format
2. Register the handler for retrieving a product by ID /main.go
mux.HandleFunc("GET /products/{productID}", apiCfg.handlerProductsGetByID)
