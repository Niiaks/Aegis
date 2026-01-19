CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    locked_balance BIGINT NOT NULL DEFAULT 0 CHECK (locked_balance >= 0),
    currency CHAR(3) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('holding', 'settlement', 'revenue')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT wallets_user_currency_type_unique UNIQUE (user_id, currency, type)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_type ON wallets(type);
CREATE INDEX idx_wallets_currency ON wallets(currency);
