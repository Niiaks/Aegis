┌─────────────────────────────────────────────────────────────────────────────┐
│ EXPONENTIAL BACKOFF │
│ (Retry with increasing delays) │
├─────────────────────────────────────────────────────────────────────────────┤
│ │
│ WHERE: Outbox Relay, Payout Worker, any PSP call │
│ │
│ FLOW: │
│ │
│ Attempt 1 ──FAIL──▶ Wait 1s │
│ Attempt 2 ──FAIL──▶ Wait 2s │
│ Attempt 3 ──FAIL──▶ Wait 4s │
│ Attempt 4 ──FAIL──▶ Wait 8s │
│ Attempt 5 ──FAIL──▶ Wait 16s (cap at 30s) │
│ Attempt 6 ──FAIL──▶ Move to DLQ │
│ │
│ STORED IN DB: │
│ transaction_outbox.retry_count = 5 │
│ transaction_outbox.last_error = "stripe: rate limited" │
│ transaction_outbox.next_retry_at = NOW() + interval │
│ │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│ DEAD LETTER QUEUE (DLQ) │
│ (Poison messages that can't be processed) │
├─────────────────────────────────────────────────────────────────────────────┤
│ │
│ KAFKA TOPICS: │
│ │
│ aegis.transactions ◀── Normal events │
│ aegis.transactions.dlq ◀── Failed after max retries │
│ aegis.settlements ◀── Normal events │
│ aegis.settlements.dlq ◀── Failed after max retries │
│ │
│ FLOW: │
│ │
│ [Kafka Consumer] │
│ │ │
│ ├── Process message │
│ │ │ │
│ │ ├── SUCCESS ──▶ Commit offset │
│ │ │ │
│ │ └── FAIL (after 6 retries) │
│ │ │ │
│ │ ▼ │
│ │ ┌───────────────────┐ │
│ │ │ Publish to DLQ │ │
│ │ │ + original error │ │
│ │ │ + retry count │ │
│ │ │ + timestamp │ │
│ │ └───────────────────┘ │
│ │ │ │
│ │ ▼ │
│ │ Alert ops team (New Relic / PagerDuty) │
│ │ │ │
│ └────── Commit offset (don't block queue) │
│ │
│ DLQ MESSAGE FORMAT: │
│ { │
│ "original_topic": "aegis.transactions", │
│ "original_message": {...}, │
│ "error": "psp timeout after 30s", │
│ "retry_count": 6, │
│ "failed_at": "2026-01-16T10:00:00Z", │
│ "correlation_id": "tx_xyz" │
│ } │
│ │
│ RECOVERY: │
│ - Manual inspection via admin tool │
│ - Fix the issue │
│ - Replay from DLQ back to main topic │
│ │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│ CIRCUIT BREAKER │
│ (Stop calling a failing service temporarily) │
├─────────────────────────────────────────────────────────────────────────────┤
│ │
│ WHERE: PSP calls (Stripe), any external dependency │
│ │
│ STATES: │
│ │
│ ┌────────────┐ 5 failures ┌────────────┐ │
│ │ CLOSED │ ────────────────▶ │ OPEN │ │
│ │ (normal) │ │ (failing) │ │
│ └────────────┘ └────────────┘ │
│ ▲ │ │
│ │ │ after 30s │
│ │ ▼ │
│ │ success ┌─────────────┐ │
│ └─────────────────────────│ HALF-OPEN │ │
│ │ (testing) │ │
│ failure └─────────────┘ │
│ ┌───────────────────────── │ │
│ ▼ │ │
│ ┌────────────┐ │ │
│ │ OPEN │ ◀──────────────────────┘ │
│ └────────────┘ │
│ │
│ CONFIG: │
│ - Failure threshold: 5 consecutive failures │
│ - Open duration: 30 seconds │
│ - Half-open max requests: 1 (probe) │
│ │
│ BEHAVIOR WHEN OPEN: │
│ - Return cached response if available │
│ - Return error immediately (fail fast) │
│ - Queue for retry later (preferred for payments) │
│ │
│ USE LIBRARY: sony/gobreaker (battle-tested) │
│ │
└─────────────────────────────────────────────────────────────────────────────┘

═══════════════════════════════════════════════════════════════════════════════
HOW THEY WORK TOGETHER: PSP CALL EXAMPLE
═══════════════════════════════════════════════════════════════════════════════

[Payout Worker] wants to call Stripe
│
▼
┌──────────────────────┐
│ CHECK CIRCUIT STATE │
└──────────────────────┘
│
├── OPEN? ──▶ Don't call Stripe
│ Queue for retry (back to outbox)
│ Return immediately
│
└── CLOSED/HALF-OPEN? ──▶ Proceed
│
▼
┌───────────────┐
│ Call Stripe │
└───────────────┘
│
├── SUCCESS ──▶ Circuit stays CLOSED
│ Commit transaction
│
└── FAILURE ──▶ Increment failure count
│
├── Count < 5 ──▶ Exponential backoff
│ retry_count++
│ next_retry_at = NOW() + 2^retry_count
│ Update outbox row
│
└── Count >= 5 ──▶ Open circuit
Move to DLQ
Alert ops

═══════════════════════════════════════════════════════════════════════════════
UPDATED OUTBOX MODEL
═══════════════════════════════════════════════════════════════════════════════

type TransactionOutbox struct {
ID int64  
 EventType string  
 Payload json.RawMessage
PartitionKey string  
 Status string // pending | processing | processed | failed | dlq
CorrelationID uuid.UUID  
 RetryCount int  
 MaxRetries int // default: 6
LastError string  
 NextRetryAt time.Time // when to retry (for backoff)
Model
}

═══════════════════════════════════════════════════════════════════════════════
UPDATED OUTBOX RELAY LOGIC
═══════════════════════════════════════════════════════════════════════════════

[Outbox Relay Worker]
│
│ // Only pick up messages ready for retry
│
├── SELECT \* FROM transaction_outbox
│ WHERE status IN ('pending', 'failed')
│ AND (next_retry_at IS NULL OR next_retry_at <= NOW())
│ ORDER BY id
│ LIMIT 100
│ FOR UPDATE SKIP LOCKED
│
├── For each message:
│ │
│ ├── Check circuit breaker for Kafka
│ │
│ ├── TRY: Publish to Kafka
│ │ │
│ │ ├── SUCCESS:
│ │ │ UPDATE SET status = 'processed'
│ │ │
│ │ └── FAILURE:
│ │ IF retry_count >= max_retries:
│ │ UPDATE SET status = 'dlq'
│ │ Publish to aegis.outbox.dlq
│ │ Alert ops
│ │ ELSE:
│ │ retry_count++
│ │ next_retry_at = NOW() + (2^retry_count seconds)
│ │ last_error = error.message
│ │ UPDATE SET status = 'failed'
