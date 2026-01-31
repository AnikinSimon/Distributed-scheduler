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
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/google/uuid"
	goredislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	mutexName = "scheduler-mutex"
)

func Start(cfg config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())

	logger, err := zap.NewProduction()

	defer cancel()
	defer logger.Sync()

	if err != nil {
		return err
	}

	pub, err := nats.NewJobPublisher(ctx, cfg.NATSURL, logger)
	if err != nil {
		logger.Error("Failed to create job publisher", zap.Error(err))
		return err
	}

	client := goredislib.NewClient(&goredislib.Options{

		Addr: config.RedisConnString(cfg.Redis),
		//Username: cfg.Redis.User,
		Password: cfg.Redis.Password,
	})

	pool := goredis.NewPool(client)

	rs := redsync.New(pool)

	redisMutex := rs.NewMutex(mutexName,
		redsync.WithExpiry(cfg.SchedulerInterval+1*time.Second),
		redsync.WithTries(1))

	jobsRepo := postgres.NewJobsRepo(ctx, cfg.Storage, logger)

	scheduler := cases.NewSchedulerCase(jobsRepo, logger, cfg.SchedulerInterval, pub, redisMutex)

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
