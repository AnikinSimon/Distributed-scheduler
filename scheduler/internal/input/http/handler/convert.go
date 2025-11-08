package handler

import (
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/input/http/gen"
	"time"
)

func toEntityJob(job *gen.JobCreate) (*entity.Job, error) {
	entityJob := &entity.Job{}
	if job.Interval != nil {
		interval, err := time.ParseDuration(*job.Interval)
		if err != nil {
			return nil, fmt.Errorf("failed to parse interval: %w", err)
		}
		entityJob.Interval = &interval
		entityJob.Kind = entity.JobKindInterval
	} else if job.Once != nil {
		once, err := time.Parse(time.RFC3339, *job.Once)
		if err != nil {
			return nil, fmt.Errorf("failed to parse once datetime: %w", err)
		}
		onceSeconds := once.UnixMilli()
		entityJob.Once = &onceSeconds
		entityJob.Kind = entity.JobKindOnce
	}
	entityJob.Payload = job.Payload

	return entityJob, nil
}

func fromEntityJobGetGenJob(job *entity.Job) gen.Job {
	genJob := gen.Job{}

	genJob.Id = job.Id.String()
	if job.Kind == entity.JobKindInterval {
		intervalString := job.Interval.String()
		genJob.Interval = &intervalString
	} else if job.Kind == entity.JobKindOnce {
		onceString := time.UnixMilli(*job.Once).String()
		genJob.Once = &onceString
	}

	genJob.Payload = *job.Payload
	genJob.Status = gen.Status(job.Status)
	genJob.LastFinishedAt = job.LastFinishedAt

	return genJob
}
