CREATE TABLE net_worth_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    total_value NUMERIC(20, 2) NOT NULL DEFAULT 0,
    accounts_value NUMERIC(20, 2) NOT NULL DEFAULT 0,
    investments_value NUMERIC(20, 2) NOT NULL DEFAULT 0,
    crypto_value NUMERIC(20, 2) NOT NULL DEFAULT 0,
    breakdown JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_networth_user_date ON net_worth_snapshots(user_id, date);
CREATE INDEX idx_networth_user_id ON net_worth_snapshots(user_id);
