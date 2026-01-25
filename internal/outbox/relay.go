package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/Niiaks/Aegis/internal/kafka"
	"github.com/Niiaks/Aegis/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Relay struct {
	db          *pgxpool.Pool
	kafkaClient *kafka.Producer
	logger      *zerolog.Logger
	batchSize   int
	interval    time.Duration
}

func NewRelay(db *pgxpool.Pool, kafkaClient *kafka.Producer, logger *zerolog.Logger) *Relay {
	return &Relay{
		db:          db,
		kafkaClient: kafkaClient,
		logger:      logger,
		batchSize:   100,              // Process 100 events at a time
		interval:    10 * time.Second, // Poll every 100ms
	}
}

func (r *Relay) Start(ctx context.Context) error {
	r.logger.Info().Msg("Starting Outbox Relay")
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info().Msg("Stopping Outbox Relay")
			return nil
		case <-ticker.C:
			if err := r.processBatch(ctx); err != nil {
				r.logger.Error().Err(err).Msg("Failed to process batch")
			}
		}
	}
}

func (r *Relay) processBatch(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, event_type, payload, partition_key
		FROM transaction_outbox
		WHERE status = 'pending'
		ORDER BY id ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, r.batchSize)
	if err != nil {
		return err
	}

	var events []model.TransactionOutbox
	for rows.Next() {
		var e model.TransactionOutbox
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.PartitionKey); err != nil {
			rows.Close()
			return err
		}
		events = append(events, e)
	}
	rows.Close()

	if len(events) == 0 {
		return nil
	}

	if len(events) > 0 {
		r.logger.Info().Int("count", len(events)).Msg("Fetched outbox events")
	}

	var processedIDs []int64
	for _, e := range events {
		topic := r.getTopicForEvent(e.EventType)
		fmt.Println("the topic is", topic)
		err := r.kafkaClient.Publish(ctx, topic, []byte(e.PartitionKey), e.Payload)

		if err != nil {
			r.logger.Error().Err(err).Int64("event_id", e.ID).Str("event_type", e.EventType).Msg("Failed to publish event to Kafka")
			continue // Do not mark as processed
		}
		processedIDs = append(processedIDs, e.ID)
	}

	if len(processedIDs) == 0 {
		return nil // Nothing to update
	}

	_, err = tx.Exec(ctx, `
		       UPDATE transaction_outbox
		       SET status = 'processed', updated_at = NOW()
		       WHERE id = ANY($1)
	       `, processedIDs)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Relay) getTopicForEvent(eventType string) string {
	switch eventType {
	case kafka.EventPaymentIntentCreated:
		return kafka.TopicPaymentCreated
	case kafka.EventWebhookReceived:
		return kafka.TopicWebhookPending
	default:
		return kafka.TopicDLQ // Send unknown events to DLQ
	}
}
