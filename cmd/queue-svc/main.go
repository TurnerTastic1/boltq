package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/turnertastic1/boltq/internal/handler"
	"github.com/turnertastic1/boltq/internal/store"
	"github.com/turnertastic1/boltq/pkg/queuepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Starting BoltQ Queue Service...")

	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort := getEnv("POSTGRES_PORT", "5432")
	pgUser := getEnv("POSTGRES_USER", "boltq")
	pgPassword := getEnv("POSTGRES_PASSWORD", "boltq_dev")
	pgDB := getEnv("POSTGRES_DB", "boltq")

	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPassword, pgDB,
	)

	logger.Info("Connecting to Postgres", "connString", connString)

	// Create database connection
	db, err := sql.Open("postgres", connString)
	if err != nil {
		logger.Error("Failed to open database connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	pgStore := store.NewPostgresStore(db)
	defer pgStore.Close()

	logger.Info("Connected to Postgres successfully")

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("Failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()

	queueHandler := handler.NewQueueHandler(logger, pgStore)
	queuepb.RegisterQueueServiceServer(grpcServer, queueHandler)

	reflection.Register(grpcServer)

	logger.Info("Starting gRPC server on :50051")

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down gracefully...")
		grpcServer.GracefulStop()
	}()

	// Start serving
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("Failed to serve gRPC server", "error", err)
		os.Exit(1)
	}
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
