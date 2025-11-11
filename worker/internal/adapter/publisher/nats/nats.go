package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

const (
	subject = "JOBS.completed"
)

type CompletionPublisher struct {
	js     jetstream.JetStream
	logger *zap.Logger
}

type JobCompletion struct {
	JobID        string  `json:"jobId"`
	Status       string  `json:"status"` // "completed" or "failed"
	FinishedAt   int64   `json:"finishedAt"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

func NewCompletionPublisher(ctx context.Context, url string, logger *zap.Logger) (*CompletionPublisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("could not connect to NATS: %w", err)
	}

	newJS, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("could not connect to JetStream: %w", err)
	}

	return &CompletionPublisher{
		js:     newJS,
		logger: logger,
	}, nil
}

func (c *CompletionPublisher) Publish(ctx context.Context, completion JobCompletion) error {
	data, err := json.Marshal(&completion)
	if err != nil {
		return fmt.Errorf("could not marshal job completion: %w", err)
	}

	c.logger.Debug("publishing completion", zap.String("jobId", completion.JobID))

	_, err = c.js.Publish(ctx, subject, data)
	if err != nil {
		c.logger.Error("could not publish job completion",
			zap.String("jobId", completion.JobID),
			zap.String("status", completion.Status),
			zap.String("subject", subject),
			zap.Error(err))
		return fmt.Errorf("could not publish job completion: %w", err)
	}

	c.logger.Info("job completion published to NATS",
		zap.String("jobId", completion.JobID),
		zap.String("status", completion.Status),
		zap.String("subject", subject))

	return nil
}
