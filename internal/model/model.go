package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Model struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID         uuid.UUID `json:"id"`
	PlatformID string    `json:"platform_id" validate:"required"`
	PspID      string    `json:"psp_id" validate:"required"`
	Name       string    `json:"name" validate:"required,min=2,max=100"`
	Email      string    `json:"email" validate:"required,email"`
	Model
}

type Wallet struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id" validate:"required"`
	Balance       int64     `json:"balance"`
	LockedBalance int64     `json:"locked_balance"`
	Currency      string    `json:"currency" validate:"required,len=3"`
	Type          string    `json:"type" validate:"required,oneof=holding settlement revenue"`
	Model
}

type LedgerEntry struct {
	ID            int64     `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id" validate:"required"`
	AccountID     uuid.UUID `json:"account_id" validate:"required"`
	Debit         int64     `json:"debit" validate:"gte=0"`
	Credit        int64     `json:"credit" validate:"gte=0"`
	BalanceAfter  int64     `json:"balance_after" validate:"gte=0"`
	Description   string    `json:"description" validate:"required,oneof=revenue,payout,fee,refund"`
	Model
}

type Transaction struct {
	ID             uuid.UUID `json:"id"`
	IdempotencyKey string    `json:"idempotency_key" validate:"required"`
	UserID         uuid.UUID `json:"user_id" validate:"required"`
	Amount         int64     `json:"amount" validate:"required,gte=0"`
	Currency       string    `json:"currency" validate:"required,len=3"`
	PspReference   string    `json:"psp_reference"`
	Status         string    `json:"status" validate:"required,oneof=pending completed failed refunded"`
	Type           string    `json:"type" validate:"required,oneof=payment_intent payout refund fee"`
	FailureReason  string    `json:"failure_reason,omitempty"`
	Model
}

type TransactionOutbox struct {
	ID            int64           `json:"id" validate:"required"`
	EventType     string          `json:"event_type" validate:"required"`
	Payload       json.RawMessage `json:"payload" validate:"required"`
	PartitionKey  string          `json:"partition_key" validate:"required"`
	Status        string          `json:"status" validate:"required,oneof=pending processed failed dlq"`
	CorrelationID uuid.UUID       `json:"correlation_id" validate:"required"`
	RetryCount    int             `json:"retry_count" validate:"gte=0"`
	LastError     string          `json:"last_error,omitempty"`
	NextRetryAt   time.Time       `json:"next_retry_at"`
	MaxRetries    int             `json:"max_retries" validate:"gte=0 default=5"`
	Model
}

type PspWebhook struct {
	ID      uuid.UUID       `json:"id"`
	EventID string          `json:"event_id" validate:"required"`
	Payload json.RawMessage `json:"payload" validate:"required"`
	Status  string          `json:"status" validate:"required,oneof=received error processed"`
	Model
}

type ReconciliationRun struct {
	ID      uuid.UUID `json:"id"`
	RunDate time.Time `json:"run_date" validate:"required"`
	Status  string    `json:"status" validate:"required,oneof=discrepancy matched"`
	Model
}

type Discrepancy struct {
	ID                  uuid.UUID `json:"id"`
	ReconciliationRunID uuid.UUID `json:"reconciliation_run_id" validate:"required"`
	ExpectedAmount      int64     `json:"expected_amount" validate:"required"`
	ActualAmount        int64     `json:"actual_amount" validate:"required"`
	Reason              string    `json:"reason" validate:"required"`
	Status              string    `json:"status" validate:"required,oneof=unresolved resolved"`
	Model
}

type IdempotencyKey struct {
	ID           uuid.UUID       `json:"id"`
	Key          string          `json:"key" validate:"required"`
	RequestHash  string          `json:"request_hash" validate:"required,len=64"`
	RequestPath  string          `json:"request_path" validate:"required"`
	ResponseCode int             `json:"response_code,omitempty"`
	ResponseBody json.RawMessage `json:"response_body,omitempty"`
	UserID       *uuid.UUID      `json:"user_id,omitempty"`
	ExpiresAt    time.Time       `json:"expires_at" validate:"required"`
	CreatedAt    time.Time       `json:"created_at"`
}
