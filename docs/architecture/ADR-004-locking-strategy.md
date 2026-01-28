# ADR-004: Pessimistic locking strategy for wallet operations

## Status
Accepted

## Context
Concurrent updates to the same wallet (e.g., multiple incoming payments) can cause race conditions where one update overwrites another, leading to incorrect balances.

## Decision
We implemented **Pessimistic Locking** using Redis.
1. **Mutex via Redis**: Before updating a wallet, we acquire a distributed lock in Redis keyed by the `wallet_id` or `user_id`.
2. **Context-Aware**: Locks are automatically released after a TTL or when the processing context is cancelled.
3. **Optimistic fallback**: While we use pessimistic locking at the application layer, the database update also includes a check (e.g., `WHERE locked_balance >= amount`) to ensure integrity.

## Consequences
- **Positive**: Guaranteed correctness even under high concurrency.
- **Negative**: Slightly increased latency due to Redis network round-trips.
