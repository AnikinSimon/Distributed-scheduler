package publisher

import (
	"context"

	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
)

type Publisher interface {
	Publish(ctx context.Context, job *entity.Job) error
}

type JobDTO struct {
	ID             string  `json:"id"`
	Kind           int     `json:"kind"` // JobKind as int
	Status         string  `json:"status"`
	Interval       *string `json:"interval,omitempty"` // duration as string
	Once           *int64  `json:"once,omitempty"`
	LastFinishedAt int64   `json:"lastFinishedAt"`
	Payload        any     `json:"payload"`
}
