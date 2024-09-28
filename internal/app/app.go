package app

import (
	"github.com/nglmq/password-keeper/internal/services/auth"
	postgres "github.com/nglmq/password-keeper/internal/storage/pg"
	"log/slog"

	grpcapp "github.com/nglmq/password-keeper/internal/app/grpc"
)

type App struct {
	GRPCServer  *grpcapp.App
	AuthService *auth.Auth
	DataService *auth.Data
}

func New(log *slog.Logger, db string, port int) *App {
	storage, err := postgres.New(db)
	if err != nil {
		log.Error("failed to create storage", err)
	}

	authService := auth.NewAuth(log, storage, storage)
	dataService := auth.NewData(log, storage, storage)

	grpcApp := grpcapp.New(log, authService, dataService, port)

	return &App{
		GRPCServer:  grpcApp,
		AuthService: authService,
		DataService: dataService,
	}
}
