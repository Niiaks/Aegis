CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL CHECK (amount >= 0),
    currency CHAR(3) NOT NULL,
    psp_reference VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    type VARCHAR(20) NOT NULL CHECK (type IN ('payment_intent', 'payout', 'refund', 'fee')),
    failure_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT transactions_idempotency_key_unique UNIQUE (idempotency_key)
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_psp_reference ON transactions(psp_reference);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
