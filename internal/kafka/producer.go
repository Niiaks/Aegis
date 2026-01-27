package kafka

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
	cfg    *Config
	logger *zerolog.Logger
}

func NewProducer(cfg *Config, logger *zerolog.Logger) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.RequiredAcks(kgo.Acks(cfg.RequiredAcks)),
		kgo.ProduceRequestTimeout(cfg.ProducerTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	return &Producer{
		client: client,
		cfg:    cfg,
		logger: logger,
	}, nil
}

// Publish sends a message to the specified Kafka topic.
// key is used for partition assignment.
func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte) error {
	return p.PublishWithHeaders(ctx, topic, key, value, nil)
}

// PublishWithHeaders sends a message with optional headers.
func (p *Producer) PublishWithHeaders(ctx context.Context, topic string, key, value []byte, headers map[string]string) error {
	record := &kgo.Record{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: mapToHeaders(headers),
	}
	// ProduceSync sends the record and waits for acknowledgment
	results := p.client.ProduceSync(ctx, record).FirstErr()
	return results
}

func mapToHeaders(m map[string]string) []kgo.RecordHeader {
	if len(m) == 0 {
		return nil
	}
	headers := make([]kgo.RecordHeader, 0, len(m))
	for k, v := range m {
		headers = append(headers, kgo.RecordHeader{
			Key:   k,
			Value: []byte(v),
		})
	}
	return headers
}

// PulishAsync sends a message to the specified Kafka topic asynchronously.
func (p *Producer) PublishAsync(ctx context.Context, topic string, key, value []byte) {
	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: value,
	}
	p.client.Produce(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			return
		}
	})
}

func (p *Producer) Close() {
	p.logger.Info().Msg("closing Kafka producer")
	p.client.Close()
}
