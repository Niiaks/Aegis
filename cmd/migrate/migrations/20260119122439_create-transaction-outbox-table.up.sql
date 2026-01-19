CREATE TABLE IF NOT EXISTS transaction_outbox (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    partition_key VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processed', 'failed', 'dlq')),
    correlation_id UUID NOT NULL,
    retry_count INT NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
    last_error TEXT,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    max_retries INT NOT NULL DEFAULT 5 CHECK (max_retries >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_transaction_outbox_status ON transaction_outbox(status);
CREATE INDEX idx_transaction_outbox_next_retry_at ON transaction_outbox(next_retry_at) WHERE status = 'pending';
CREATE INDEX idx_transaction_outbox_correlation_id ON transaction_outbox(correlation_id);
CREATE INDEX idx_transaction_outbox_event_type ON transaction_outbox(event_type);
