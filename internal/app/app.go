package app

import (
	"log/slog"

	grpcapp "github.com/nglmq/password-keeper/internal/app/grpc"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, db string, port int) *App {
	// init storage

	// init auth service

	grpcApp := grpcapp.New(log, port)

	return &App{
		GRPCServer: grpcApp,
	}
}
