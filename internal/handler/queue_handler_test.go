package handler

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/turnertastic1/boltq/internal/store"
	"github.com/turnertastic1/boltq/pkg/queuepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupTestHandler(t *testing.T) (*QueueHandler, sqlmock.Sqlmock, func()) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	pgStore := store.NewPostgresStore(db)

	handler := NewQueueHandler(logger, pgStore)

	cleanup := func() {
		db.Close()
	}

	return handler, mock, cleanup
}

func TestEnqueueJob(t *testing.T) {
	handler, mock, cleanup := setupTestHandler(t)
	defer cleanup()

	mock.ExpectExec("INSERT INTO jobs").
		WithArgs(sqlmock.AnyArg(), "JOB_TYPE_WEBHOOK_DELIVERY", sqlmock.AnyArg(), store.JobStatusQueued).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_TYPE_WEBHOOK_DELIVERY,
		Payload: []byte("test payload"),
	}

	resp, err := handler.EnqueueJob(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.JobId)

	_, err = uuid.Parse(resp.JobId)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEnqueueJob_UnspecifiedType(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_TYPE_UNSPECIFIED,
		Payload: []byte("test payload"),
	}

	resp, err := handler.EnqueueJob(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "invalid job type")
}

func TestEnqueueJob_EmptyPayload(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_TYPE_WEBHOOK_DELIVERY,
		Payload: []byte{},
	}

	resp, err := handler.EnqueueJob(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "payload cannot be empty")
}

func TestEnqueueJob_OversizedPayload(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	oversizedPayload := make([]byte, maxPayloadSize+1)
	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_TYPE_WEBHOOK_DELIVERY,
		Payload: oversizedPayload,
	}

	resp, err := handler.EnqueueJob(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "exceeds maximum limit")
}

func TestEnqueueJob_DatabaseError(t *testing.T) {
	handler, mock, cleanup := setupTestHandler(t)
	defer cleanup()

	// Simulate database failure
	mock.ExpectExec("INSERT INTO jobs").
		WithArgs(sqlmock.AnyArg(), "JOB_TYPE_WEBHOOK_DELIVERY", []byte("test payload"), store.JobStatusQueued).
		WillReturnError(assert.AnError)

	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_TYPE_WEBHOOK_DELIVERY,
		Payload: []byte("test payload"),
	}

	resp, err := handler.EnqueueJob(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to enqueue job")

	assert.NoError(t, mock.ExpectationsWereMet())
}
