# ADR-001: Why PostgreSQL over MongoDB for financial ledgers

## Status
Accepted

## Context
When building a financial system, data integrity and ACID compliance are non-negotiable. We needed to choose between a relational database (PostgreSQL) and a NoSQL document store (MongoDB) for storing transactions, wallets, and ledger entries.

## Decision
We chose **PostgreSQL** for the following reasons:
1. **ACID Transactions**: Financial operations (like moving money between wallets) require atomic updates across multiple tables. Postgres provides robust transaction support that ensures partial failures don't leave the system in an inconsistent state.
2. **Schema Enforcement**: Money movement requires strict structure. Postgres ensures that amounts are integers, foreign keys are valid, and constraints (like `balance >= 0`) are always respected.
3. **Maturity & Tooling**: Postgres has a long-standing reputation for reliability in banking and fintech.

## Consequences
- **Positive**: We gain strong consistency and reliable locking mechanisms.
- **Negative**: Horizontal scaling is slightly more complex than NoSQL, but vertical scaling is sufficient for most throughput needs.
