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
