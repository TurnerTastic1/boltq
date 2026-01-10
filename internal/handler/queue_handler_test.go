package handler

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/turnertastic1/boltq/internal/queue"
	"github.com/turnertastic1/boltq/internal/store"
	"github.com/turnertastic1/boltq/pkg/queuepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type testDeps struct {
	handler        *QueueHandler
	store          *store.PostgresStore
	queue          *queue.RedisQueue
	pgContainer    *postgres.PostgresContainer
	redisContainer *redis.RedisContainer
	db             *sql.DB
}

func setupTestHandler(t *testing.T) (*testDeps, func()) {
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	pgContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase("boltq"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	// Run migrations
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS jobs (
            id UUID PRIMARY KEY,
            type VARCHAR(50) NOT NULL,
            payload BYTEA NOT NULL,
            status VARCHAR(20) NOT NULL DEFAULT 'queued',
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            started_at TIMESTAMP,
            completed_at TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
        CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC);
    `)
	require.NoError(t, err)

	// Start Redis container
	redisContainer, err := redis.Run(ctx,
		"redis:8.4",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
	)
	require.NoError(t, err)

	// Get Redis connection string
	redisAddr, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Remove "redis://" prefix if present
	redisAddr = strings.TrimPrefix(redisAddr, "redis://")

	logger.Info("Redis address for tests", "addr", redisAddr)

	// Create Redis queue
	redisQueue, err := queue.NewRedisQueue(redisAddr, "", 0)
	require.NoError(t, err)

	// Create store and handler
	pgStore := store.NewPostgresStore(db)
	handler := NewQueueHandler(logger, pgStore, redisQueue)

	deps := &testDeps{
		handler:        handler,
		store:          pgStore,
		queue:          redisQueue,
		pgContainer:    pgContainer,
		redisContainer: redisContainer,
		db:             db,
	}

	cleanup := func() {
		redisQueue.Close()
		db.Close()
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis container: %s", err)
		}
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %s", err)
		}
	}

	return deps, cleanup
}

func TestEnqueueJob_Integration(t *testing.T) {
	deps, cleanup := setupTestHandler(t)
	defer cleanup()

	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_STANDARD,
		Payload: []byte("integration test payload"),
	}

	resp, err := deps.handler.EnqueueJob(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobId)

	// Verify job is stored in Postgres
	var count int
	err = deps.db.QueryRow("SELECT COUNT(*) FROM jobs WHERE id = $1", resp.JobId).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify job in Redis queue
	queueLen, err := deps.queue.GetQueueLength(context.Background(), "JOB_STANDARD")
	require.NoError(t, err)
	assert.Equal(t, int64(1), queueLen)

	// Dequeue the job and verify
	msg, err := deps.queue.Dequeue(context.Background(), "JOB_STANDARD", 1*time.Second)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, resp.JobId, msg.JobID.String())
	assert.Equal(t, "JOB_STANDARD", msg.Type)
}

func TestEnqueueJob_UnspecifiedType(t *testing.T) {
	deps, cleanup := setupTestHandler(t)
	defer cleanup()

	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_TYPE_UNSPECIFIED,
		Payload: []byte("test payload"),
	}

	resp, err := deps.handler.EnqueueJob(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "invalid job type")
}

func TestEnqueueJob_EmptyPayload(t *testing.T) {
	deps, cleanup := setupTestHandler(t)
	defer cleanup()

	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_STANDARD,
		Payload: []byte{},
	}

	resp, err := deps.handler.EnqueueJob(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "payload cannot be empty")
}

func TestEnqueueJob_OversizedPayload(t *testing.T) {
	deps, cleanup := setupTestHandler(t)
	defer cleanup()

	oversizedPayload := make([]byte, maxPayloadSize+1)
	req := &queuepb.EnqueueJobRequest{
		Type:    queuepb.JobType_JOB_STANDARD,
		Payload: oversizedPayload,
	}

	resp, err := deps.handler.EnqueueJob(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "exceeds maximum limit")
}

func TestEnqueueJob_MultipleJobs(t *testing.T) {
	deps, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Enqueue 3 jobs
	jobIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		req := &queuepb.EnqueueJobRequest{
			Type:    queuepb.JobType_JOB_STANDARD,
			Payload: []byte("test payload"),
		}

		resp, err := deps.handler.EnqueueJob(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		jobIDs[i] = resp.JobId
	}

	// Verify count in PostgreSQL
	var dbCount int
	err := deps.db.QueryRow("SELECT COUNT(*) FROM jobs").Scan(&dbCount)
	require.NoError(t, err)
	assert.Equal(t, 3, dbCount)

	// Verify count in Redis
	queueLen, err := deps.queue.GetQueueLength(ctx, "JOB_STANDARD")
	require.NoError(t, err)
	assert.Equal(t, int64(3), queueLen)

	// Dequeue all jobs (FIFO order)
	for i := 0; i < 3; i++ {
		msg, err := deps.queue.Dequeue(ctx, "JOB_STANDARD", 1*time.Second)
		require.NoError(t, err, "job %d should exist", i)
		require.NotNil(t, msg, "job %d should not be nil", i)
		assert.Equal(t, jobIDs[i], msg.JobID.String())
		assert.Equal(t, "JOB_STANDARD", msg.Type)
	}

	// Queue should be empty now
	queueLen, err = deps.queue.GetQueueLength(ctx, "JOB_STANDARD")
	require.NoError(t, err)
	assert.Equal(t, int64(0), queueLen)
}
