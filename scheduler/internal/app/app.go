package app

import (
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/config"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/adapter/repo/postgres"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/cases"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/input/http/handler"
)

func Start(cfg config.Config) error {
	// TODO: Create jobs repo
	jobsRepo := postgres.NewJobsRepo() // TODO: pg config

	scheduler := cases.NewSchedulerCase(jobsRepo)
	handler.NewServer(scheduler)

	// TODO: Add http server.
	return nil
}
