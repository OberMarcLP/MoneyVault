DROP INDEX IF EXISTS idx_accounts_not_deleted;
DROP INDEX IF EXISTS idx_transactions_not_deleted;
DROP INDEX IF EXISTS idx_budgets_not_deleted;
DROP INDEX IF EXISTS idx_recurring_not_deleted;
DROP INDEX IF EXISTS idx_holdings_not_deleted;
DROP INDEX IF EXISTS idx_dividends_not_deleted;
DROP INDEX IF EXISTS idx_alert_rules_not_deleted;

ALTER TABLE accounts DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE transactions DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE budgets DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE recurring_transactions DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE holdings DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE dividends DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE alert_rules DROP COLUMN IF EXISTS deleted_at;
