# ADR-006: Circuit breaker pattern for external service calls

## Status
Accepted

## Context
Our system depends on external services like Paystack and Kafka. If these services become slow or unavailable, a backlog of requests can overwhelm Aegis (cascading failure).

## Decision
We implemented **Circuit Breakers** via `gobreaker`.
1. **Fail-Fast**: If an external service fails multiple times in a short window, the circuit "opens" and subsequent calls fail immediately.
2. **Auto-Recovery**: After a timeout, the circuit enters a "half-open" state, allowing a single probe request to check if the remote service has recovered.

## Consequences
- **Positive**: Protects the system from wasting resources on calls that are likely to fail.
- **Negative**: Requires careful tuning of failure thresholds to avoid accidental triggering.
