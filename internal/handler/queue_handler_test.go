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
		Type:    "webhook.delivery",
		Payload: []byte("test"),
	}

	resp, err := handler.EnqueueJob(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.JobId)
}
