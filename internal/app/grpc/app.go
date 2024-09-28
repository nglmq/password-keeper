package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	authgrpc "github.com/nglmq/password-keeper/internal/grpc/auth"
	"google.golang.org/grpc"
)

// App
type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

// New create new gRPC server
func New(log *slog.Logger, authService authgrpc.Auth, dataService authgrpc.Data, port int) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService, dataService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// Run runs gRPC server
func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("error Run: %w", err)
	}

	a.log.Info("grpc server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("error serve: %w", err)
	}

	return nil
}

// Stop stops gRPC server
func (a *App) Stop() {
	a.log.Info("stopping grpc server")

	a.gRPCServer.GracefulStop()
}
