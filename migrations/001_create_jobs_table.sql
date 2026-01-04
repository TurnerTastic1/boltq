-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    payload BYTEA NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Create index on status for efficient queries
CREATE INDEX idx_jobs_status ON jobs(status);

-- Create index on created_at for sorting
CREATE INDEX idx_jobs_created_at ON jobs(created_at DESC);

-- Create index for finding stale jobs
CREATE INDEX idx_jobs_stale ON jobs(status, updated_at) WHERE status = 'processing';