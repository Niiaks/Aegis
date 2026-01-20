package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Message is a simplified wrapper around Kafka records
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Partition int32
	Offset    int64
	Timestamp time.Time
	Headers   map[string]string
}

// Handler processes a single message. Return error to trigger retry.
type Handler func(ctx context.Context, msg *Message) error

type Consumer struct {
	client *kgo.Client
	cfg    *Config
	topic  string
	group  string
}

func NewConsumer(cfg *Config, group, topic string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.SessionTimeout(cfg.SessionTimeout),
		kgo.HeartbeatInterval(cfg.HeartbeatInterval),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), // Start from earliest if no offset
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	return &Consumer{
		client: client,
		cfg:    cfg,
		topic:  topic,
		group:  group,
	}, nil
}

// Run starts consuming messages and calls handler for each.
// Blocks until context is cancelled.
func (c *Consumer) Run(ctx context.Context, handler Handler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fetches := c.client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			// Log errors but continue - transient errors are common
			for _, err := range errs {
				fmt.Printf("fetch error: topic=%s partition=%d err=%v\n",
					err.Topic, err.Partition, err.Err)
			}
		}

		fetches.EachRecord(func(record *kgo.Record) {
			msg := &Message{
				Topic:     record.Topic,
				Key:       record.Key,
				Value:     record.Value,
				Partition: record.Partition,
				Offset:    record.Offset,
				Timestamp: record.Timestamp,
				Headers:   headersToMap(record.Headers),
			}

			if err := c.processWithRetry(ctx, handler, msg); err != nil {
				// After all retries failed - will be handled by worker (DLQ)
				fmt.Printf("message processing failed after retries: %v\n", err)
			}
		})

		// Commit offsets after processing batch
		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			fmt.Printf("failed to commit offsets: %v\n", err)
		}
	}
}

func (c *Consumer) processWithRetry(ctx context.Context, handler Handler, msg *Message) error {
	var lastErr error

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff:
			backoff := c.cfg.RetryBackoff * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		if err := handler(ctx, msg); err != nil {
			lastErr = err
			continue
		}
		return nil // Success
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Consumer) Close() {
	c.client.Close()
}

func headersToMap(headers []kgo.RecordHeader) map[string]string {
	m := make(map[string]string, len(headers))
	for _, h := range headers {
		m[h.Key] = string(h.Value)
	}
	return m
}
