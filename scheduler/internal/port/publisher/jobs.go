package publisher

import (
	"fmt"

	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
)

func JobDTOFromEntity(job *entity.Job) (*JobDTO, error) {
	dto := &JobDTO{}
	dto.ID = job.ID.String()
	dto.Kind = int(job.Kind)
	switch job.Kind {
	case entity.JobKindOnce:
		dto.Once = job.Once
	case entity.JobKindInterval:
		durationStr := job.Interval.String()
		dto.Interval = &durationStr
	default:
		return nil, fmt.Errorf("undefined job kind %d", job.Kind)
	}
	dto.Status = string(job.Status)

	dto.Payload = job.Payload
	dto.LastFinishedAt = job.LastFinishedAt

	return dto, nil
}
