package app

import (
	"context"
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/config"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/adapter/publisher/nats"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/adapter/repo/postgres"
	nats2 "github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/adapter/subscriber/nats"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/cases"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/input/http/gen"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/input/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start(cfg config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())

	logger, err := zap.NewProduction()

	defer cancel()
	defer logger.Sync()

	if err != nil {
		return err
	}

	// TODO: add  nats url

	pub, err := nats.NewJobPublisher(ctx, cfg.NATSURL, logger)
	if err != nil {
		logger.Error("Failed to create job publisher", zap.Error(err))
		return err
	}

	jobsRepo := postgres.NewJobsRepo(ctx, cfg.Storage, logger)

	scheduler := cases.NewSchedulerCase(jobsRepo, logger, cfg.SchedulerInterval, pub)

	sub, err := nats2.NewCompletionSubscriber(ctx, cfg.NATSURL, logger)
	if err != nil {
		logger.Error("Failed to create completion subscriber", zap.Error(err))
		return err
	}

	if err = sub.Subscribe(ctx, func(ctx context.Context, completion nats2.JobCompletion) error {
		id, errUUID := uuid.Parse(completion.JobID)
		if errUUID != nil {
			return errUUID
		}
		return scheduler.HandleJobCompletion(ctx, id, completion.Status, completion.FinishedAt)
	}); err != nil {
		logger.Error("Failed to subscribe to completion", zap.Error(err))
	}

	go func() {
		if err := scheduler.Start(context.TODO()); err != nil {
			logger.Error("Scheduler tick loop failed", zap.Error(err))
		}
	}()

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

	logger.Info("Stopping application", zap.String("signal", sign.String()))

	httpServer.Shutdown(ctx)

	return nil
}
