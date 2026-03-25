package main

import (
	"log"
	"net/http"
	"time"

	repo "github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/jackc/pgx/v5"

	"github.com/carloscfgos1980/ecom-api/internal/orders"
	"github.com/carloscfgos1980/ecom-api/internal/products"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// mount
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good for now"))
	})

	productService := products.NewService(*repo.New(app.db))
	productsHandler := products.NewHandler(productService)
	r.Get("/products", productsHandler.GetProducts)
	r.Get("/products/{id}", productsHandler.GetProductByID)

	orderService := orders.NewService(repo.New(app.db), app.db)
	ordersHandler := orders.NewHandler(orderService)
	r.Post("/orders", ordersHandler.PlaceOrder)
	r.Get("/orders", ordersHandler.GetOrders)
	r.Get("/orders/{id}", ordersHandler.GetOrderByID)

	return r
}

// run
func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Starting server on %s", app.config.addr)

	return srv.ListenAndServe()
}

type application struct {
	config config
	db     *pgx.Conn
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}
