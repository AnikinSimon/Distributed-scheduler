package entity

import (
	"context"
	"github.com/google/uuid"
	"time"
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
	Id             uuid.UUID
	Interval       *time.Duration
	LastFinishedAt int64
	Once           *int64
	Payload        *map[string]interface{}
	Status         JobStatus
	Kind           JobKind
}

type RunningJob struct {
	*Job

	Cancel context.CancelFunc
}
