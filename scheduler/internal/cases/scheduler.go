package cases

import (
	"context"
	"fmt"
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
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	job.Id = id
	job.CreatedAt = time.Now().UnixMilli()
	job.Status = "queued"

	jobDto := repo.JobDTO(*job)
	return id.String(), r.jobsRepo.Create(ctx, &jobDto)
}

func (r *SchedulerCase) Get(ctx context.Context, jobID string) (*entity.Job, error) {
	id, err := uuid.Parse(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID format: %w", err)
	}

	jobDTO, err := r.jobsRepo.Read(ctx, id)

	if err != nil {
		return nil, err
	}

	job := entity.Job(*jobDTO)

	return &job, nil
}

func (r *SchedulerCase) Delete(ctx context.Context, jobID string) error {
	id, err := uuid.Parse(jobID)
	if err != nil {
		return fmt.Errorf("invalid job ID format: %w", err)
	}

	return r.jobsRepo.Delete(ctx, id)
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
