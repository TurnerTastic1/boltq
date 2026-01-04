package store

import (
	"time"

	"github.com/google/uuid"
)

// Job represents a job in the jobs table
type Job struct {
	ID          uuid.UUID  `db:"id"`
	Type        string     `db:"type"`
	Payload     []byte     `db:"payload"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	StartedAt   *time.Time `db:"started_at"`
	CompletedAt *time.Time `db:"completed_at"`
}

// Job status constants
const (
	JobStatusQueued     = "queued"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)
