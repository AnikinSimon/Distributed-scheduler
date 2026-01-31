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
	Payload        any
	Status         JobStatus
	Kind           JobKind
}

type RunningJob struct {
	*Job

	Cancel context.CancelFunc
}
