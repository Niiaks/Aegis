ACTORS:
[Customer] - Buyer on merchant's platform
[Merchant App] - Shopify/Amazon store calling your API
[Aegis API] - Your settlement system
[Redis] - Cache & locks
[Postgres] - Source of truth
[Kafka] - Event bus
[PSP] - Stripe/Paystack
[Workers] - Background processors

═══════════════════════════════════════════════════════════════════════════════
STEP 1: CUSTOMER CLICKS "PAY" → MERCHANT CREATES PAYMENT INTENT
═══════════════════════════════════════════════════════════════════════════════

[Merchant App] ──POST /api/v1/payment-intents──▶ [Aegis API]
{
"amount": 10000, // $100.00 in cents
"currency": "USD",
"merchant_id": "m_123",
"idempotency_key": "order_abc_123"
}

═══════════════════════════════════════════════════════════════════════════════
STEP 2: IDEMPOTENCY CHECK (Redis → Postgres fallback)
═══════════════════════════════════════════════════════════════════════════════

[Aegis API]
│
├──GET idempotency:m_123:order_abc_123──▶ [Redis]
│ │
│◀─────────── NULL (not found) ───────────────┘
│
├──SELECT \* FROM transactions
│ WHERE idempotency_key = 'order_abc_123'
│ AND user_id = 'm_123'──────────────────▶ [Postgres]
│ │
│◀─────────── NULL (not found) ───────────────┘
│
└── PROCEED (first time seeing this request)

═══════════════════════════════════════════════════════════════════════════════
STEP 3: CALL PSP TO CREATE PAYMENT INTENT
═══════════════════════════════════════════════════════════════════════════════

[Aegis API] ──POST /v1/payment_intents──▶ [Stripe]
{
"amount": 10000,
"currency": "usd",
"metadata": {"aegis_tx_id": "tx_xyz"}
}
│
◀─────────────────────────────────┘
{
"id": "pi_stripe_abc",
"status": "requires_payment_method",
"client_secret": "pi_xxx_secret_yyy"
}

═══════════════════════════════════════════════════════════════════════════════
STEP 4: SAVE TRANSACTION + OUTBOX IN SINGLE DB TRANSACTION
═══════════════════════════════════════════════════════════════════════════════

[Aegis API] ──BEGIN TRANSACTION──▶ [Postgres]

    INSERT INTO transactions (
        id, idempotency_key, user_id, amount, currency,
        psp_reference, status, type
    ) VALUES (
        'tx_xyz', 'order_abc_123', 'm_123', 10000, 'USD',
        'pi_stripe_abc', 'pending', 'payment_intent'
    );

    INSERT INTO transaction_outbox (
        event_type, payload, partition_key, status, correlation_id
    ) VALUES (
        'payment_intent.created',
        '{"tx_id": "tx_xyz", "amount": 10000}',
        'm_123',           -- partition by merchant for ordering
        'pending',
        'tx_xyz'
    );

──COMMIT──▶

═══════════════════════════════════════════════════════════════════════════════
STEP 5: CACHE IDEMPOTENCY RESPONSE IN REDIS
═══════════════════════════════════════════════════════════════════════════════

[Aegis API] ──SETEX idempotency:m_123:order_abc_123 86400
'{"tx_id":"tx_xyz","client_secret":"pi_xxx_secret_yyy"}'
──▶ [Redis]

═══════════════════════════════════════════════════════════════════════════════
STEP 6: RETURN CLIENT SECRET TO MERCHANT
═══════════════════════════════════════════════════════════════════════════════

[Aegis API] ──HTTP 201──▶ [Merchant App]
{
"transaction_id": "tx_xyz",
"client_secret": "pi_xxx_secret_yyy",
"status": "pending"
}

[Merchant App] uses client_secret with Stripe.js to collect card details

═══════════════════════════════════════════════════════════════════════════════
STEP 7: OUTBOX RELAY PUBLISHES TO KAFKA (Background Worker)
═══════════════════════════════════════════════════════════════════════════════

