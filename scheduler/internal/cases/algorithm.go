package cases

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/port/repo"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *SchedulerCase) Start(ctx context.Context) error {
	for {
		select {
		case <-time.NewTimer(s.interval).C:
			mutexVal := s.redisMutex.Value()
			s.logger.Info("Redis mutex value", zap.String("Value", mutexVal))
			if mutexVal == "" {
				if err := s.redisMutex.TryLockContext(ctx); err != nil {
					s.logger.Warn("Failed to lock", zap.Error(err))
					continue
				}
			} else {
				if ok, err := s.redisMutex.ExtendContext(ctx); !ok || err != nil {
					s.logger.Warn("Failed to extend lock", zap.Error(err))
					continue
				}
			}

			s.logger.Info("Lock acquired")
			s.logger.Info("Tick started")
			if err := s.tick(ctx); err != nil {
				s.logger.Error("Error tick", zap.Error(err))
			}
		case <-ctx.Done():
			if ok, err := s.redisMutex.UnlockContext(ctx); !ok && err != nil {
				s.logger.Error("Error stopping redis mutex", zap.Error(err))
			}
			return ctx.Err()
		}
	}
}

func (s *SchedulerCase) removeDeletedJobs(repoJobs map[string]*entity.Job) {
	for jobID, job := range s.running {
		if _, ok := repoJobs[jobID]; !ok {
			s.logger.Debug("Stop deleted job", zap.String("jobID", jobID))
			job.Cancel()
			delete(s.running, jobID)
		}
	}
}

func (s *SchedulerCase) tick(ctx context.Context) error {
	jobs, err := s.jobsRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("listing jobs: %w", err)
	}

	repoJobs := make(map[string]*entity.Job, len(jobs))
	for _, job := range jobs {
		j := repo.JobEntityFromDTO(job)
		repoJobs[job.ID.String()] = j
		s.logger.Info("Job", zap.Any("job", j))
	}

	s.removeDeletedJobs(repoJobs)

	now := time.Now().UnixMilli()

	var updates []*repo.JobDTO
	wg := &sync.WaitGroup{}

	for jobID, job := range repoJobs {
		if _, ok := s.running[jobID]; ok {
			s.logger.Debug("Skip running job", zap.String("jobID", jobID))
			continue
		}
		s.logger.Info("Start job",
			zap.String("jobID", jobID),
			zap.Int("kind", int(job.Kind)),
			zap.Any("payload", job.Payload),
			zap.String("status", string(job.Status)),
		)
		switch job.Kind {
		case entity.JobKindInterval:
			s.logger.Info("Start interval job", zap.String("jobID", jobID), zap.Duration("interval", *job.Interval))
			if now > job.Interval.Milliseconds()+job.LastFinishedAt {
				wg.Add(1)
				go s.runJob(ctx, job, wg)
			}
		default:
			if job.Once != nil {
				s.logger.Info(
					"Start job once",
					zap.String("jobID", jobID),
					zap.Int64("Once", *job.Once),
					zap.Int64("Now", now),
					zap.Int64("Diff", *job.Once-now),
				)
				if job.LastFinishedAt == 0 && now > *job.Once {
					wg.Add(1)
					go s.runJob(ctx, job, wg)
				}
			} else {
				s.logger.Error("No jobOnce info", zap.String("jobID", jobID))
			}
		}
		job.Status = entity.JobStatusRunning
		updates = append(updates, repo.JobDTOFromEntity(job))
	}
	wg.Wait()
	if err := s.jobsRepo.Upsert(ctx, updates); err != nil {
		return fmt.Errorf("upserting jobs failed: %w", err)
	}

	return nil
}

func (s *SchedulerCase) runJob(ctx context.Context, job *entity.Job, wg *sync.WaitGroup) {
	defer wg.Done()
	s.logger.Info("Start running job", zap.Any("job", job))
	ctx, cancel := context.WithCancel(ctx)
	s.running[job.ID.String()] = &entity.RunningJob{
		Job:    job,
		Cancel: cancel,
	}

	if s.publisher != nil {
		exec := &entity.Execution{
			ID:       uuid.New(),
			JobID:    job.ID,
			Status:   entity.ExecutionStatusQueued,
			QueuedAt: time.Now().UnixMilli(),
		}
		if err := s.executionsRepo.Upsert(ctx, exec); err != nil {
			s.logger.Error("Failed to upsert job", zap.String("job_id", job.ID.String()), zap.Error(err))
			return
		}

		if err := s.publisher.Publish(ctx, job); err != nil {
			s.logger.Error("Failed to publish job", zap.String("jobID", job.ID.String()), zap.Error(err))
		}
	} else {
		s.logger.Info("Skip publishing job PUBLISHER IF NIL", zap.String("jobID", job.ID.String()))
	}
}

func (s *SchedulerCase) HandleJobCompletion(
	ctx context.Context,
	jobID uuid.UUID,
	status string,
	finishedAt int64,
) error {
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
