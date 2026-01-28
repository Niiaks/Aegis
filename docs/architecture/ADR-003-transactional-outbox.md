# ADR-003: Transactional outbox vs dual-write to Kafka

## Status
Accepted

## Context
When a transaction is saved to the database, we often need to trigger downstream actions (like sending notifications or updating remote balances). Writing to both DB and Kafka simultaneously ("Dual Write") is dangerous because one might fail while the other succeeds.

## Decision
We implemented the **Transactional Outbox Pattern**.
1. **Atomicity**: The business change and the "event to be sent" are saved in the same DB transaction.
2. **Outbox Relay**: A dedicated worker (Relay) polls the `transaction_outbox` table and publishes messages to Kafka.
3. **At-least-once delivery**: Events are only marked as "processed" after a successful Kafka ACK.

## Consequences
- **Positive**: Guarantees that every database change eventually leads to a Kafka message.
- **Negative**: Introduces a small amount of latency (the relay polling interval).
