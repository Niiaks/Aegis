package kafka

import (
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Topic name contains all kafka topics used in the application
const (
	TopicPaymentCreated = "aegis.payment.created"
	TopicLedgerEntries  = "aegis.ledger.entries"
	TopicBalanceUpdate  = "aegis.balance.update"

	TopicWebhookPending = "aegis.webhook.pending"

	TopicPayoutPending      = "aegis.payout.pending"
	TopicPayoutStatusUpdate = "aegis.payout.status.update"

	TopicReconciliationJob   = "aegis.reconciliation.job"
	TopicDiscrepancyDetected = "aegis.discrepancy.detected"

	TopicDLQ = "aegis.dlq"
)

// ConsumerGroup names for different Kafka consumers
const (
	GroupTransactionWorker = "aegis.transaction.worker"
	GroupSettlementWorker  = "aegis.settlement.worker"
	GroupBalanceWorker     = "aegis.balance.worker"
	GroupWebhookWorker     = "aegis.webhook.worker"
	GroupPayoutWorker      = "aegis.payout.worker"
	GroupReconciliation    = "aegis.reconciliation.worker"
)

type Config struct {
	Brokers           []string
	ProducerTimeout   time.Duration
	RequiredAcks      kgo.Acks
	SessionTimeout    time.Duration
	HeartbeatInterval time.Duration
	MaxPollRecords    int
	MaxRetries        int
	RetryBackoff      time.Duration
}

func DefaultConfig(brokers []string) *Config {
	return &Config{
		Brokers:           brokers,
		ProducerTimeout:   10 * time.Second,
		RequiredAcks:      kgo.Acks{},
		SessionTimeout:    10 * time.Second,
		HeartbeatInterval: 3 * time.Second,
		MaxPollRecords:    100,
		MaxRetries:        5,
		RetryBackoff:      1 * time.Second,
	}
}
