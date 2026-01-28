# ADR-005: Idempotency implementation with Redis

## Status
Accepted

## Context
Network failures are common. A client might send a payment request, the server processes it, but the client never receives the response. The client then retries the request. Without idempotency, this results in double-spending.

## Decision
We implemented a **Redis-based Idempotency Layer**.
1. **Idempotency Key**: Clients provide a unique `Idempotency-Key` header.
2. **State Storage**: Redis stores the status and the cached response of the initial request for 24 hours.
3. **Atomic Check-and-Set**: We use Redis to ensure only one process handles a specific key at a time.

## Consequences
- **Positive**: Safe retries for clients.
- **Negative**: Adds a dependency on Redis for the core path of mutating requests.
