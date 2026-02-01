package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/entity"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/port/publisher"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

var _ publisher.Publisher = (*JobPublisher)(nil)

type JobPublisher struct {
	logger *zap.Logger
	js     jetstream.JetStream
}

func NewJobPublisher(ctx context.Context, url string, logger *zap.Logger) (*JobPublisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS: %w", err)
	}

	newJS, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("creating JetStream: %w", err)
	}

	streams := newJS.ListStreams(ctx)
	for stream := range streams.Info() {
		logger.Info("Stream", zap.Any("name", stream))
	}

	logger.Info("Connected to NATS", zap.String("url", url))

	return &JobPublisher{
		logger: logger,
		js:     newJS,
	}, nil
}

func (n *JobPublisher) Publish(ctx context.Context, job *entity.Job) error {
	dto, err := publisher.JobDTOFromEntity(job)
	if err != nil {
		return fmt.Errorf("failed to convert to DTO: %w", err)
	}
	dto.Status = "queued"
	subject := n.subjectForJob(dto)

	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	n.logger.Info("Publishing job", zap.String("subject", subject), zap.ByteString("data", data))

	_, err = n.js.Publish(ctx, subject, data)
	if err != nil {
		n.logger.Error(
			"Failed to publish payload",
			zap.Error(err),
			zap.String("subject", subject),
			zap.String("id", dto.ID),
		)
		return fmt.Errorf("failed to publish Job to NATS: %w", err)
	}

	return nil
}

func (n *JobPublisher) subjectForJob(job *publisher.JobDTO) string {
	switch job.Kind {
	case entity.JobKindInterval:
		return fmt.Sprintf("JOBS.interval.%s", job.Status)
	case entity.JobKindOnce:
		return fmt.Sprintf("JOBS.once.%s", job.Status)
	}

	return ""
}
