package repo

import (
	"context"
	"github.com/google/uuid"
)

type Jobs interface {
	Create(ctx context.Context, job *JobDTO) error
	Read(ctx context.Context, jobID uuid.UUID) (*JobDTO, error)
	Update(ctx context.Context, job *JobDTO) error
	Delete(ctx context.Context, jobID uuid.UUID) error
	List(ctx context.Context, status string) ([]*JobDTO, error)
}

type JobDTO struct {
	// Interface specific entity-like struct
	CreatedAt      int64
	Id             uuid.UUID
	Interval       *string
	LastFinishedAt int64
	Once           *string
	Payload        map[string]interface{}
	Status         string
}
