# BoltQ - gRPC Webhook Delivery System

A distributed webhook delivery system powered by gRPC.

## Architecture Overview

```
┌─────────────┐      gRPC       ┌──────────────┐
│   Client    │ ─────────────> │ Queue Service│
│             │                 │   (queue-svc)│
└─────────────┘                 └──────┬───────┘
                                       │
                                       ▼
                                ┌──────────────┐
                                │   Queue      │
                                │ (Redis/etc)  │
                                └──────┬───────┘
                                       │
                                       ▼
                                ┌──────────────┐
                                │   Worker     │
                                │  (Webhook    │
                                │   Delivery)  │
                                └──────────────┘
```

## Components

### 1. Queue Service (cmd/queue-svc)
- gRPC server that receives webhook delivery requests
- Validates incoming requests
- Enqueues jobs for processing
- Returns job ID to clients

### 2. Worker (cmd/worker)
- Consumes jobs from the queue
- Performs actual webhook delivery
- Implements retry logic
- Updates job status

### 3. API (cmd/api)
- Optional REST API gateway
- Translates REST requests to gRPC calls

## Getting Started

### Prerequisites
```bash
# Install Go dependencies
go mod tidy

# Install protoc compiler and plugins (if not already installed)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Running the Queue Service

```bash
# Run the server
go run cmd/queue-svc/main.go
```

The gRPC server will start on port 50051.

### Testing with grpcurl

```bash
# Install grpcurl
brew install grpcurl  # macOS
# or
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List available services
grpcurl -plaintext localhost:50051 list

# Call EnqueueJob
grpcurl -plaintext -d '{
  "type": "webhook.delivery",
  "payload": "eyJ1cmwiOiAiaHR0cHM6Ly9leGFtcGxlLmNvbS93ZWJob29rIiwgIm1ldGhvZCI6ICJQT1NUIn0="
}' localhost:50051 queue.QueueService/EnqueueJob
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run handler tests with verbose output
go test -v ./internal/handler/

# Run with coverage
go test -cover ./...
```

## Request Validation

The handler validates:
- ✅ Request is not nil
- ✅ Job type is present and valid (alphanumeric, dots, underscores, hyphens only)
- ✅ Payload is present
- ✅ Payload size is within limits (max 1MB)

## Next Steps

1. **Implement Queue Storage**
   - Add Redis client for job queue
   - Or use RabbitMQ, Kafka, etc.

2. **Implement Worker**
   - Create worker service to consume jobs
   - Implement HTTP client for webhook delivery
   - Add retry logic with exponential backoff

3. **Add Persistence**
   - Store job metadata in database (PostgreSQL, MongoDB)
   - Track job status and delivery attempts

4. **Add Monitoring**
   - Integrate metrics (Prometheus)
   - Add distributed tracing (OpenTelemetry)
   - Implement logging (structured logs)

5. **Enhance Validation**
   - Validate webhook URL format in payload
   - Validate HTTP methods
   - Add authentication/authorization

6. **Add Features**
   - Job prioritization
   - Scheduled delivery
   - Webhook retry policies
   - Dead letter queue for failed deliveries

## Proto Definition

The service is defined in [proto/queue.proto](proto/queue.proto):

- `EnqueueJob`: Submit a new webhook delivery job
- Future: `GetJobStatus`, `CancelJob`, `ListJobs`, etc.

## Directory Structure

```
boltq/
├── cmd/
│   ├── api/          # REST API gateway (optional)
│   ├── queue-svc/    # gRPC queue service
│   └── worker/       # Webhook delivery worker
├── internal/
│   └── handler/      # gRPC handler implementations
├── pkg/
│   └── queuepb/      # Generated protobuf code
└── proto/
    └── queue.proto   # Service definitions
```
