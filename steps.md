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

