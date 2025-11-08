package cases

import (
	"context"
	"time"
)

func (s *SchedulerCase) Start(ctx context.Context) error {
	for {
		select {
		case <-time.NewTimer(s.interval).C:
			if err := s.tick(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *SchedulerCase) tick(ctx context.Context) error {
	//jobs, err := s.jobsRepo.List(ctx)
	//if err != nil {
	//	return fmt.Errorf("listing jobs: %w", err)
	//}
	//
	//repoJobs := make(map[string]*entity.Job, len(jobs))
	//for _, job := range jobs {
	//	j := entity.Job(*job)
	//	repoJobs[job.Id.String()] = &j
	//}
	//
	//for jobId, job := range s.running {
	//	if _, ok := s.running[jobId]; !ok {
	//		s.logger.Debug("Stop deleted job", zap.String("jobId", jobId))
	//		job.Cancel()
	//		delete(s.running, jobId)
	//	}
	//}

	//now := time.Now().UnixMilli()
	//
	//var updates []*entity.Job
	//
	//for jobId, job := range repoJobs {
	//	if _, ok := s.running[jobId]; ok {
	//		s.logger.Debug("Skip running job", zap.String("jobId", jobId))
	//		continue
	//	}
	//	if job.Interval != nil {
	//		if now > job.Interval {
	//		}
	//	}
	//}

	return nil
}
