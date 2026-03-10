export type UserRole = 'admin' | 'user';

export interface User {
  id: string;
  email: string;
  role: UserRole;
  totp_enabled: boolean;
  email_verified: boolean;
  e2e_enabled: boolean;
  preferences: UserPreferences;
  created_at: string;
  updated_at: string;
}

export interface LoginResponse {
  access_token: string;
  user: User;
  e2e_encrypted_dek?: string;
  e2e_kek_salt?: string;
}

export interface E2EMigrateAccount {
  id: string;
  name: string;
  balance: string;
}

export interface E2EMigrateTransaction {
  id: string;
  amount: string;
  description: string;
}

export interface E2EMigrateData {
  accounts: E2EMigrateAccount[];
  transactions: E2EMigrateTransaction[];
}

export interface AuditLog {
  id: string;
  user_id: string | null;
  action: string;
  resource_type: string | null;
  resource_id: string | null;
  ip_address: string | null;
  user_agent: string | null;
  details: Record<string, unknown> | null;
  created_at: string;
}

export interface UserPreferences {
  theme: 'light' | 'dark' | 'system';
  currency: string;
  locale: string;
  onboarding_dismissed?: boolean;
}

export interface WebAuthnCredential {
  id: string;
  name: string;
  created_at: string;
  last_used_at: string | null;
}

export type AccountType = 'checking' | 'savings' | 'credit' | 'investment' | 'crypto_wallet';

