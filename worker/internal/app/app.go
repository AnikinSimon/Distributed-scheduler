package app

import (
	"context"
	"fmt"
	"time"

	"github.com/AnikinSimon/Distributed-scheduler/worker/config"
	nats2 "github.com/AnikinSimon/Distributed-scheduler/worker/internal/adapter/publisher/nats"
	"github.com/AnikinSimon/Distributed-scheduler/worker/internal/adapter/subscriber/nats"
	"github.com/AnikinSimon/Distributed-scheduler/worker/internal/entity"
	"go.uber.org/zap"
)

func Start(cfg *config.Config) error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	sub, err := nats.NewJobSubscriber(context.Background(), cfg.NATSURL, logger)
	if err != nil {
		return fmt.Errorf("failed to create nats job subscriber: %w", err)
	}

	pub, err := nats2.NewCompletionPublisher(context.Background(), cfg.NATSURL, logger)
	if err != nil {
		return fmt.Errorf("failed to create nats completion publisher: %w", err)
	}

	ctx := context.Background()

	if err := sub.Subscribe(ctx, func(ctx context.Context, job *entity.Job) error {
		logger.Info(
			"Received Job",
			zap.String("job_id", job.ID.String()),
			zap.Int("kind", int(job.Kind)),
			zap.String("status", string(job.Status)),
			zap.Any("payload", job.Payload),
		)

		completion := nats2.JobCompletion{
			JobID:      job.ID.String(),
			Status:     entity.JobStatusCompleted,
			FinishedAt: time.Now().UnixMilli(),
		}

		time.Sleep(1 * time.Second)

		if errPub := pub.Publish(ctx, completion); err != nil {
			logger.Error("failed to publish completion", zap.Error(errPub))
			return errPub
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to subscribe to job: %w", err)
	}

	logger.Info("Worker started")
	select {}
}
