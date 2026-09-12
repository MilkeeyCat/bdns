package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/MilkeeyCat/bdns/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer stop()

	srv := server.New(8080)

	if err := srv.Start(ctx); err != nil {
		slog.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}
