-- Add soft delete column to major tables
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE budgets ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE recurring_transactions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE holdings ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE dividends ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE alert_rules ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

-- Partial indexes for efficient queries on non-deleted rows
CREATE INDEX IF NOT EXISTS idx_accounts_not_deleted ON accounts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_not_deleted ON transactions(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_budgets_not_deleted ON budgets(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_recurring_not_deleted ON recurring_transactions(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_holdings_not_deleted ON holdings(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dividends_not_deleted ON dividends(holding_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_alert_rules_not_deleted ON alert_rules(user_id) WHERE deleted_at IS NULL;
