package handler

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/turnertastic1/boltq/pkg/queuepb"
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

func (h *QueueHandler) EnqueueJob(ctx context.Context, req *queuepb.EnqueueJobRequest) (*queuepb.EnqueueJobResponse, error) {
	h.logger.Info("Received EnqueueJob request", "type", req.GetType(), "payload_size", len(req.GetPayload()))

	jobId := uuid.New()

	return &queuepb.EnqueueJobResponse{
		JobId: jobId.String(),
	}, nil
}
