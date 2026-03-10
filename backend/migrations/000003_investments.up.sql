CREATE TABLE holdings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    asset_type TEXT NOT NULL CHECK (asset_type IN ('stock', 'etf', 'crypto', 'mutual_fund')),
    symbol TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    quantity NUMERIC(20, 8) NOT NULL,
    cost_basis NUMERIC(15, 4) NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    acquired_at DATE NOT NULL DEFAULT CURRENT_DATE,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_holdings_user_id ON holdings(user_id);
CREATE INDEX idx_holdings_symbol ON holdings(symbol);
CREATE INDEX idx_holdings_account_id ON holdings(account_id);

CREATE TABLE price_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol TEXT NOT NULL,
    asset_type TEXT NOT NULL DEFAULT 'stock',
    price NUMERIC(20, 8) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    change_percent NUMERIC(10, 4) NOT NULL DEFAULT 0,
    name TEXT NOT NULL DEFAULT '',
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(symbol, asset_type)
);

CREATE INDEX idx_price_cache_symbol ON price_cache(symbol);

CREATE TABLE price_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol TEXT NOT NULL,
    asset_type TEXT NOT NULL DEFAULT 'stock',
    price NUMERIC(20, 8) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    date DATE NOT NULL,
    source TEXT NOT NULL DEFAULT 'yahoo',
    UNIQUE(symbol, asset_type, date)
);

CREATE INDEX idx_price_history_symbol_date ON price_history(symbol, date);

CREATE TABLE trade_lots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    holding_id UUID NOT NULL REFERENCES holdings(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quantity NUMERIC(20, 8) NOT NULL,
    cost_per_unit NUMERIC(15, 4) NOT NULL,
    acquired_at DATE NOT NULL,
    sold_at DATE,
    sold_price NUMERIC(15, 4),
    sold_quantity NUMERIC(20, 8),
    is_closed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trade_lots_holding_id ON trade_lots(holding_id);
CREATE INDEX idx_trade_lots_user_id ON trade_lots(user_id);
