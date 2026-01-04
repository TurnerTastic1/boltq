package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test CreateJob separately
func TestPostgresStore_CreateJob(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{db: db}
	ctx := context.Background()

	jobID := uuid.New()

	// Only mock the INSERT
	mock.ExpectExec("INSERT INTO jobs").
		WithArgs(jobID, "webhook.delivery", []byte("test payload"), JobStatusQueued).
		WillReturnResult(sqlmock.NewResult(1, 1))

	job := &Job{
		ID:      jobID,
		Type:    "webhook.delivery",
		Payload: []byte("test payload"),
		Status:  JobStatusQueued,
	}

	err = store.CreateJob(ctx, job)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_GetJobByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{db: db}
	ctx := context.Background()

	jobID := uuid.New()
	now := time.Now()

	// Only mock the SELECT - no INSERT needed!
	rows := sqlmock.NewRows([]string{
		"id", "type", "payload", "status", "created_at", "started_at", "completed_at",
	}).AddRow(
		jobID,
		"webhook.delivery",
		[]byte("test payload"),
		JobStatusQueued,
		now,
		nil,
		nil,
	)

	mock.ExpectQuery("SELECT (.+) FROM jobs WHERE id").
		WithArgs(jobID).
		WillReturnRows(rows)

	retrieved, err := store.GetJobByID(ctx, jobID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	assert.Equal(t, jobID, retrieved.ID)
	assert.Equal(t, "webhook.delivery", retrieved.Type)
	assert.Equal(t, []byte("test payload"), retrieved.Payload)
	assert.Equal(t, JobStatusQueued, retrieved.Status)
	assert.NotZero(t, retrieved.CreatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_GetJobByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{db: db}
	ctx := context.Background()

	jobID := uuid.New()

	// Expectation for SELECT (GetJobByID) with no rows returned
	mock.ExpectQuery("SELECT (.+) FROM jobs WHERE id").
		WithArgs(jobID).
		WillReturnError(sql.ErrNoRows)

	retrieved, err := store.GetJobByID(ctx, jobID)
	assert.Error(t, err)
	assert.Nil(t, retrieved)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_MarkJobAsQueued(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{db: db}
	ctx := context.Background()
	jobID := uuid.New()

	mock.ExpectExec("UPDATE jobs SET status").
		WithArgs(JobStatusQueued, jobID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.MarkJobAsQueued(ctx, jobID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_MarkJobAsProcessing(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{db: db}
	ctx := context.Background()
	jobID := uuid.New()

	mock.ExpectExec("UPDATE jobs SET status").
		WithArgs(JobStatusProcessing, jobID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.MarkJobAsProcessing(ctx, jobID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_MarkJobAsCompleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{db: db}
	ctx := context.Background()
	jobID := uuid.New()

	mock.ExpectExec("UPDATE jobs SET status").
		WithArgs(JobStatusCompleted, jobID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.MarkJobAsCompleted(ctx, jobID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_MarkJobAsFailed(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{db: db}
	ctx := context.Background()
	jobID := uuid.New()

	mock.ExpectExec("UPDATE jobs SET status").
		WithArgs(JobStatusFailed, jobID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.MarkJobAsFailed(ctx, jobID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
