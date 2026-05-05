package main

import (
	"log"
	"net/http"
	"time"

	repo "github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/jackc/pgx/v5"

	"github.com/carloscfgos1980/ecom-api/internal/authmiddleware"
	"github.com/carloscfgos1980/ecom-api/internal/customers"
	"github.com/carloscfgos1980/ecom-api/internal/orders"
	"github.com/carloscfgos1980/ecom-api/internal/products"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// mount sets up the routes and middleware for the application and returns the http.Handler
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
	// health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good for now"))
	})
	// customers endpoints
	// create the customer service and handler
	customerService := customers.NewService(repo.New(app.db), app.db)
	customerHandler := customers.NewHandler(customerService, app.config.JWTSecret)
	// set up the customers routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", customerHandler.CreateCustomer)
		r.Post("/login", customerHandler.LoginCustomer)
	})
	// products endpoints
	productService := products.NewService(repo.New(app.db))
	productsHandler := products.NewHandler(productService)
	r.Get("/products", productsHandler.GetProducts)
	r.Get("/products/{id}", productsHandler.GetProductByID)

	// protected routes
	r.Route("/api", func(r chi.Router) {
		// Add authentication middleware here if available
		r.Use(func(next http.Handler) http.Handler {
			return authmiddleware.AuthMiddleware(next, app.config.JWTSecret)
		})
		// orders endpoints
		orderService := orders.NewService(repo.New(app.db), app.db)
		ordersHandler := orders.NewHandler(orderService)
		r.Post("/orders", ordersHandler.PlaceOrder)
		// r.Get("/orders", ordersHandler.GetOrders)
		// r.Get("/orders/{id}", ordersHandler.GetOrderByID)
	})
	return r
}

// run starts the HTTP server
func (app *application) run(h http.Handler) error {
	// create the HTTP server
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Starting server on %s", app.config.addr)

	return srv.ListenAndServe()
}

// application is the main application struct that holds the configuration and database connection
type application struct {
	config config
	db     *pgx.Conn
}

// config holds the configuration for the application
type config struct {
	addr      string
	db        dbConfig
	JWTSecret string
}

// dbConfig holds the database configuration for the application
type dbConfig struct {
	dsn string
}
