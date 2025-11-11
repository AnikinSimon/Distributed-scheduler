package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/worker/internal/entity"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"time"
)

type JobSubscriber struct {
	js     jetstream.JetStream
	logger *zap.Logger
}

type JobHandler = func(ctx context.Context, job *entity.Job) error

type JobDTO struct {
	ID             string  `json:"id"`
	Kind           int     `json:"kind"`
	Status         string  `json:"status"`
	Interval       *string `json:"interval,omitempty"`
	Once           *int64  `json:"once,omitempty"`
	LastFinishedAt int64   `json:"lastFinishedAt"`
	Payload        any     `json:"payload"`
}

func NewJobSubscriber(ctx context.Context, url string, logger *zap.Logger) (*JobSubscriber, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connection error: %w", err)
	}

	newJS, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("nats jetstream error: %w", err)
	}

	logger.Info("Connected to NATS JetStream", zap.String("url", url))

	return &JobSubscriber{
		logger: logger,
		js:     newJS,
	}, nil
}

func (j *JobSubscriber) Subscribe(ctx context.Context, handler JobHandler) error {
	intervalCons, err := j.js.CreateOrUpdateConsumer(ctx, "JOBS", jetstream.ConsumerConfig{
		FilterSubject: "JOBS.interval.queued",
		Durable:       "worker-interval",
	})
	if err != nil {
		return fmt.Errorf("failed to create interval consumer: %w", err)
	}

	onceCons, err := j.js.CreateOrUpdateConsumer(ctx, "JOBS", jetstream.ConsumerConfig{
		FilterSubject: "JOBS.once.queued",
		Durable:       "worker-once",
	})
	if err != nil {
		return fmt.Errorf("failed to create once consumer: %w", err)
	}

	streams, err := intervalCons.Info(ctx)
	if err != nil {
		j.logger.Error("Failed to get stream info", zap.Error(err))
		return fmt.Errorf("failed to get streams: %w", err)
	}

	j.logger.Info("Consumer info", zap.String("name", streams.Name))

	go j.processMessages(ctx, intervalCons, handler)
	go j.processMessages(ctx, onceCons, handler)

	return nil
}

func (j *JobSubscriber) processMessages(ctx context.Context, cons jetstream.Consumer, handler JobHandler) {
	iter, err := cons.Messages()
	if err != nil {
		j.logger.Error("failed to subscribe to messages", zap.Error(err))
		return
	}

	j.logger.Info("Subscriber started")

	defer iter.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			j.logger.Info("Listening to messages")
			msg, err := iter.Next()
			if err != nil {
				j.logger.Error("failed to receive message", zap.Error(err))
				continue
			}

			var job JobDTO
			if err = json.Unmarshal(msg.Data(), &job); err != nil {
				j.logger.Error("failed to unmarshal message", zap.Error(err))
				msg.Nak()
				continue
			}

			j.logger.Info("Received message", zap.Any("job", job))

			ent, err := JobEntityFromDTO(&job)
			if err != nil {
				j.logger.Error("failed to convert message to entity", zap.Error(err))
				msg.Nak()
				continue
			}

			if err := handler(ctx, ent); err != nil {
				j.logger.Error("failed to process message", zap.Error(err))
				msg.Nak()
				continue
			}

			msg.Ack()
		}
	}
}

func JobEntityFromDTO(job *JobDTO) (*entity.Job, error) {
	ent := &entity.Job{}
	id, err := uuid.Parse(job.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse UUID", zap.String("id", job.ID))
	}

	ent.Id = id
	ent.Kind = entity.JobKind(job.Kind)
	if ent.Kind == entity.JobKindOnce {
		ent.Once = job.Once
	} else if ent.Kind == entity.JobKindInterval {
		dur, err := time.ParseDuration(*job.Interval)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse duration", zap.String("interval", *job.Interval))
		}
		ent.Interval = &dur
	}
	ent.LastFinishedAt = job.LastFinishedAt
	ent.Payload = job.Payload
	ent.Status = entity.JobStatus(job.Status)

	return ent, nil
}
