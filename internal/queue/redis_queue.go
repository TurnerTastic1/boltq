// Package queue provides a Redis-based job queue management for BoltQ.
// It handles enqueung, dequeuing, and queue monitoring using Redis.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	QueueKeyPrefix   = "boltq:queue:"
	JobKeyPrefix     = "boltq:job:"
	redisPingTimeout = 5 * time.Second
)

// RedisQueue manages job queues using Redis as the backing store.
type RedisQueue struct {
	client *redis.Client
}

// JobMessage represents a lightweight job reference in the queue.
type JobMessage struct {
	JobID uuid.UUID `json:"job_id"`
	Type  string    `json:"type"`
}

// NewRedisQueue initializes a new RedisQueue with the given Redis connection parameters.
func NewRedisQueue(addr, password string, db int) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisQueue{client: client}, nil
}

// Enqueue adds a job reference to the Redis queue.
func (rq *RedisQueue) Enqueue(ctx context.Context, jobID uuid.UUID, jobType string) error {
	msg := JobMessage{
		JobID: jobID,
		Type:  jobType,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	queueKey := QueueKeyPrefix + jobType
	return rq.client.RPush(ctx, queueKey, data).Err()
}

// Dequeue retrieves the next job reference from the Redis queue.
// Workers should call this with their specific job type.
// Returns nil if no job is available within the timeout.
func (rq *RedisQueue) Dequeue(ctx context.Context, jobType string, timeout time.Duration) (*JobMessage, error) {
	queueKey := QueueKeyPrefix + jobType

	result, err := rq.client.BLPop(ctx, timeout, queueKey).Result()
	if err == redis.Nil {
		// No job available within the timeout
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected result format from Redis with length: %d", len(result))
	}

	var msg JobMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// GetQueueLength returns the number of jobs waiting in a specific queue.
func (rq *RedisQueue) GetQueueLength(ctx context.Context, jobType string) (int64, error) {
	queueKey := QueueKeyPrefix + jobType
	length, err := rq.client.LLen(ctx, queueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}
	return length, nil
}

// Close closes the Redis client connection.
func (rq *RedisQueue) Close() error {
	return rq.client.Close()
}
