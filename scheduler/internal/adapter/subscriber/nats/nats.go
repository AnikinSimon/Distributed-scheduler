package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type CompletionSubscriber struct {
	logger *zap.Logger
	js     jetstream.JetStream
}

type JobCompletion struct {
	JobID        string  `json:"jobId"`
	Status       string  `json:"status"` // "completed" or "failed"
	FinishedAt   int64   `json:"finishedAt"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

type CompletionHandler = func(ctx context.Context, completion JobCompletion) error

func NewCompletionSubscriber(ctx context.Context, url string, logger *zap.Logger) (*CompletionSubscriber, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server: %w", err)
	}

	newJS, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstrea : %w", err)
	}

	logger.Info("Connected to NATS Jetstream for completion subscriber", zap.String("url", url))

	return &CompletionSubscriber{
		logger: logger,
		js:     newJS,
	}, nil
}

func (c *CompletionSubscriber) Subscribe(ctx context.Context, handler CompletionHandler) error {
	cons, err := c.js.CreateOrUpdateConsumer(ctx, "JOBS", jetstream.ConsumerConfig{
		FilterSubject: "JOBS.completed",
		Durable:       "scheduler-completion",
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	c.logger.Info("Start listening to Nats Job completion", zap.String("subject", "JOBS.completed"))

	go c.processMessages(ctx, cons, handler)

	return nil
}

func (c *CompletionSubscriber) processMessages(
	ctx context.Context,
	cons jetstream.Consumer,
	handler CompletionHandler,
) {
	iter, err := cons.Messages()
	if err != nil {
		c.logger.Error("failed to create message iterator", zap.Error(err))
		return
	}

	defer iter.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := iter.Next()
			if err != nil {
				c.logger.Error("failed to receive message from NATS", zap.Error(err))
				return
			}

			var completion JobCompletion
			if err := json.Unmarshal(msg.Data(), &completion); err != nil {
				c.logger.Error("failed to unmarshal message from NATS", zap.Error(err))
				_ = msg.Nak()
				continue
			}

			c.logger.Info("Received JobCompletion from NATS", zap.Any("job_completion", completion))

			if err := handler(ctx, completion); err != nil {
				c.logger.Error("failed to process message from NATS", zap.Error(err))
				_ = msg.Nak()
				continue
			}

			_ = msg.Ack()
		}
	}
}
