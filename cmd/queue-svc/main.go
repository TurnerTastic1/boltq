package main

import (
	"log/slog"
	"net"
	"os"

	"github.com/turnertastic1/boltq/internal/handler"
	"github.com/turnertastic1/boltq/pkg/queuepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("Failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()

	queueHandler := handler.NewQueueHandler(logger)
	queuepb.RegisterQueueServiceServer(grpcServer, queueHandler)

	reflection.Register(grpcServer)

	logger.Info("Starting gRPC server on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("Failed to serve gRPC server", "error", err)
		os.Exit(1)
	}
}
