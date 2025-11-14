package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"inteam/internal/config"
	"inteam/internal/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	zapLogger, err := logger.New(cfg.Logging)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer zapLogger.Sync()

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		zapLogger.Fatal("failed to listen", logger.Error(err))
	}

	server := grpc.NewServer()
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	go func() {
		zapLogger.Info("starting gRPC server", logger.String("addr", lis.Addr().String()))
		if err := server.Serve(lis); err != nil {
			zapLogger.Fatal("gRPC server error", logger.Error(err))
		}
	}()

	<-ctx.Done()
	zapLogger.Info("shutting down gRPC server")
	server.GracefulStop()
	os.Exit(0)
}