export interface Account {
  id: string;
  user_id: string;
  name: string;
  type: AccountType;
  currency: string;
  balance: string;
  institution: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export type TransactionType = 'income' | 'expense' | 'transfer';

export interface Transaction {
  id: string;
  account_id: string;
  user_id: string;
  type: TransactionType;
  amount: string;
  currency: string;
  category_id: string | null;
  description: string;
  date: string;
  tags: string[];
  import_source: string;
  transfer_account_id?: string;
  created_at: string;
}

export type CategoryType = 'income' | 'expense';

export interface Category {
  id: string;
  user_id: string;
  name: string;
  type: CategoryType;
  icon: string;
  color: string;
  parent_id: string | null;
  created_at: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface TransactionFilter {
  account_id?: string;
  type?: TransactionType;
  category_id?: string;
  date_from?: string;
  date_to?: string;
  search?: string;
  page?: number;
  per_page?: number;
}

export interface LoginRequest {
  email: string;
  password: string;
  totp_code?: string;
}

export interface CreateAccountRequest {
  name: string;
  type: AccountType;
  currency: string;
  balance: string;
  institution?: string;
}

export interface CreateTransactionRequest {
  account_id: string;
  type: TransactionType;
  amount: string;
  currency: string;
  category_id?: string;
  description?: string;
  date: string;
  tags?: string[];
  transfer_account_id?: string;
}

export interface CreateCategoryRequest {
  name: string;
  type: CategoryType;
  icon: string;
  color: string;
  parent_id?: string;
}

// Budgets
export type BudgetPeriod = 'weekly' | 'monthly' | 'yearly';

export interface Budget {
  id: string;
  user_id: string;
  category_id: string | null;
  amount: number;
  period: BudgetPeriod;
  start_date: string;
  end_date: string | null;
  rollover: boolean;
  spent: number;
  remaining: number;
  percentage: number;
  category_name: string;
  category_icon: string;
  category_color: string;
  created_at: string;
}

export interface CreateBudgetRequest {
  category_id: string;
  amount: number;
  period: BudgetPeriod;
  start_date: string;
  end_date?: string;
  rollover: boolean;
}

// Recurring Transactions
export type Frequency = 'daily' | 'weekly' | 'biweekly' | 'monthly' | 'quarterly' | 'yearly';

export interface RecurringTransaction {
  id: string;
  user_id: string;
  account_id: string;
  type: TransactionType;
  amount: string;
  currency: string;
  category_id: string | null;
  description: string;
  frequency: Frequency;
  next_date: string;
  end_date: string | null;
  transfer_account_id: string | null;
  is_active: boolean;
  last_created: string | null;
  created_at: string;
}

export interface CreateRecurringRequest {
  account_id: string;
  type: TransactionType;
  amount: string;
  currency: string;
  category_id?: string;
  description?: string;
  frequency: Frequency;
  next_date: string;
  end_date?: string;
  transfer_account_id?: string;
}

// Investments
export type AssetType = 'stock' | 'etf' | 'crypto' | 'mutual_fund' | 'defi_position';

export interface Holding {
  id: string;
  user_id: string;
  account_id: string;
  asset_type: AssetType;
  symbol: string;
  name: string;
  quantity: number;
  cost_basis: number;
  currency: string;
  acquired_at: string;
  notes: string;
  current_price: number;
  market_value: number;
  total_return: number;
  return_percent: number;
  day_change: number;
  asset_name: string;
  created_at: string;
  updated_at: string;
}

export interface CreateHoldingRequest {
  account_id: string;
  asset_type: AssetType;
  symbol: string;
  name?: string;
  quantity: number;
  cost_basis: number;
  currency?: string;
  acquired_at: string;
  notes?: string;
  token_address?: string;
  network?: string;
  metadata?: string;
}

export interface UpdateHoldingRequest {
  quantity?: number;
  cost_basis?: number;
  notes?: string;
}

export type CostBasisMethod = 'fifo' | 'lifo' | 'average';

export interface SellHoldingRequest {
  quantity: number;
  price: number;
  sold_at: string;
  method?: CostBasisMethod;
}

export interface PortfolioSummary {
  total_value: number;
  total_cost: number;
  total_return: number;
  total_return_pct: number;
  day_change: number;
  day_change_pct: number;
  holdings_count: number;
}

export interface Dividend {
  id: string;
  holding_id: string;
  user_id: string;
  amount: string;
  currency: string;
  ex_date: string;
  pay_date: string | null;
  notes: string;
  created_at: string;
}

export interface CreateDividendRequest {
  holding_id: string;
  amount: string;
  currency?: string;
  ex_date: string;
  pay_date?: string;
  notes?: string;
}

export interface DividendSummary {
  total_dividends: number;
  dividends_ytd: number;
  dividend_count: number;
  last_dividend_at?: string;
}

export interface TradeLot {
  id: string;
  holding_id: string;
  quantity: number;
  cost_per_unit: number;
  acquired_at: string;
  sold_at: string | null;
  sold_price: number | null;
  sold_quantity: number | null;
  is_closed: boolean;
}

export interface PriceHistory {
  id: string;
  symbol: string;
  price: number;
  currency: string;
  date: string;
  source: string;
}

// CSV Import
export interface CSVPreview {
  headers: string[];
  rows: { values: Record<string, string> }[];
  total: number;
}

export interface CSVImportMapping {
  date: string;
  amount: string;
  description: string;
  merchant?: string;
  category?: string;
  sub_category?: string;
  type?: string;
  status?: string;
  currency?: string;
}

export interface ImportJob {
  id: string;
  filename: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  total_rows: number;
  imported_rows: number;
  duplicate_rows: number;
  error_message: string | null;
  created_at: string;
}

// Crypto
export interface CryptoSummary {
  total_value: number;
  total_cost: number;
  total_return: number;
  total_return_pct: number;
  defi_value: number;
  defi_positions: number;
  token_count: number;
  wallet_count: number;
  total_gas_fees: number;
}

export interface Wallet {
  id: string;
  user_id: string;
  address: string;
  network: string;
  label: string;
  is_active: boolean;
  last_synced: string | null;
  created_at: string;
}

export interface WalletTransaction {
  id: string;
  wallet_id: string;
  tx_hash: string;
  block_number: number;
  from_address: string;
  to_address: string;
  value: string;
  token_symbol: string;
  token_address: string;
  gas_used: number;
  gas_price: string;
  gas_fee_eth: number;
  tx_type: string;
  timestamp: string;
}

export interface CoinGeckoToken {
  id: string;
  symbol: string;
  name: string;
}

export interface DeFiMetadata {
  protocol: string;
  pool_name: string;
  token0: string;
  token1: string;
  apy: number;
  rewards_token: string;
  position_type: string;
}

// Analytics
export interface NetWorthSnapshot {
  id: string;
  date: string;
  total_value: number;
  accounts_value: number;
  investments_value: number;
  crypto_value: number;
  breakdown: Record<string, number>;
}

export interface SpendingByCategory {
  category_id: string;
  category_name: string;
  category_color: string;
  category_icon: string;
  total: number;
  count: number;
}

export interface SpendingTrend {
  period: string;
  income: number;
  expense: number;
  net: number;
}

export interface TopExpense {
  description: string;
  amount: number;
  date: string;
  category: string;
}

export interface BudgetVsActual {
  category_id: string;
  category_name: string;
  category_color: string;
  budget_amount: number;
  actual_amount: number;
  difference: number;
  percentage: number;
}

export interface BudgetHistory {
  period: string;
  budgets: BudgetVsActual[];
  total_budget: number;
  total_actual: number;
}

export interface CashFlowForecast {
  period: string;
  projected_income: number;
  projected_expense: number;
  net_cash_flow: number;
  running_balance: number;
}

export interface UpcomingBill {
  description: string;
  amount: number;
  due_date: string;
  frequency: string;
  account_name: string;
}

export interface CashFlowResult {
  forecast: CashFlowForecast[];
  runway: {
    monthly_savings: number;
    monthly_income: number;
    monthly_expenses: number;
    current_balance: number;
    runway_months: number;
  };
  upcoming_bills: UpcomingBill[];
}

export interface AssetAllocation {
  asset_type: string;
  value: number;
  percentage: number;
  count: number;
}

// Notifications
export type NotificationType = 'budget_alert' | 'price_alert' | 'milestone' | 'info' | 'import_complete' | 'summary';

export interface Notification {
  id: string;
  user_id: string;
  type: NotificationType;
  title: string;
  message: string;
  is_read: boolean;
  link: string;
  created_at: string;
}

export type AlertRuleType = 'budget_overspend' | 'price_drop' | 'price_rise' | 'net_worth_milestone';

export interface AlertRule {
  id: string;
  user_id: string;
  type: AlertRuleType;
  condition: Record<string, unknown>;
  is_active: boolean;
  last_triggered: string | null;
  created_at: string;
}

export interface CreateAlertRuleRequest {
  type: AlertRuleType;
  condition: Record<string, unknown>;
}

// Exchange Connections
export type ExchangeName = 'binance' | 'coinbase' | 'kraken';

export interface ExchangeConnection {
  id: string;
  user_id: string;
  exchange: ExchangeName;
  label: string;
  is_active: boolean;
  last_synced: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateExchangeConnectionRequest {
  exchange: ExchangeName;
  api_key: string;
  api_secret: string;
  label?: string;
}

export interface ExchangeBalance {
  symbol: string;
  free: number;
  locked: number;
  total: number;
  usd_value?: number;
}

export interface ExchangeSyncResult {
  connection_id: string;
  exchange: ExchangeName;
  balances: ExchangeBalance[];
  synced_at: string;
}
