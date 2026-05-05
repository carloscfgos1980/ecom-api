# ECOM-API

## Project Description

RestFul API with 2 endpoints:

* customers
* products
* orders

## Main features

### Read and Write (excel document)

### auth

* Register
* Login

### products

* Get all products
* Get a single product by id

### orders with auth. Just authentication. No authorization implemented

* Place an order
* Get all the orders
* Get a single order by id

### Data persisted

* Postgres
* Migrations are running with **goose** using **pgx** package

### routes

* I use **chi** package

## ⚙️ Installation

Inside a Go module:

```bash
go get -b version2 github.com/carloscfgos1980/ecom-api
```

## 🚀 Quick Start Consumer

```go
go run cmd/*.go
```

## 📖 Usage

### Read and write (excel)

```bash
go run cmd/import-export-products/main.go -file=path/to/products.xlsx -sheet=Sheet1 -mode=import
go run cmd/import-export-products/main.go -file=path/to/products.xlsx -sheet=Sheet1 -mode=export
```

Note: I ran into issues with the path. The problem was that sometimes I ran the command from the root directory and passing the path and sometimes I did it from the directory where the main.go file host the logic to read and write (cmd/import-export-products/main.go). Solution is to pick a single way to do it. I recommend to run it from the root directory and to have a related path in the .env file

```bash
go run cmd/import-export-products/main.go -mode=import
go run cmd/import-export-products/main.go -mode=export
```

### programs needed to run the api


1. goose (migrations)
2. SQLC (generate Go code from SQL queries)
3. pgx (package to connect to databse)
4. chi (package tp build the routes)
5. Argon2id (encrypt password)
6. golang-jwt (create JWT token)

## 🤝 Contributing

### Clone the repo

```bash
git clone -b version2 github.com/carloscfgos1980/ecom-api
cd ecom-api
```

### Build the compiled binary

```bash
go build
```

### Submit a pull request

If you'd like to contribute, please fork the repository and open a pull request to the `main` branch.

## Building a Production API in Golang from Scratch

[Ecommerce project](https://www.youtube.com/watch?v=s3XItrqfccw&t=4710s)
