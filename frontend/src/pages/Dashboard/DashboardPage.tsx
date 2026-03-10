import { Link } from 'react-router-dom';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { CardSkeleton, ListSkeleton } from '@/components/ui/skeleton';
import { OnboardingChecklist } from '@/components/dashboard/OnboardingChecklist';
import { useAccounts, useTransactions, useBudgets, usePortfolioSummary, useCryptoSummary, useExchangeRates } from '@/api/hooks';
import { useAuthStore } from '@/stores/auth';
import { formatCurrency, formatDate, getAccountTypeLabel } from '@/lib/utils';
import { Wallet, TrendingUp, TrendingDown, ArrowLeftRight, Plus, PieChart, AlertTriangle, Briefcase, ArrowUpRight, ArrowDownRight, Bitcoin, Layers } from 'lucide-react';

export function DashboardPage() {
  const { data: accountsData, isLoading: accountsLoading } = useAccounts();
  const { data: txData, isLoading: txLoading } = useTransactions({ per_page: 5 });
  const { data: budgetData } = useBudgets();
  const { data: portfolio } = usePortfolioSummary();
  const { data: cryptoSummary } = useCryptoSummary();
  const user = useAuthStore((s) => s.user);
  const baseCurrency = user?.preferences?.currency || 'USD';
  const { data: ratesData } = useExchangeRates(baseCurrency);
  const rates = ratesData?.rates;

  const accounts = accountsData?.accounts ?? [];
  const transactions = (txData?.data ?? []);
  const budgets = budgetData?.budgets ?? [];

  const totalBalance = accounts.reduce((sum, a) => {
    const bal = parseFloat(a.balance || '0');
    if (a.currency !== baseCurrency && rates && rates[a.currency]) {
      return sum + bal * (1 / rates[a.currency]);
    }
    return sum + bal;
  }, 0);
  const activeAccounts = accounts.filter((a) => a.is_active).length;
  const hasMultipleCurrencies = new Set(accounts.map(a => a.currency)).size > 1;

  const income = transactions.filter((t) => t.type === 'income').reduce((s, t) => s + parseFloat(t.amount || '0'), 0);
  const expenses = transactions.filter((t) => t.type === 'expense').reduce((s, t) => s + parseFloat(t.amount || '0'), 0);

  const overBudgetCount = budgets.filter((b) => b.percentage >= 100).length;
  const warningBudgetCount = budgets.filter((b) => b.percentage >= 80 && b.percentage < 100).length;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">Your financial overview</p>
      </div>

      {!accountsLoading && (
        <OnboardingChecklist
          hasAccounts={accounts.length > 0}
          hasTransactions={transactions.length > 0}
          hasBudgets={budgets.length > 0}
          serverDismissed={user?.preferences?.onboarding_dismissed}
        />
      )}

      {(accountsLoading || txLoading) ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => <CardSkeleton key={i} />)}
        </div>
      ) : (
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Net Worth</CardTitle>
            <Wallet className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(totalBalance, baseCurrency)}</div>
            <p className="text-xs text-muted-foreground">
              {activeAccounts} active accounts{hasMultipleCurrencies && ' (converted)'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Income</CardTitle>
            <TrendingUp className="h-4 w-4 text-success" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-success">{formatCurrency(income)}</div>
            <p className="text-xs text-muted-foreground">Recent transactions</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Expenses</CardTitle>
            <TrendingDown className="h-4 w-4 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-destructive">{formatCurrency(expenses)}</div>
            <p className="text-xs text-muted-foreground">Recent transactions</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Transactions</CardTitle>
            <ArrowLeftRight className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{txData?.total ?? 0}</div>
            <p className="text-xs text-muted-foreground">Total recorded</p>
          </CardContent>
        </Card>
      </div>
      )}

      {budgets.length > 0 && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg">Budget Overview</CardTitle>
            <Link to="/budgets">
              <Button variant="outline" size="sm"><PieChart className="mr-1 h-3 w-3" /> View All</Button>
            </Link>
          </CardHeader>
          <CardContent>
            {overBudgetCount > 0 && (
              <div className="flex items-center gap-2 mb-4 rounded-lg border border-destructive/30 bg-destructive/5 p-3">
                <AlertTriangle className="h-4 w-4 text-destructive shrink-0" />
                <p className="text-sm text-destructive">
                  {overBudgetCount} budget{overBudgetCount > 1 ? 's' : ''} exceeded this period
                  {warningBudgetCount > 0 && `, ${warningBudgetCount} approaching limit`}
                </p>
              </div>
            )}
            <div className="space-y-3">
              {budgets.slice(0, 5).map((b) => (
                <div key={b.id} className="space-y-1.5">
                  <div className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-2">
                      <div className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: b.category_color || '#94a3b8' }} />
                      <span className="font-medium">{b.category_name}</span>
                    </div>
                    <span className="text-muted-foreground">
                      {formatCurrency(b.spent)} / {formatCurrency(b.amount)}
                    </span>
                  </div>
                  <div className="h-2 w-full rounded-full bg-muted overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all ${
                        b.percentage >= 100 ? 'bg-destructive' : b.percentage >= 80 ? 'bg-amber-500' : 'bg-primary'
                      }`}
                      style={{ width: `${Math.min(b.percentage, 100)}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {portfolio && portfolio.holdings_count > 0 && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg">Portfolio Snapshot</CardTitle>
            <Link to="/investments">
              <Button variant="outline" size="sm"><Briefcase className="mr-1 h-3 w-3" /> View All</Button>
            </Link>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-3">
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Total Value</p>
                <p className="text-xl font-bold">{formatCurrency(portfolio.total_value)}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Total Return</p>
                <div className="flex items-center gap-1">
                  {portfolio.total_return >= 0 ? (
                    <ArrowUpRight className="h-4 w-4 text-success" />
                  ) : (
                    <ArrowDownRight className="h-4 w-4 text-destructive" />
                  )}
                  <p className={`text-xl font-bold ${portfolio.total_return >= 0 ? 'text-success' : 'text-destructive'}`}>
                    {portfolio.total_return >= 0 ? '+' : ''}{formatCurrency(portfolio.total_return)}
                  </p>
                </div>
                <p className={`text-xs ${portfolio.total_return_pct >= 0 ? 'text-success' : 'text-destructive'}`}>
                  {portfolio.total_return_pct >= 0 ? '+' : ''}{portfolio.total_return_pct.toFixed(2)}%
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Day Change</p>
                <div className="flex items-center gap-1">
                  {portfolio.day_change >= 0 ? (
                    <ArrowUpRight className="h-4 w-4 text-success" />
                  ) : (
                    <ArrowDownRight className="h-4 w-4 text-destructive" />
                  )}
                  <p className={`text-xl font-bold ${portfolio.day_change >= 0 ? 'text-success' : 'text-destructive'}`}>
                    {portfolio.day_change >= 0 ? '+' : ''}{formatCurrency(portfolio.day_change)}
                  </p>
                </div>
                <p className={`text-xs ${portfolio.day_change_pct >= 0 ? 'text-success' : 'text-destructive'}`}>
                  {portfolio.day_change_pct >= 0 ? '+' : ''}{portfolio.day_change_pct.toFixed(2)}%
                </p>
              </div>
            </div>
            <p className="text-xs text-muted-foreground mt-3">{portfolio.holdings_count} holding{portfolio.holdings_count !== 1 ? 's' : ''} tracked</p>
          </CardContent>
        </Card>
      )}

      {cryptoSummary && cryptoSummary.token_count > 0 && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg">Crypto Summary</CardTitle>
            <Link to="/crypto">
              <Button variant="outline" size="sm"><Bitcoin className="mr-1 h-3 w-3" /> View All</Button>
            </Link>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-4">
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Crypto Value</p>
                <p className="text-xl font-bold">{formatCurrency(cryptoSummary.total_value)}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Return</p>
                <div className="flex items-center gap-1">
                  {cryptoSummary.total_return >= 0 ? (
                    <ArrowUpRight className="h-4 w-4 text-success" />
                  ) : (
                    <ArrowDownRight className="h-4 w-4 text-destructive" />
                  )}
                  <p className={`text-xl font-bold ${cryptoSummary.total_return >= 0 ? 'text-success' : 'text-destructive'}`}>
                    {cryptoSummary.total_return >= 0 ? '+' : ''}{formatCurrency(cryptoSummary.total_return)}
                  </p>
                </div>
              </div>
              <div className="space-y-1">
                <div className="flex items-center gap-1">
                  <Layers className="h-3 w-3 text-muted-foreground" />
                  <p className="text-sm text-muted-foreground">DeFi</p>
                </div>
                <p className="text-xl font-bold">{formatCurrency(cryptoSummary.defi_value)}</p>
                <p className="text-xs text-muted-foreground">{cryptoSummary.defi_positions} positions</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Overview</p>
                <p className="text-sm">{cryptoSummary.token_count} tokens &middot; {cryptoSummary.wallet_count} wallets</p>
                <p className="text-xs text-muted-foreground">Gas: {cryptoSummary.total_gas_fees.toFixed(4)} ETH</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg">Accounts</CardTitle>
            <Link to="/accounts">
              <Button variant="outline" size="sm"><Plus className="mr-1 h-3 w-3" /> Add</Button>
            </Link>
          </CardHeader>
          <CardContent>
            {accountsLoading ? (
              <ListSkeleton rows={3} />
            ) : accounts.length === 0 ? (
              <p className="text-muted-foreground">No accounts yet. Create one to get started.</p>
            ) : (
              <div className="space-y-3">
                {accounts.map((account) => (
                  <div key={account.id} className="flex items-center justify-between rounded-lg border p-3">
                    <div>
                      <p className="font-medium">{account.name}</p>
                      <p className="text-sm text-muted-foreground">{getAccountTypeLabel(account.type)}</p>
                    </div>
                    <p className={`font-semibold ${parseFloat(account.balance) >= 0 ? 'text-success' : 'text-destructive'}`}>
                      {formatCurrency(account.balance, account.currency)}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg">Recent Transactions</CardTitle>
            <Link to="/transactions">
              <Button variant="outline" size="sm"><Plus className="mr-1 h-3 w-3" /> Add</Button>
            </Link>
          </CardHeader>
          <CardContent>
            {txLoading ? (
              <ListSkeleton rows={3} />
            ) : transactions.length === 0 ? (
              <p className="text-muted-foreground">No transactions yet.</p>
            ) : (
              <div className="space-y-3">
                {transactions.map((tx) => (
                  <div key={tx.id} className="flex items-center justify-between rounded-lg border p-3">
                    <div>
                      <p className="font-medium">{tx.description || 'No description'}</p>
                      <p className="text-sm text-muted-foreground">{formatDate(tx.date)}</p>
                    </div>
                    <p className={`font-semibold ${
                      tx.type === 'income' ? 'text-success' : tx.type === 'expense' ? 'text-destructive' : 'text-primary'
                    }`}>
                      {tx.type === 'income' ? '+' : tx.type === 'expense' ? '-' : ''}
                      {formatCurrency(tx.amount, tx.currency)}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
