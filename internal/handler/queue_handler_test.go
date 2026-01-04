package handler

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turnertastic1/boltq/pkg/queuepb"
)

func TestEnqueueJob(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	handler := NewQueueHandler(logger)

	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_TYPE_WEBHOOK_DELIVERY,
		Payload: []byte("test payload"),
	}

	resp, err := handler.EnqueueJob(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.JobId)
}

func TestEnqueueJob_UnspecifiedType(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	handler := NewQueueHandler(logger)

	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_TYPE_UNSPECIFIED,
		Payload: []byte("test payload"),
	}

	resp, err := handler.EnqueueJob(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestEnqueueJob_EmptyPayload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	handler := NewQueueHandler(logger)

	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_TYPE_WEBHOOK_DELIVERY,
		Payload: []byte{},
	}

	resp, err := handler.EnqueueJob(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestEnqueueJob_OversizedPayload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	handler := NewQueueHandler(logger)

	oversizedPayload := make([]byte, maxPayloadSize+1)
	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_TYPE_WEBHOOK_DELIVERY,
		Payload: oversizedPayload,
	}

	resp, err := handler.EnqueueJob(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}
