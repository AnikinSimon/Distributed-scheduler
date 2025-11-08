package repo

import (
	"context"
	"errors"
	"github.com/google/uuid"
)

type Jobs interface {
	Create(ctx context.Context, job *JobDTO) error
	Read(ctx context.Context, jobID uuid.UUID) (*JobDTO, error)
	Update(ctx context.Context, job *JobDTO) error
	Delete(ctx context.Context, jobID uuid.UUID) error
	List(ctx context.Context) ([]*JobDTO, error)
}

var (
	ErrJobIdNotFound = errors.New("job not found")
)

type JobDTO struct {
	// Interface specific entity-like struct
	Id             uuid.UUID
	Interval       int64
	LastFinishedAt int64
	Once           *int64
	Payload        map[string]any
	Status         string
	Kind           int
}
