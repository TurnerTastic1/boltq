package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connString string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresStore{db: db}, nil
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
