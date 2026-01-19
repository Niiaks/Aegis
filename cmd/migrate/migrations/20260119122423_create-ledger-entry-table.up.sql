CREATE TABLE IF NOT EXISTS ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT,
    account_id UUID NOT NULL,
    debit BIGINT NOT NULL DEFAULT 0 CHECK (debit >= 0),
    credit BIGINT NOT NULL DEFAULT 0 CHECK (credit >= 0),
    balance_after BIGINT NOT NULL CHECK (balance_after >= 0),
    description VARCHAR(50) NOT NULL CHECK (description IN ('revenue', 'payout', 'fee', 'refund')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT ledger_entries_debit_or_credit CHECK (debit = 0 OR credit = 0)
);

CREATE INDEX idx_ledger_entries_transaction_id ON ledger_entries(transaction_id);
CREATE INDEX idx_ledger_entries_account_id ON ledger_entries(account_id);
CREATE INDEX idx_ledger_entries_created_at ON ledger_entries(created_at);
