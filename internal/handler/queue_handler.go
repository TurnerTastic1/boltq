package handler

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/turnertastic1/boltq/internal/store"
	"github.com/turnertastic1/boltq/pkg/queuepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QueueHandler struct {
	queuepb.UnimplementedQueueServiceServer
	logger *slog.Logger
	store  *store.PostgresStore
}

func NewQueueHandler(logger *slog.Logger, store *store.PostgresStore) *QueueHandler {
	return &QueueHandler{
		logger: logger,
		store:  store,
	}
}

const maxPayloadSize = 1024 * 1024 // 1 MB

func (h *QueueHandler) EnqueueJob(ctx context.Context, req *queuepb.EnqueueJobRequest) (*queuepb.EnqueueJobResponse, error) {
	if req.GetType() == queuepb.JobType_JOB_TYPE_UNSPECIFIED {
		h.logger.Warn("Invalid job type", "type", req.GetType())
		return nil, status.Error(codes.InvalidArgument, "invalid job type")
	}

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

	job := &store.Job{
		ID:      jobId,
		Type:    req.GetType().String(),
		Payload: req.GetPayload(),
		Status:  store.JobStatusQueued,
	}

	if err := h.store.CreateJob(ctx, job); err != nil {
		h.logger.Error("Failed to create job in store", "error", err)
		return nil, status.Error(codes.Internal, "failed to enqueue job")
	}

	h.logger.Info("Job enqueued successfully", "job_id", jobId.String())

	return &queuepb.EnqueueJobResponse{
		JobId: jobId.String(),
	}, nil
}

func (h *QueueHandler) GetJobStatus(ctx context.Context, req *queuepb.GetJobStatusRequest) (*queuepb.GetJobStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetJobStatus not implemented")
}
