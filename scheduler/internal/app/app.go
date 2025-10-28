package app

import (
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/config"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/adapter/repo/postgres"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/cases"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/input/http/gen"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/input/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

func Start(cfg config.Config) error {
	// TODO: Create jobs repo
	jobsRepo := postgres.NewJobsRepo(cfg.Storage) // TODO: pg config

	scheduler := cases.NewSchedulerCase(jobsRepo)

	server := handler.NewServer(scheduler)

	strictHandler := gen.NewStrictHandler(server, nil)

	// Настраиваем Chi роутер
	router := chi.NewRouter()

	// Добавляем middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	gen.HandlerWithOptions(strictHandler, gen.ChiServerOptions{
		BaseRouter: router,
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler: router,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
