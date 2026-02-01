package repo

import (
	"context"

	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
	"github.com/google/uuid"
)

type Executions interface {
	Upsert(ctx context.Context, exec *entity.Execution) error
	Read(ctx context.Context, ID uuid.UUID) (*entity.Execution, error)
	List(ctx context.Context, filter *entity.ListExecutionFilter) ([]*entity.Execution, error)
}
