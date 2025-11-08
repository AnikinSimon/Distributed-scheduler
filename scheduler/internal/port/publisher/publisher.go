package publisher

import (
	"context"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
)

type Publisher interface {
	Publish(ctx context.Context, job *entity.Job) error
}
