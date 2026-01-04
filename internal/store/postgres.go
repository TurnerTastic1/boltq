package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (ps *PostgresStore) Close() error {
	return ps.db.Close()
}

func (ps *PostgresStore) CreateJob(ctx context.Context, job *Job) error {
	query := `
		INSERT INTO jobs (id, type, payload, status)
		VALUES ($1, $2, $3, $4)
	`

	_, err := ps.db.ExecContext(ctx, query, job.ID, job.Type, job.Payload, job.Status)

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

func (ps *PostgresStore) GetJobByID(ctx context.Context, id uuid.UUID) (*Job, error) {
	query := `
		SELECT id, type, payload, status, created_at, started_at, completed_at
		FROM jobs
		WHERE id = $1
	`

	job := &Job{}
	err := ps.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.CreatedAt,
		&job.StartedAt,
		&job.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found for ID %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job by ID: %w", err)
	}

	return job, nil
}

func (ps *PostgresStore) MarkJobAsQueued(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE jobs
		SET status = $1, started_at = NULL, completed_at = NULL
		WHERE id = $2
	`

	_, err := ps.db.ExecContext(ctx, query, JobStatusQueued, id)
	if err != nil {
		return fmt.Errorf("failed to mark job as queued: %w", err)
	}

	return nil
}

func (ps *PostgresStore) MarkJobAsProcessing(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE jobs
		SET status = $1, started_at = NOW()
		WHERE id = $2
	`

	_, err := ps.db.ExecContext(ctx, query, JobStatusProcessing, id)
	if err != nil {
		return fmt.Errorf("failed to mark job as processing: %w", err)
	}

	return nil
}

func (ps *PostgresStore) MarkJobAsCompleted(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE jobs
		SET status = $1, completed_at = NOW()
		WHERE id = $2
	`

	_, err := ps.db.ExecContext(ctx, query, JobStatusCompleted, id)
	if err != nil {
		return fmt.Errorf("failed to mark job as completed: %w", err)
	}

	return nil
}

func (ps *PostgresStore) MarkJobAsFailed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE jobs
		SET status = $1, completed_at = NOW()
		WHERE id = $2
	`

	_, err := ps.db.ExecContext(ctx, query, JobStatusFailed, id)
	if err != nil {
		return fmt.Errorf("failed to mark job as failed: %w", err)
	}

	return nil
}