[Outbox Relay Worker] (polling every 100ms)
│
├──SELECT \* FROM transaction_outbox
│ WHERE status = 'pending'
│ ORDER BY id LIMIT 100
│ FOR UPDATE SKIP LOCKED──────────────────▶ [Postgres]
│ │
│◀───────── [{id: 1, event_type: 'payment_intent.created', ...}]
│
├──PRODUCE──▶ [Kafka: aegis.transactions]
│ {
│ topic: "aegis.transactions",
│ partition_key: "m_123",
│ value: {"event": "payment_intent.created", ...}
│ }
│
└──UPDATE transaction_outbox
SET status = 'processed'
WHERE id = 1────────────────────────────▶ [Postgres]

═══════════════════════════════════════════════════════════════════════════════
STEP 8: CUSTOMER PAYS → STRIPE SENDS WEBHOOK
═══════════════════════════════════════════════════════════════════════════════

Customer enters card → Stripe charges → Stripe sends webhook

[Stripe] ──POST /webhooks/stripe──▶ [Aegis API]
{
"id": "evt_123",
"type": "payment_intent.succeeded",
"data": {
"object": {
"id": "pi_stripe_abc",
"amount": 10000,
"status": "succeeded"
}
}
} + Stripe-Signature header

═══════════════════════════════════════════════════════════════════════════════
STEP 9: VERIFY WEBHOOK + CHECK IDEMPOTENCY
═══════════════════════════════════════════════════════════════════════════════

[Aegis API]
│
├── Verify Stripe-Signature (HMAC)
│
├──GET webhook:evt_123──▶ [Redis]
│ │
│◀──────── NULL ─────────────┘
│
├──SELECT \* FROM psp_webhooks
│ WHERE event_id = 'evt_123'──▶ [Postgres]
│ │
│◀──────── NULL ─────────────────────┘
│
└── PROCEED (first time processing this webhook)

═══════════════════════════════════════════════════════════════════════════════
STEP 10: ACQUIRE DISTRIBUTED LOCK ON MERCHANT WALLET
═══════════════════════════════════════════════════════════════════════════════

[Aegis API]
│
├──SET lock:wallet:w_merchant_123
│ {owner: "worker-1", acquired: now()}
│ NX EX 30──────────────────────────▶ [Redis]
│ │
│◀──────── OK (lock acquired) ──────────────┘
│
└── PROCEED WITH WALLET UPDATE

═══════════════════════════════════════════════════════════════════════════════
STEP 11: UPDATE BALANCES + LEDGER (Single Transaction)
═══════════════════════════════════════════════════════════════════════════════

[Aegis API] ──BEGIN TRANSACTION──▶ [Postgres]

    -- 1. Mark webhook as processing
    INSERT INTO psp_webhooks (id, event_id, payload, status)
    VALUES ('wh_1', 'evt_123', '{...}', 'processing');

    -- 2. Update transaction status
    UPDATE transactions
    SET status = 'completed', updated_at = NOW()
    WHERE psp_reference = 'pi_stripe_abc';

    -- 3. Calculate platform fee (e.g., 2.9% + 30¢)
    -- amount = 10000, fee = 320 (2.9% + 30), net = 9680

    -- 4. Credit merchant's holding wallet (money lands here first)
    UPDATE wallets
    SET balance = balance + 9680, updated_at = NOW()
    WHERE user_id = 'm_123' AND type = 'holding'
    RETURNING balance AS new_balance;

    -- 5. Credit platform revenue wallet (your fee)
    UPDATE wallets
    SET balance = balance + 320, updated_at = NOW()
    WHERE type = 'revenue' AND user_id = 'platform';

    -- 6. Create ledger entries (MUST sum to zero)
    INSERT INTO ledger_entries
        (transaction_id, account_id, debit, credit, balance_after, description)
    VALUES
        -- Customer pays (external, debit = money coming in)
        ('tx_xyz', 'external_customer', 10000, 0, NULL, 'revenue'),
        -- Platform takes fee (credit to revenue)
        ('tx_xyz', 'w_platform_revenue', 0, 320, 320, 'fee'),
        -- Merchant receives net (credit to holding)
        ('tx_xyz', 'w_merchant_holding', 0, 9680, 9680, 'revenue');

    -- Verify: 10000 (debit) = 320 + 9680 (credits) ✓

    -- 7. Queue settlement event
    INSERT INTO transaction_outbox (event_type, payload, partition_key, status)
    VALUES ('payment.completed', '{"tx_id": "tx_xyz", "net": 9680}', 'm_123', 'pending');

    -- 8. Mark webhook processed
    UPDATE psp_webhooks SET status = 'processed' WHERE id = 'wh_1';

