package store

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPostgresStore_CreateAndGetJob(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{db: db}
	ctx := context.Background()

	jobID := uuid.New()
	now := time.Now()

	// Expectation for INSERT (CreateJob)
	mock.ExpectExec("INSERT INTO jobs").
		WithArgs(sqlmock.AnyArg(), "webhook.delivery", []byte("test payload"), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	job := &Job{
		ID:      jobID,
		Type:    "webhook.delivery",
		Payload: []byte("test payload"),
		Status:  JobStatusQueued,
	}

	// Insert the job
	err = store.CreateJob(ctx, job)
	assert.NoError(t, err)

	// Expectation for SELECT (GetJobByID)
	rows := sqlmock.NewRows([]string{"id", "type", "payload", "status", "created_at", "started_at", "completed_at"}).
		AddRow(job.ID, job.Type, job.Payload, job.Status, now, nil, nil)

	mock.ExpectQuery("SELECT (.+) FROM jobs WHERE id").
		WithArgs(job.ID).
		WillReturnRows(rows)

	// Retrieve the job
	retrieved, err := store.GetJobByID(ctx, job.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	assert.Equal(t, job.ID, retrieved.ID)
	assert.Equal(t, job.Type, retrieved.Type)
	assert.Equal(t, job.Payload, retrieved.Payload)
	assert.Equal(t, job.Status, retrieved.Status)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
