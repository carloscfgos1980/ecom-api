package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/carloscfgos1980/ecom-api/internal/env"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	// load env variables
	cfg := config{
		addr: ":5000",
		db: dbConfig{
			dsn: env.GetEnv("GOOSE_DBSTRING", "host=localhost user=postgres password=postgres dbname=db_ecom sslmode=disable"),
		},
	}
	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// database connection
	conn, err := pgx.Connect(ctx, cfg.db.dsn)

	if err != nil {
		slog.Error("unable to connect to database", "err", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	logger.Info("successfully connected to database", "dsn", cfg.db.dsn)

	api := &application{
		config: cfg,
		db:     conn,
	}
	if err := api.run(api.mount()); err != nil {
		slog.Error("server has failed to start", "err", err)
		os.Exit(1)
	}
}
