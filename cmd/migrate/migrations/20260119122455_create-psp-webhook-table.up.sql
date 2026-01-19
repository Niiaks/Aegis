CREATE TABLE IF NOT EXISTS psp_webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'received' CHECK (status IN ('received', 'error', 'processed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT psp_webhooks_event_id_unique UNIQUE (event_id)
);

CREATE INDEX idx_psp_webhooks_status ON psp_webhooks(status);
CREATE INDEX idx_psp_webhooks_event_id ON psp_webhooks(event_id);
CREATE INDEX idx_psp_webhooks_created_at ON psp_webhooks(created_at);
