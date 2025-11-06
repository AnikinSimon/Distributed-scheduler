package cases

import (
	"context"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/port/repo"
	"go.uber.org/zap"
	"time"

	"github.com/google/uuid"
)

type SchedulerCase struct {
	jobsRepo repo.Jobs
	logger   *zap.Logger
}

func NewSchedulerCase(jobsRepo repo.Jobs, logger *zap.Logger) *SchedulerCase {
	return &SchedulerCase{
		jobsRepo: jobsRepo,
		logger:   logger,
	}
}

func (r *SchedulerCase) Create(ctx context.Context, job *entity.Job) (string, error) {
	job.Id = uuid.NewString()
	job.CreatedAt = time.Now().UnixMilli()
	job.Status = "queued"

	jobDto := repo.JobDTO(*job)
	return job.Id, r.jobsRepo.Create(ctx, &jobDto)
}

func (r *SchedulerCase) Get(ctx context.Context, jobID string) (*entity.Job, error) {
	jobDTO, err := r.jobsRepo.Read(ctx, jobID)

	if err != nil {
		return nil, err
	}

	job := entity.Job(*jobDTO)

	return &job, nil
}

func (r *SchedulerCase) Delete(ctx context.Context, jobID string) error {
	return r.jobsRepo.Delete(ctx, jobID)
}

func (r *SchedulerCase) List(ctx context.Context, status string) ([]*entity.Job, error) {
	jobsDTO, err := r.jobsRepo.List(ctx, status)
	if err != nil {
		return nil, err
	}

	jobs := make([]*entity.Job, len(jobsDTO))

	for i := range jobsDTO {
		jb := entity.Job(*jobsDTO[i])
		jobs[i] = &jb
	}
	return jobs, nil
}
