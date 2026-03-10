DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS wallets;

ALTER TABLE holdings DROP CONSTRAINT IF EXISTS holdings_asset_type_check;
ALTER TABLE holdings ADD CONSTRAINT holdings_asset_type_check
  CHECK (asset_type IN ('stock', 'etf', 'crypto', 'mutual_fund'));

ALTER TABLE holdings DROP COLUMN IF EXISTS metadata;
ALTER TABLE holdings DROP COLUMN IF EXISTS token_address;
ALTER TABLE holdings DROP COLUMN IF EXISTS network;
