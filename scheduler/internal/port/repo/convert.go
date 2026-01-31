package repo

import (
	"time"

	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
)

func JobDTOFromEntity(job *entity.Job) *JobDTO {
	dto := &JobDTO{}
	dto.ID = job.ID
	dto.Kind = int(job.Kind)
	switch job.Kind {
	case entity.JobKindOnce:
		dto.Once = job.Once
	case entity.JobKindInterval:
		dto.Interval = int64(job.Interval.Seconds())
	}
	dto.Status = string(job.Status)
	dto.Payload = *job.Payload
	dto.LastFinishedAt = job.LastFinishedAt

	return dto
}

func JobEntityFromDTO(job *JobDTO) *entity.Job {
	ent := &entity.Job{}
	ent.ID = job.ID
	ent.Kind = entity.JobKind(job.Kind)
	switch ent.Kind {
	case entity.JobKindOnce:
		ent.Once = job.Once
	case entity.JobKindInterval:
		interval := time.Duration(job.Interval) * time.Second
		ent.Interval = &interval
	}
	ent.LastFinishedAt = job.LastFinishedAt
	ent.Payload = &job.Payload
	ent.Status = entity.JobStatus(job.Status)

	return ent
}
