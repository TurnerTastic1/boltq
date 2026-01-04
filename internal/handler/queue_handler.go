package handler

import (
	"context"
	"log/slog"

	"github.com/turnertastic1/boltq/pkg/queuepb"
)

type QueueHandler struct {
	queuepb.UnimplementedQueueServiceServer
	logger *slog.Logger
}

func (h *QueueHandler) EnqueueJob(ctx context.Context, req *queuepb.EnqueueJobRequest) (*queuepb.EnqueueJobResponse, error) {
	h.logger.Info("Received EnqueueJob request", "type", req.GetType(), "payload_size", len(req.GetPayload()))

	h.logger.Debug("Job enqueued successfully", "job_id", "job-123")

	return &queuepb.EnqueueJobResponse{
		JobId: "job-123",
	}, nil
}
