package repo

import (
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
	"time"
)

func JobDTOFromEntity(job *entity.Job) *JobDTO {
	dto := &JobDTO{}
	dto.Id = job.Id
	dto.Kind = int(job.Kind)
	if job.Kind == entity.JobKindOnce {
		dto.Once = job.Once
	} else if job.Kind == entity.JobKindInterval {
		dto.Interval = int64(job.Interval.Seconds())
	}
	dto.Status = string(job.Status)
	dto.Payload = *job.Payload
	dto.LastFinishedAt = job.LastFinishedAt

	return dto
}

func JobEntityFromDTO(job *JobDTO) *entity.Job {
	ent := &entity.Job{}
	ent.Id = job.Id
	ent.Kind = entity.JobKind(job.Kind)
	if ent.Kind == entity.JobKindOnce {
		ent.Once = job.Once
	} else if ent.Kind == entity.JobKindInterval {
		interval := time.Duration(job.Interval) * time.Second
		ent.Interval = &interval
	}
	ent.LastFinishedAt = job.LastFinishedAt
	ent.Payload = &job.Payload
	ent.Status = entity.JobStatus(job.Status)

	return ent
}
