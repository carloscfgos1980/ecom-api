# STEPS

## 1. Set up

* Create database

```bash
psql postgres
CREATE DATABASE ecom_db;
exit
```

* Run Migrations

goose postgres "postgres://carlosinfante:@localhost:5432/db_ecom?sslmode=disable" up

* Chances to the tables price from cents to float
* Add a description column to the products table

pxg driver for database
go get "github.com/jackc/pgx/v5"

package to encrypt password
go get "github.com/alexedwards/argon2id"

package to create jwt
go get "github.com/golang-jwt/jwt/v5"

## 2. Read and Write to exel

### 1. Path to the files

XLS_FILE_PATH_READ="../../data/products_start.xls"
XLS_FILE_PATH_WRITE="../../data/products_export.xlsx"

Note: It does not allow me to create .xls, instead .xlsx. Also take in account that the path needs to go to levels up

### 3. Create folder and file to handle read and write /cmd/import-products/main.go

1. This tool imports products from an .xls or .xlsx file into the database, replacing existing products.
2. Command-line flags
3. Load .env values (if available) to allow env vars to override them
4. Determine DSN for database connection
5. Connect to the database
6. Use pgx directly for simplicity; in a real app, you might use a connection pool or an ORM
7. Run the appropriate mode
7.1 Determine file path for import
7.2 If file path is still empty, use default
7.3 Import products from the specified file and sheet
8. For export mode, determine file path and handle .xls extension by switching to .xlsx
8.1 Determine file path for export
8.2 export to data folder by default, but allow override via env or flag

// Usage:
// go run main.go -file=path/to/products.xlsx -sheet=Sheet1 -mode=import
// go run main.go -file=path/to/products.xlsx -sheet=Sheet1 -mode=export

Note: The whole code was written by Copilot after my promps. I Had to write it a couple times to get the result I want

```bash
cd cmd/import-export-products
go run . -mode import
go run . -mode export
```

Note: When I re run the import I ran into an issue with the relative path so I had to reset the path the root directory since I am running the app from the root directory

## 3. Get products

Note. I had to drop the tables cos the migrations was given a head to change the previos data type

1. queries to get lists of products
2. run generate go code from sql (sqlc generate)
3. Create response struct /internal/products/types.go

4. Set up service /internal/products/service.go
4.1 svc is the implementation of the Service interface
4.2 NewService creates a new service for products
4.3 Service defines the interface for the products service

5. GetProducts method of svc retrieves all products from the database
6. Add GetProducts method to service interface

7. GetProducts handles the GET /products endpoint to retrieve all products /internal/products/handler.go
7.1 Call the service to get all products
7.2 Convert products to ProductResponse and write JSON response
7.3 Write the JSON response with a 200 OK status

8. products endpoints cmd/api.go
 productService := products.NewService(repo.New(app.db))
 productsHandler := products.NewHandler(productService)
 r.Get("/products", productsHandler.GetProducts)

## 4. Get a product

1. queries to get a single products by id
2. run generate go code from sql (sqlc generate)
3. GetProductByID method of svc retrieves a product by its ID from the database /internal/products/service.go
4. Add GetProductByID to service interface

5. GetProductByID handles the GET /products/{id} endpoint to retrieve a product by its ID
5.1 get the product id from the URL parameters
5.2 call the service to get the product by its ID and return a 200 OK with the product in the response body
5.3 Convert the product to ProductResponse
5.4 Write the JSON response with a 200 OK status

6. products endpoints /cmd/api.go
 r.Get("/products/{id}", productsHandler.GetProductByID)

## 5. Register customer

1. Schema for customers (sql/schemas/003_customers.sql)
1.1 goose postgres "postgres://carlosinfante:@localhost:5432/db_ecom?sslmode=disable" up
2. Queries add customer, get customer by id and email
2.1 sqlc generate

3. auxiliar functions for auth
3.1 HashPassword
3.2 CheckPasswordHash
3.3 MakeJWT
3.4 ValidateJWT
3.5 GetBearerToken
3.6 IsStrongPassword
3.7 IsValidEmail
Note: I copied the functions from taskSpehre. Just I made an adjustment. Instead of using google uuid, I use ppgtype.UUID which the data type that comes with the driver. This will avoid me later the need to conver customer id

4. Types for auth
4.1 structs and handler for creating a new user in the system
4.2 CustomerRequest is the struct for the request body when creating a new customer
4.3 LoginResponse is the response body when logging in a user.

5. Service set up /internal/customers/types.go
5.1 Service defines the interface for the customers service
5.2 svc defines the struct for the customers service
5.3 NewService creates a new service for the customers package

6. CreateCustomer method of svc creates a new customer in the database
/internal/customers/service.go
7. Add CreateCustomer to service interface

8. Handler set up /internal/customers/handler.go
8.1 handler is the HTTP handler for users endpoints
8.2 NewHandler creates a new handler for users endpoints

9. CreateCustomer handles the HTTP request for creating a new customer
9.1 Parse the JSON request body into a CustomerRequest struct
9.2 Check if any field is empty
9.3 Validate email format
9.4 Validate the password strength
9.5 Hash the password before storing it in the database
9.6 Update the customer request with the hashed password
9.7 Call the service to create the customer
9.8 Check if the error is a unique constraint violation (duplicate email)
9.9 Create a response struct to send back to the client, excluding the password
9.10 Write the response as JSON with a 201 Created status code

10. customers endpoints
10.1 create the customer service and handler
 customerService := customers.NewService(repo.New(app.db), app.db)
 customerHandler := customers.NewHandler(customerService, app.config.JWTSecret)
 // set up the customers routes
 r.Route("/auth", func(r chi.Router) {
  r.Post("/register", customerHandler.CreateCustomer)
 })

11. Add JWT secret
11.1 Add a field in config struct for jwt secret cmd/api.go
type config struct {
 addr      string
 db        dbConfig
 JWTSecret string
}
11.2 Get the JWT secret from environment variables
11.3 Load JWT secret to cfg variable

## 6 Login