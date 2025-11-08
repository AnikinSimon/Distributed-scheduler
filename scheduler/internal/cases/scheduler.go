package cases

import (
	"context"
	"errors"
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/port/publisher"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/port/repo"
	"go.uber.org/zap"
	"time"

	"github.com/google/uuid"
)

var (
	ErrJobNotFound = errors.New("job not found")
)

type SchedulerCase struct {
	jobsRepo  repo.Jobs
	logger    *zap.Logger
	interval  time.Duration
	running   map[string]*entity.RunningJob
	publisher publisher.Publisher
}

func NewSchedulerCase(jobsRepo repo.Jobs, logger *zap.Logger, interval time.Duration) *SchedulerCase {
	schedulerCase := &SchedulerCase{
		jobsRepo: jobsRepo,
		logger:   logger,
		running:  make(map[string]*entity.RunningJob),
		interval: time.Minute,
	}
	return schedulerCase
}

func (s *SchedulerCase) Create(ctx context.Context, job *entity.Job) (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	job.Id = id
	job.Status = "queued"

	s.logger.Info("Job Creating", zap.String("job_id", job.Id.String()), zap.Any("job", job))

	jobDto := repo.JobDTOFromEntity(job)

	return id.String(), s.jobsRepo.Create(ctx, jobDto)
}

func (s *SchedulerCase) Get(ctx context.Context, jobID string) (*entity.Job, error) {
	id, err := uuid.Parse(jobID)
	if err != nil {
		s.logger.Error("failed to parse job id", zap.String("job_id", jobID), zap.Error(err))
		return nil, ErrJobNotFound
	}

	jobDTO, err := s.jobsRepo.Read(ctx, id)

	if err != nil {
		if errors.Is(err, repo.ErrJobIdNotFound) {
			return nil, ErrJobNotFound
		}
		s.logger.Error("failed to read job", zap.String("job_id", id.String()), zap.String("job_id", jobID), zap.Error(err))
		return nil, err
	}

	job := repo.JobEntityFromDTO(jobDTO)

	return job, nil
}

func (s *SchedulerCase) Delete(ctx context.Context, jobID string) error {
	id, err := uuid.Parse(jobID)
	if err != nil {
		return fmt.Errorf("invalid job ID format: %w", err)
	}

	return s.jobsRepo.Delete(ctx, id)
}

func (s *SchedulerCase) List(ctx context.Context, status string) ([]*entity.Job, error) {
	jobsDTO, err := s.jobsRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	jobs := make([]*entity.Job, len(jobsDTO))

	for i := range jobsDTO {
		jobs[i] = repo.JobEntityFromDTO(jobsDTO[i])
	}
	return jobs, nil
}
