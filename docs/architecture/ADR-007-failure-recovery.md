# ADR-007: Exponential backoff and DLQ strategy

## Status
Accepted

## Context
Transient failures (like temporary database locks or Kafka partition rebalances) should not cause a permanent loss of data or error to the user.

## Decision
We implemented a **Tiered Failure Recovery** strategy.
1. **Exponential Backoff**: For retriable errors, we wait progressively longer before each retry (1s, 2s, 4s, etc.).
2. **Dead Letter Queue (DLQ)**: If a message fails after the maximum number of retries, it is moved to a special Kafka topic (`aegis.dlq`) for manual inspection.

## Consequences
- **Positive**: Extremely high reliability; transient errors are self-healing.
- **Negative**: "Poison pill" messages must be monitored manually in the DLQ to prevent them from staying hidden.
