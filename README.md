# ECOM-API

## Project Description

RestFul API with 2 endpoints:

* products
* orders

## Main features

### products

* Get all products
* Get a single product by id

### orders

* Place an order
* Get all the orders
* Get a single order by id

### Data persisted

* Postgres 16 using a Docker image
* Migrations are running with **goose** using **pgx** package

### routes

* I use **chi** package

## ⚙️ Installation

Inside a Go module:

```bash
go get github.com/carloscfgos1980/ecom-api
```

## 🚀 Quick Start Consumer

```go
go run cmd/*.go
```

## 📖 Usage

### programs needed to run the api

1. postgres (Docker image)
2. goose (migrations)
3. SQLC (generate Go code from SQL queries)
4. pgx (package to connect to databse)
5. chi (package tp build the routes)

## 🤝 Contributing

### Clone the repo

```bash
git clone github.com/carloscfgos1980/ecom-api
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
