CREATE TABLE IF NOT EXISTS idempotency_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(255) NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    request_path VARCHAR(255) NOT NULL,
    response_code INT,
    response_body JSONB,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT idempotency_keys_key_path_unique UNIQUE (key, request_path)
);

CREATE INDEX idx_idempotency_keys_key ON idempotency_keys(key);
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);
CREATE INDEX idx_idempotency_keys_user_id ON idempotency_keys(user_id);
