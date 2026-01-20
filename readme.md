# Aegis: Distributed Settlement & Ledger Infrastructure

Aegis is a production-grade financial settlement engine designed to handle mission-critical money movement at scale. Built with a focus on absolute data integrity, auditability, and high concurrency, Aegis implements the same architectural patterns used by industry leaders like Stripe and Paystack to manage ledgers, seller wallets, and payouts.

## The Core Challenge

In financial systems, "close enough" is a failure. Aegis solves the hardest problems in distributed systems:

| Problem                    | Solution                                                           |
| -------------------------- | ------------------------------------------------------------------ |
| **The Dual-Write Problem** | Transactional Outbox ensures database and Kafka are always in sync |
| **Race Conditions**        | Redis-backed distributed locking prevents double-spending          |
| **Auditability**           | Immutable double-entry ledger maintains zero-sum integrity         |
| **Failure Recovery**       | Circuit breakers, exponential backoff, and DLQ handle failures     |

## Tech Stack

| Component               | Technology          | Purpose                                             |
| ----------------------- | ------------------- | --------------------------------------------------- |
| **Language**            | Go 1.25.5+          | High-performance concurrency                        |
| **Database**            | PostgreSQL + pgx    | ACID-compliant transactions with native driver      |
| **Message Broker**      | Apache Kafka        | Event-driven orchestration via transactional outbox |
| **Distributed Locking** | Redis               | Idempotency keys and pessimistic locking            |
| **Observability**       | New Relic + Zerolog | APM, distributed tracing, structured logging        |

## Architecture & Design Patterns

### 1. Immutable Double-Entry Ledger

Aegis implements strict **Double-Entry Bookkeeping**. Money is never "updated"; it is moved via balanced debits and credits.

- **Zero-Sum Integrity**: Every transaction's debits and credits must sum to zero before commitment
- **Cents-Only Precision**: All currency handled as `int64` (minor units) to eliminate floating-point errors
- **Audit Trail**: Every balance change creates an immutable `LedgerEntry` record

### 2. Transactional Outbox Pattern

To ensure atomicity between the database and Kafka, Aegis uses the **Outbox Pattern**:

```
┌─────────────────────────────────────────────────────────┐
│                    Single Transaction                    │
│  ┌──────────────┐    ┌──────────────────────────────┐   │
│  │ transactions │ +  │ transaction_outbox (pending) │   │
│  └──────────────┘    └──────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
              ┌─────────────────────────┐
              │   Outbox Relay Worker   │
              │   (polls & publishes)   │
              └─────────────────────────┘
                            │
                            ▼
                    ┌───────────┐
                    │   Kafka   │
                    └───────────┘
```

### 3. Redis-Backed Idempotency

Every mutating request is keyed by an `Idempotency-Key` header. Redis stores:

- **Request fingerprint**: Hash of the payload to detect duplicate submissions
- **Response cache**: Returns cached response for retried requests
- **TTL**: 24-hour expiration for idempotency records

```
Key:    idempotency:{merchant_id}:{idempotency_key}
Value:  {request_hash, response, status, created_at}
TTL:    86400 seconds
```

### 4. Distributed Locking for Wallet Operations

To prevent race conditions during concurrent settlements, Aegis uses **Redis distributed locks**:

```
Key:    lock:wallet:{wallet_id}
Value:  {owner_id, acquired_at}
TTL:    30 seconds (auto-release on crash)
```

Locks are acquired in **sorted order by wallet ID** to mathematically eliminate deadlocks when a transaction touches multiple wallets.

### 5. Reliability Patterns

#### Circuit Breaker

Protects against cascading failures when calling external services (PSPs, Kafka).

**States:**

- **Closed** (normal): All requests pass through
- **Open** (failing): Requests fail immediately without calling service
- **Half-Open** (testing): Single probe request to check if service recovered

**Configuration:**