──COMMIT──▶

═══════════════════════════════════════════════════════════════════════════════
STEP 12: RELEASE LOCK + CACHE WEBHOOK
═══════════════════════════════════════════════════════════════════════════════

[Aegis API]
│
├──DEL lock:wallet:w_merchant_123──▶ [Redis]
│
└──SETEX webhook:evt_123 86400 "processed"──▶ [Redis]

═══════════════════════════════════════════════════════════════════════════════
STEP 13: SETTLEMENT (Move holding → settlement, daily batch)
═══════════════════════════════════════════════════════════════════════════════

[Settlement Worker] (runs daily at midnight)
│
├── For each merchant with holding balance > 0:
│
├──BEGIN TRANSACTION──▶ [Postgres]
│
│ -- Move from holding to settlement
│ UPDATE wallets SET balance = 0
│ WHERE user_id = 'm_123' AND type = 'holding';
│  
 │ UPDATE wallets SET balance = balance + 9680
│ WHERE user_id = 'm_123' AND type = 'settlement';
│  
 │ -- Ledger entries
│ INSERT INTO ledger_entries (...) VALUES
│ ('settle_1', 'w_merchant_holding', 9680, 0, 0, 'payout'),
│ ('settle_1', 'w_merchant_settlement', 0, 9680, 9680, 'payout');
│
│ -- Queue payout event
│ INSERT INTO transaction_outbox (event_type, ...)
│ VALUES ('settlement.ready', ...);
│
└──COMMIT──▶

═══════════════════════════════════════════════════════════════════════════════
STEP 14: PAYOUT TO MERCHANT (via PSP)
═══════════════════════════════════════════════════════════════════════════════

[Payout Worker] (consumes settlement.ready events)
│
├──CONSUME──▶ [Kafka: aegis.settlements]
│
├──POST /v1/transfers──▶ [Stripe Connect]
│ {
│ "amount": 9680,
│ "destination": "acct_merchant_stripe_id",
│ "transfer_group": "settle_1"
│ }
│ │
│◀───────────────────────────┘
│ {"id": "tr_stripe_xyz", "status": "paid"}
│
└──UPDATE + LEDGER──▶ [Postgres]

       UPDATE wallets SET balance = 0
       WHERE user_id = 'm_123' AND type = 'settlement';

       INSERT INTO ledger_entries (...) VALUES
           ('payout_1', 'w_merchant_settlement', 9680, 0, 0, 'payout'),
           ('payout_1', 'external_bank', 0, 9680, NULL, 'payout');

═══════════════════════════════════════════════════════════════════════════════
STEP 15: RECONCILIATION (Daily job)
═══════════════════════════════════════════════════════════════════════════════

[Reconciliation Worker] (runs daily)
│
│ For each merchant:
│
├── Expected balance = SUM(credits) - SUM(debits) from ledger_entries
│
├── Actual balance = wallet.balance
│
├── IF expected ≠ actual:
│ │
│ └──INSERT INTO reconciliation_runs (status) VALUES ('discrepancy');
│ INSERT INTO discrepancies (
│ reconciliation_run_id, expected_amount, actual_amount, reason
│ ) VALUES (...);
│  
 │ ALERT ops team!
│
└── ELSE:
│
└──INSERT INTO reconciliation_runs (status) VALUES ('matched');
