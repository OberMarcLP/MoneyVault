CREATE TABLE IF NOT EXISTS exchange_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exchange VARCHAR(20) NOT NULL CHECK (exchange IN ('binance', 'coinbase', 'kraken')),
    api_key TEXT NOT NULL,
    api_secret TEXT NOT NULL,
    label VARCHAR(100) DEFAULT '',
    is_active BOOLEAN DEFAULT true,
    last_synced TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, exchange, label)
);

CREATE INDEX idx_exchange_connections_user ON exchange_connections(user_id);
