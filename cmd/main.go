package main

import (
	"log/slog"
	"os"
)

func main() {
	cfg := config{
		addr: ":4000",
	}

	api := &application{
		config: cfg,
	}

	// loger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := api.run(api.mount()); err != nil {
		slog.Error("server has failed to start", "err:", err)
		os.Exit(1)
	}
}
