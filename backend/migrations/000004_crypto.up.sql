ALTER TABLE holdings ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}';
ALTER TABLE holdings ADD COLUMN IF NOT EXISTS token_address TEXT NOT NULL DEFAULT '';
ALTER TABLE holdings ADD COLUMN IF NOT EXISTS network TEXT NOT NULL DEFAULT '';

ALTER TABLE holdings DROP CONSTRAINT IF EXISTS holdings_asset_type_check;
ALTER TABLE holdings ADD CONSTRAINT holdings_asset_type_check
  CHECK (asset_type IN ('stock', 'etf', 'crypto', 'mutual_fund', 'defi_position'));

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address TEXT NOT NULL,
    network TEXT NOT NULL DEFAULT 'ethereum',
    label TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_synced TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE UNIQUE INDEX idx_wallets_user_address ON wallets(user_id, address, network);

CREATE TABLE wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tx_hash TEXT NOT NULL,
    block_number BIGINT NOT NULL DEFAULT 0,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    value TEXT NOT NULL DEFAULT '0',
    token_symbol TEXT NOT NULL DEFAULT 'ETH',
    token_address TEXT NOT NULL DEFAULT '',
    gas_used BIGINT NOT NULL DEFAULT 0,
    gas_price TEXT NOT NULL DEFAULT '0',
    gas_fee_eth NUMERIC(20, 18) NOT NULL DEFAULT 0,
    tx_type TEXT NOT NULL DEFAULT 'transfer',
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallet_tx_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_tx_user_id ON wallet_transactions(user_id);
CREATE UNIQUE INDEX idx_wallet_tx_hash ON wallet_transactions(wallet_id, tx_hash);
