package handler

import (
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/input/http/gen"
)

func toEntityJob(job *gen.JobCreate) *entity.Job {

	res := &entity.Job{}
	res.Interval = job.Interval
	res.Once = job.Once
	if job.Payload != nil {
		res.Payload = *job.Payload
	}

	return res

}

func fromEntityJobGetResponse(job *entity.Job) gen.GetJobsJobId200JSONResponse {
	genJob := gen.Job{
		Id:             job.Id,
		Once:           job.Once,
		Payload:        job.Payload,
		Interval:       job.Interval,
		Status:         gen.Status(job.Status),
		LastFinishedAt: job.LastFinishedAt,
		CreatedAt:      job.CreatedAt,
	}

	return gen.GetJobsJobId200JSONResponse(genJob)
}
