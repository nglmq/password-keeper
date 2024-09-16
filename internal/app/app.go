package app

import (
	"github.com/nglmq/password-keeper/internal/services/auth"
	postgres "github.com/nglmq/password-keeper/internal/storage/pg"
	"log/slog"

	grpcapp "github.com/nglmq/password-keeper/internal/app/grpc"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, db string, port int) *App {
	storage, err := postgres.New(db)
	if err != nil {
		log.Error("failed to create storage", err)
	}

	authService := auth.New(log, storage, storage)

	grpcApp := grpcapp.New(log, authService, port)

	return &App{
		GRPCServer: grpcApp,
	}
}
