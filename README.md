# ECOM-API

## Project Description

RestFul API with 3 endpoints:

* register costumer(auth -JWT)
* products
* orders

* I use **gin** framework

## Main features

### Read and Write (excel document)

### auth

* Register
* Login

### products

* Get all products
* Get a single product by id

### orders with auth. Just authentication.

* Place an order
* Get all the orders. URL param: role = admin, Gets all the orders. role = customer return only the orders of the registered customer. If not URL params is given it wuld return an error
* Get a single order by id. Same auth feature implemented for the route orders

### Data persisted

* Postgres
* Migrations are running with **goose** using **pgx** package


## ⚙️ Installation

Inside a Go module:

```bash
go get -b gin_framework github.com/carloscfgos1980/ecom-api
```

## 🚀 Quick Start Consumer

```go
go run cmd/*.go
```

## 📖 Usage

### Read and write (excel)

```bash
go run cmd/seed/main.go -file=path/to/products.xlsx -sheet=Sheet1 -mode=import
go run cmd/seed/main.go -file=path/to/products.xlsx -sheet=Sheet1 -mode=export
```

Note: I ran into issues with the path. The problem was that sometimes I ran the command from the root directory and passing the path and sometimes I did it from the directory where the main.go file host the logic to read and write (cmd/import-export-products/main.go). Solution is to pick a single way to do it. I recommend to run it from the root directory and to have a related path in the .env file

```bash
go run cmd/seed/main.go -mode=import
go run cmd/seed/main.go -mode=export
```

### programs needed to run the api

1. goose (migrations)
2. SQLC (generate Go code from SQL queries)
3. pq (driver for postgres)
4. gin (framework)
5. Argon2id (encrypt password)
6. golang-jwt (create JWT token)

## 🤝 Contributing

### Clone the repo

```bash
git clone -b gin_framework github.com/carloscfgos1980/ecom-api
cd ecom-api
```

### Build the compiled binary

```bash
go build
```

### Submit a pull request

If you'd like to contribute, please fork the repository and open a pull request to the `gin_framework` branch.

## Building a Production API in Golang from Scratch

[Ecommerce project](https://www.youtube.com/watch?v=s3XItrqfccw&t=4710s)
