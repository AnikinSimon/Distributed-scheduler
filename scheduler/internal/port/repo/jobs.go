package repo

import (
	"context"
)

type Jobs interface {
	Create(ctx context.Context, job *JobDTO) error
	Read(ctx context.Context, jobID string) (*JobDTO, error)
	Update(ctx context.Context, job *JobDTO) error
	Delete(ctx context.Context, jobID string) error
	List(ctx context.Context, status string) ([]*JobDTO, error)
}

type JobDTO struct {
	// Interface specific entity-like struct
	CreatedAt      int64
	Id             string
	Interval       *string
	LastFinishedAt int64
	Once           *string
	Payload        map[string]interface{}
	Status         string
}
