package app

import (
	"context"
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/config"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/adapter/repo/postgres"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/cases"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/input/http/gen"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/input/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start(cfg config.Config) error {
	// TODO: Create jobs repo

	ctx, cancel := context.WithCancel(context.Background())

	jobsRepo := postgres.NewJobsRepo(ctx, cfg.Storage) // TODO: pg config

	scheduler := cases.NewSchedulerCase(jobsRepo)

	server := handler.NewServer(scheduler)

	strictHandler := gen.NewStrictHandler(server, nil)

	router := chi.NewRouter()

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

	go httpServer.ListenAndServe()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	cancel()

	log.Println("Stopping application", slog.String("signal", sign.String()))

	httpServer.Shutdown(ctx)

	return nil
}
