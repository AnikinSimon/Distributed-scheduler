package cases

import (
	"context"
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/port/repo"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

func (s *SchedulerCase) Start(ctx context.Context) error {

	for {
		select {
		case <-time.NewTimer(s.interval).C:
			if err := s.redisMutex.TryLock(); err != nil {
				s.logger.Warn("Failed to lock", zap.Error(err))
				continue
			}
			s.logger.Info("Lock acquired")
			s.logger.Info("Tick started")
			if err := s.tick(ctx); err != nil {
				s.logger.Error("Error tick", zap.Error(err))
			}
		case <-ctx.Done():
			if ok, err := s.redisMutex.Unlock(); !ok && err != nil {
				s.logger.Error("Error stopping redis mutex", zap.Error(err))
			}
			return ctx.Err()
		}
	}
}

func (s *SchedulerCase) tick(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobs, err := s.jobsRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("listing jobs: %w", err)
	}

	repoJobs := make(map[string]*entity.Job, len(jobs))
	for _, job := range jobs {
		j := repo.JobEntityFromDTO(job)
		repoJobs[job.Id.String()] = j
		s.logger.Info("Job", zap.Any("job", j))
	}

	for jobId, job := range s.running {
		if _, ok := repoJobs[jobId]; !ok {
			s.logger.Debug("Stop deleted job", zap.String("jobId", jobId))
			job.Cancel()
			delete(s.running, jobId)
		}
	}

	now := time.Now().UnixMilli()

	var updates []*repo.JobDTO

	for jobId, job := range repoJobs {
		if _, ok := s.running[jobId]; ok {
			s.logger.Debug("Skip running job", zap.String("jobId", jobId))
			continue
		}
		s.logger.Info("Start job",
			zap.String("jobId", jobId),
			zap.Int("kind", int(job.Kind)),
			zap.Any("payload", job.Payload),
			zap.String("status", string(job.Status)),
		)
		if job.Kind == entity.JobKindInterval {
			s.logger.Info("Start interval job", zap.String("jobId", jobId), zap.Duration("interval", *job.Interval))
			if now > job.Interval.Milliseconds()+job.LastFinishedAt {
				go s.runJob(ctx, job)
			}
		} else {
			if job.Once != nil {
				s.logger.Info("Start job once", zap.String("jobId", jobId), zap.Int64("Once", *job.Once), zap.Int64("Now", now),
					zap.Int64("Diff", *job.Once-now))
				if job.LastFinishedAt == 0 && now > *job.Once {
					go s.runJob(ctx, job)
				}
			} else {
				s.logger.Error("No jobOnce info", zap.String("jobId", jobId))
			}
		}
		job.Status = entity.JobStatusRunning
		updates = append(updates, repo.JobDTOFromEntity(job))
	}

	if err := s.jobsRepo.Upsert(ctx, updates); err != nil {
		return fmt.Errorf("upserting jobs failed: %w", err)
	}

	return nil
}

func (s *SchedulerCase) runJob(ctx context.Context, job *entity.Job) {
	s.logger.Info("Start running job", zap.Any("job", job))
	ctx, cancel := context.WithCancel(ctx)
	s.running[job.Id.String()] = &entity.RunningJob{
		Job:    job,
		Cancel: cancel,
	}

	if s.publisher != nil {
		if err := s.publisher.Publish(ctx, job); err != nil {
			s.logger.Error("Failed to publish job", zap.String("jobId", job.Id.String()), zap.Error(err))
		}
	} else {
		s.logger.Info("Skip publishing job PUBLISHER IF NIL", zap.String("jobId", job.Id.String()))
	}

}

func (s *SchedulerCase) HandleJobCompletion(ctx context.Context, jobID uuid.UUID, status string, finishedAt int64) error {
	job, err := s.jobsRepo.Read(ctx, jobID)
	if err != nil {
		return fmt.Errorf("getting job: %w", err)
	}

	s.logger.Info("Job completion init",
		zap.String("job_id", jobID.String()),
		zap.String("status", status),
		zap.Int64("finished_at", finishedAt))

	switch status {
	case entity.JobStatusCompleted:
		job.Status = entity.JobStatusCompleted
	case entity.JobStatusFailed:
		job.Status = entity.JobStatusFailed
	default:
		return fmt.Errorf("unknown job status: %s", status)
	}

	job.LastFinishedAt = finishedAt

	s.mu.Lock()
	defer s.mu.Unlock()

	if runningJob, ok := s.running[jobID.String()]; ok {
		runningJob.Cancel()
		delete(s.running, jobID.String())
	}

	if err := s.jobsRepo.Upsert(ctx, []*repo.JobDTO{job}); err != nil {
		s.logger.Info("Failed to upsert")
		return fmt.Errorf("upserting job failed: %w", err)
	}

	s.logger.Info("Job completion handled",
		zap.String("job_id", jobID.String()),
		zap.String("status", status),
		zap.Int64("finished_at", finishedAt))

	return nil

}