- Failure threshold: 5 consecutive failures
- Open duration: 30 seconds
- Library: `github.com/sony/gobreaker`

#### Exponential Backoff

Retries failed operations with increasing delays to handle transient failures.

**Retry Schedule:**

```
Attempt 1 → Wait 1s
Attempt 2 → Wait 2s
Attempt 3 → Wait 4s
Attempt 4 → Wait 8s
Attempt 5 → Wait 16s
Attempt 6 → Move to DLQ
```

**Applied to:** Outbox relay, PSP calls, Kafka producers

#### Dead Letter Queue (DLQ)

Isolates poison messages that fail after maximum retries for manual inspection.

**Kafka Topics:**

- `aegis.transactions` → `aegis.transactions.dlq`
- `aegis.settlements` → `aegis.settlements.dlq`

**DLQ Message Structure:**

```json
{
  "original_topic": "aegis.transactions",
  "original_message": {...},
  "error": "psp timeout after 30s",
  "retry_count": 6,
  "failed_at": "2026-01-16T10:00:00Z",
  "correlation_id": "tx_xyz"
}
```

**Recovery:** Messages can be replayed from DLQ after fixing root cause.

## Observability

Aegis is built with the belief that a system is only as good as its visibility.

| Capability              | Implementation                                               |
| ----------------------- | ------------------------------------------------------------ |
| **Distributed Tracing** | New Relic APM with correlation IDs across service boundaries |
| **Structured Logging**  | JSON-based logging via Zerolog with request context          |
| **Database Tracing**    | pgx query tracer integrated with New Relic                   |
| **Health Checks**       | Liveness and readiness probes for Kubernetes                 |


## Architecture Decision Records

Aegis includes documentation on engineering trade-offs:

| ADR     | Decision                                               |
| ------- | ------------------------------------------------------ |
| ADR-001 | Why PostgreSQL over MongoDB for financial ledgers      |
| ADR-002 | Double-entry ledger design and balance materialization |
| ADR-003 | Transactional outbox vs dual-write to Kafka            |
| ADR-004 | Pessimistic locking strategy for wallet operations     |
| ADR-005 | Idempotency implementation with Redis                  |
| ADR-006 | Circuit breaker pattern for external service calls     |
| ADR-007 | Exponential backoff and DLQ strategy                   |

## Dependencies

| Library                                  | Purpose                        |
| ---------------------------------------- | ------------------------------ |
| `github.com/go-chi/chi/v5`               | HTTP router and middleware     |
| `github.com/jackc/pgx/v5`                | PostgreSQL native driver       |
| `github.com/twmb/franz-go`               | Kafka client                   |
| `github.com/redis/go-redis/v9`           | Redis client                   |
| `github.com/sony/gobreaker`              | Circuit breaker implementation |
| `github.com/cenkalti/backoff/v4`         | Exponential backoff            |
| `github.com/go-playground/validator/v10` | Struct validation              |
| `github.com/rs/zerolog`                  | Structured logging             |
| `github.com/newrelic/go-agent/v3`        | APM and distributed tracing    |

## Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make

### Run Infrastructure

```bash
docker-compose up -d
```

### Run Migrations

```bash
make migrate-up
```

### Run Aegis

```bash
make run
```

### Run Tests

```bash
make test
```

## Configuration

Aegis is configured via environment variables:

| Variable        | Description                                  | Default        |
| --------------- | -------------------------------------------- | -------------- |
| `PRIMARY_ENV`   | Environment (development/staging/production) | development    |
| `DATABASE_HOST` | PostgreSQL host                              | localhost      |
| `DATABASE_PORT` | PostgreSQL port                              | 5432           |
| `REDIS_ADDRESS` | Redis connection string                      | localhost:6379 |
| `SERVER_PORT`   | HTTP server port                             | 8080           |

## 🤝 Contact

Project Link: [https://github.com/Niiaks/Aegis](https://github.com/Niiaks/Aegis)
