package main

import (
	"fmt"
	"github.com/nglmq/password-keeper/internal/app/tui"
	api "github.com/nglmq/password-keeper/internal/clients/sso"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nglmq/password-keeper/internal/app"
	"github.com/nglmq/password-keeper/internal/config"
)

func main() {
	cfg := config.MustLoad()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)

	appl := app.New(log, cfg.DBConnection, cfg.Port)

	go func() {
		if err := appl.GRPCServer.Run(); err != nil {
			log.Error("Failed to run grpc server: ", err)
		}
	}()

	apiClient, err := api.New(log, fmt.Sprintf("localhost:%d", cfg.Port))
	if err != nil {
		log.Error("Failed to create client: ", err)
	}

	tui.StartCLI(apiClient)

	<-stop

	appl.GRPCServer.Stop()
}
