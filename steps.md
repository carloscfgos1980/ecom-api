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

1. queries to get lists of products and a single product
2. run generate go code from sql (sqlc generate)
3. Create response struct /internal/products/types.go

4. Set up service /internal/products/service.go
4.1 svc is the implementation of the Service interface
4.2 NewService creates a new service for products
4.3 Service defines the interface for the products service

5. GetProducts method of svc retrieves all products from the database
6. Add GetProducts method to service interface

7. GetProducts handles the GET /products endpoint to retrieve all products
7.1 all the service to get all products
7.2 Convert products to ProductResponse and write JSON response
7.3 Write the JSON response with a 200 OK status

8. products endpoints
 productService := products.NewService(repo.New(app.db))
 productsHandler := products.NewHandler(productService)
 r.Get("/products", productsHandler.GetProducts)
