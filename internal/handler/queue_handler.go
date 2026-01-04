package handler

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/turnertastic1/boltq/pkg/queuepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QueueHandler struct {
	queuepb.UnimplementedQueueServiceServer
	logger *slog.Logger
}

func NewQueueHandler(logger *slog.Logger) *QueueHandler {
	return &QueueHandler{
		logger: logger,
	}
}

const maxPayloadSize = 1024 * 1024 // 1 MB

func (h *QueueHandler) EnqueueJob(ctx context.Context, req *queuepb.EnqueueJobRequest) (*queuepb.EnqueueJobResponse, error) {
	// Validate job type
	if req.GetType() == queuepb.JobType_JOB_TYPE_UNSPECIFIED {
		h.logger.Warn("Invalid job type", "type", req.GetType())
		return nil, status.Error(codes.InvalidArgument, "invalid job type")
	}

	// Validate payload exists
	if len(req.GetPayload()) == 0 {
		h.logger.Warn("Payload is empty")
		return nil, status.Error(codes.InvalidArgument, "payload cannot be empty")
	}

	if len(req.GetPayload()) > maxPayloadSize {
		h.logger.Warn("Payload size exceeds maximum limit", "size", len(req.GetPayload()))
		return nil, status.Errorf(codes.InvalidArgument, "payload size exceeds maximum limit: %d", maxPayloadSize)
	}

	h.logger.Info("Received EnqueueJob request", "type", req.GetType(), "payload_size", len(req.GetPayload()))

	jobId := uuid.New()

	return &queuepb.EnqueueJobResponse{
		JobId: jobId.String(),
	}, nil
}
