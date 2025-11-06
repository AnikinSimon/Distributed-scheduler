package handler

import (
	"context"
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/cases"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/input/http/gen"
)

var _ gen.StrictServerInterface = (*Server)(nil)

type Server struct {
	schedulerCase *cases.SchedulerCase
}

func NewServer(schCase *cases.SchedulerCase) *Server {
	return &Server{
		schedulerCase: schCase,
	}
}

// Create a new job
// (POST /jobs)
func (r *Server) PostJobs(ctx context.Context, request gen.PostJobsRequestObject) (gen.PostJobsResponseObject, error) {
	jobID, err := r.schedulerCase.Create(ctx, toEntityJob(request.Body))

	if err != nil {
		return nil, err // 500
	}

	return gen.PostJobs201JSONResponse(jobID), nil
}

// List jobs
// (GET /jobs)
func (r *Server) GetJobs(ctx context.Context, request gen.GetJobsRequestObject) (gen.GetJobsResponseObject, error) {
	jobsEntity, err := r.schedulerCase.List(ctx, string(*request.Params.Status))
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	// Преобразуем entity.Job в gen.Job
	jobs := make([]gen.Job, 0, len(jobsEntity))
	for _, jobEntity := range jobsEntity {
		job := gen.Job{
			Id:             jobEntity.Id.String(),
			Status:         gen.Status(jobEntity.Status),
			CreatedAt:      jobEntity.CreatedAt,
			LastFinishedAt: jobEntity.LastFinishedAt,
			Payload:        jobEntity.Payload,
		}

		// Добавляем опциональные поля
		if jobEntity.Interval != nil {
			job.Interval = jobEntity.Interval

		} else {
			job.Once = jobEntity.Once
		}

		jobs = append(jobs, job)
	}

	return gen.GetJobs200JSONResponse(jobs), nil
}

// Delete a job
// (DELETE /jobs/{job_id})
func (r *Server) DeleteJobsJobId(ctx context.Context, request gen.DeleteJobsJobIdRequestObject) (gen.DeleteJobsJobIdResponseObject, error) {
	err := r.schedulerCase.Delete(ctx, request.JobId)
	if err != nil {
		return gen.DeleteJobsJobId404Response{}, err
	}

	return gen.DeleteJobsJobId204Response{}, nil
}

// Get job details
// (GET /jobs/{job_id})
func (r *Server) GetJobsJobId(ctx context.Context, request gen.GetJobsJobIdRequestObject) (gen.GetJobsJobIdResponseObject, error) {
	job, err := r.schedulerCase.Get(ctx, request.JobId)

	if err != nil {
		return gen.GetJobsJobId404Response{}, nil
	}

	return fromEntityJobGetResponse(job), nil
}

// Get job executions
// (GET /jobs/{job_id}/executions)
func (r *Server) GetJobsJobIdExecutions(ctx context.Context, request gen.GetJobsJobIdExecutionsRequestObject) (gen.GetJobsJobIdExecutionsResponseObject, error) {
	panic("not implemented") // TODO: Implement
}
