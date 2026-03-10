CREATE TABLE IF NOT EXISTS dividends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    holding_id UUID NOT NULL REFERENCES holdings(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount TEXT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    ex_date DATE NOT NULL,
    pay_date DATE,
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dividends_user ON dividends(user_id);
CREATE INDEX IF NOT EXISTS idx_dividends_holding ON dividends(holding_id);
CREATE INDEX IF NOT EXISTS idx_dividends_user_date ON dividends(user_id, ex_date DESC);
