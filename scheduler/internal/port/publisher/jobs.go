package publisher

import (
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
)

func JobDTOFromEntity(job *entity.Job) (*JobDTO, error) {
	dto := &JobDTO{}
	dto.ID = job.Id.String()
	dto.Kind = int(job.Kind)
	if job.Kind == entity.JobKindOnce {
		dto.Once = job.Once
	} else if job.Kind == entity.JobKindInterval {
		durationStr := job.Interval.String()
		dto.Interval = &durationStr
	} else {
		return nil, fmt.Errorf("undefined job kind %d", job.Kind)
	}
	dto.Status = string(job.Status)

	dto.Payload = job.Payload
	dto.LastFinishedAt = job.LastFinishedAt

	return dto, nil
}
