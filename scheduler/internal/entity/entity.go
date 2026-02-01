package entity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type JobKind int

const (
	JobKindUndefined = iota
	JobKindOnce
	JobKindInterval
)

type JobStatus string

const (
	JobStatusQueued    = "queued"
	JobStatusRunning   = "running"
	JobStatusFailed    = "failed"
	JobStatusCompleted = "completed"
)

type Job struct {
	ID             uuid.UUID
	Interval       *time.Duration
	LastFinishedAt int64
	Once           *int64
	Payload        *map[string]any
	Status         JobStatus
	Kind           JobKind
}

type RunningJob struct {
	*Job

	Cancel context.CancelFunc
}

type ExecutionStatus string

const (
	ExecutionStatusQueued    = "queued"
	ExecutionStatusRunning   = "running"
	ExecutionStatusFailed    = "failed"
	ExecutionStatusCompleted = "completed"
)

type Execution struct {
	ID         uuid.UUID
	JobID      uuid.UUID
	WorkerID   string
	Status     ExecutionStatus
	QueuedAt   int64
	StartedAt  *int64
	FinishedAt *int64
	Error      *ExecutionError
}

type ExecutionError struct {
	Details string `json:"details"`
}

type ListExecutionFilter struct {
	IDs    []uuid.UUID
	JobIDs []uuid.UUID
	Status *ExecutionStatus
}
