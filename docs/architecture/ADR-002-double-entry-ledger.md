# ADR-002: Double-entry ledger design and balance materialization

## Status
Accepted

## Context
We need to track every movement of money with absolute auditability. A simple "balance" column on a user profile is insufficient because it doesn't leave a trail of "where the money came from."

## Decision
We implemented a **Double-Entry Ledger** system.
1. **Immutability**: Every transaction creates balanced `LedgerEntry` records (debits and credits). Entries are never updated; only new entries are appended.
2. **Zero-Sum**: Every transaction must sum to zero.
3. **Materialized View**: While the ledger is the source of truth, we maintain a `balance` and `locked_balance` in the `wallets` table for performance (avoiding summing millions of rows for every balance check).

## Consequences
- **Positive**: Complete audit trail. Any discrepancy can be identified by re-summing the ledger.
- **Negative**: Slightly higher storage requirements and slightly more complex database transactions.
