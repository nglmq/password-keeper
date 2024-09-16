package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nglmq/password-keeper/internal/app"
	"github.com/nglmq/password-keeper/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)

	appl := app.New(log, cfg.DBConnection, cfg.Port)

	appl.GRPCServer.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	appl.GRPCServer.Stop()
}
