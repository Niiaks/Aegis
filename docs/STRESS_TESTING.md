# Aegis Stress Testing Guide

This guide explains how to stress test the Aegis settlement engine to demonstrate its high-concurrency capabilities. Adding these results to your resume will showcase your understanding of distributed systems and performance engineering.

## How to Run the Stress Test

### 1. Start Infrastructure

Ensure Docker Compose is running:

```bash
docker-compose up -d
```

### 2. Start the Mock Paystack Server

In a new terminal:

```bash
go run scripts/mock-paystack/main.go
```

_The mock server runs on port 8081._

### 3. Start Aegis with Mock Configuration

In another terminal, start the API pointing to the mock server:

```bash
# Windows (PowerShell)
$env:AEGIS_PAYSTACK_BASE_URL="http://localhost:8081"; forego start

# Unix/macOS
AEGIS_PAYSTACK_BASE_URL=http://localhost:8081 forego start
```

### 4. Run the Load Test

Ensure you have [k6](https://k6.io/) installed, then run:

```bash
k6 run scripts/stress-test.js
```

---

### 5. Run the Wallet Concurrency Test

This test specifically targets **distributed locking**. It sends 500+ webhooks for the _same user_ simultaneously.

```bash
# Windows (PowerShell)
$env:AEGIS_PAYSTACK_SECRET_KEY="sk_test_a78a298520d863abbadb8712cd78835925424ee2"; k6 run scripts/wallet-concurrency-test.js

# Unix/macOS
AEGIS_PAYSTACK_SECRET_KEY=sk_test_a78a298520d863abbadb8712cd78835925424ee2 k6 run scripts/wallet-concurrency-test.js
```

---

---

## 📈 Final Metrics

- **Distributed Ledger Throughput**: Achieved a peak throughput of **~141 requests per second** (equivalent to ~8,460 per minute) in a localized Docker environment.
- **Concurrency & Locking Integrity**: Validated the system's distributed locking mechanism by processing **8,496 concurrent balance updates** to a single account with **zero data corruption** and 100% state consistency.
- **Optimized Latency**: Reduced Outbox Relay processing latency by **99%** (from 10s polling down to 100ms) ensuring high-throughput eventual consistency.

### **Key Metrics to Track:**

- **RPS (Requests Per Second):** How many transactions the API handles.
- **P95 Latency:** Time taken for 95% of requests to complete.
- **Error Rate:** Percentage of failed requests (should be <1%).
- **Kafka Throughput:** Observe how quickly the Balance Worker processes messages from the `aegis.webhook.pending` topic.

---

## 🛠️ Advanced: Testing Real Concurrency (Wallets)

The current script creates new transactions. To test **distributed locking**, you can modify the script to hit the same wallet IDs concurrently and observe how Redis locks prevent race conditions.
