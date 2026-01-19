CREATE TABLE IF NOT EXISTS reconciliation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('discrepancy', 'matched')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_reconciliation_runs_run_date ON reconciliation_runs(run_date);
CREATE INDEX idx_reconciliation_runs_status ON reconciliation_runs(status);

CREATE TABLE IF NOT EXISTS discrepancies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reconciliation_run_id UUID NOT NULL REFERENCES reconciliation_runs(id) ON DELETE CASCADE,
    expected_amount BIGINT NOT NULL,
    actual_amount BIGINT NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'unresolved' CHECK (status IN ('unresolved', 'resolved')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_discrepancies_reconciliation_run_id ON discrepancies(reconciliation_run_id);
CREATE INDEX idx_discrepancies_status ON discrepancies(status);
