import { lazy, Suspense, useState } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { CardSkeleton } from '@/components/ui/skeleton';
import { useCryptoSummary, useRefreshCryptoPrices } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { formatCurrency } from '@/lib/utils';
import {
  RefreshCw, Wallet, TrendingUp, TrendingDown,
  Bitcoin, Layers, Fuel,
} from 'lucide-react';

const CryptoPortfolio = lazy(() => import('./CryptoPortfolio'));
const CryptoDeFi = lazy(() => import('./CryptoDeFi'));
const CryptoWallets = lazy(() => import('./CryptoWallets'));

type Tab = 'portfolio' | 'defi' | 'wallets';

function SummaryCard({ label, value, sub, positive, icon: Icon }: {
  label: string; value: string; sub?: string; positive?: boolean;
  icon: React.ElementType;
}) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-center justify-between mb-2">
          <span className="text-xs text-muted-foreground">{label}</span>
          <Icon className="h-4 w-4 text-muted-foreground" />
        </div>
        <p className="text-xl font-bold">{value}</p>
        {sub && (
          <p className={`text-xs mt-1 ${positive !== undefined ? (positive ? 'text-success' : 'text-destructive') : 'text-muted-foreground'}`}>
            {sub}
          </p>
        )}
      </CardContent>
    </Card>
  );
}

export function CryptoPage() {
  const { data: summary, isLoading: summaryLoading } = useCryptoSummary();
  const refreshPrices = useRefreshCryptoPrices();
  const { toast } = useToast();

  const [tab, setTab] = useState<Tab>('portfolio');

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Crypto</h1>
          <p className="text-muted-foreground">Track your crypto portfolio, DeFi positions, and wallets</p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => refreshPrices.mutate(undefined, {
            onSuccess: () => toast('Crypto prices refreshed', 'success'),
            onError: (e) => toast(e.message, 'error'),
          })}
          loading={refreshPrices.isPending}
        >
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      {summaryLoading ? (
        <div className="grid gap-4 grid-cols-2 lg:grid-cols-5">
          {[...Array(5)].map((_, i) => <CardSkeleton key={i} />)}
        </div>
      ) : summary && (
        <div className="grid gap-4 grid-cols-2 lg:grid-cols-5">
          <SummaryCard label="Crypto Value" value={formatCurrency(summary.total_value)} icon={Bitcoin} />
          <SummaryCard
            label="Total Return"
            value={formatCurrency(summary.total_return)}
            sub={`${summary.total_return_pct >= 0 ? '+' : ''}${summary.total_return_pct.toFixed(2)}%`}
            positive={summary.total_return >= 0}
            icon={summary.total_return >= 0 ? TrendingUp : TrendingDown}
          />
          <SummaryCard label="DeFi Value" value={formatCurrency(summary.defi_value)} sub={`${summary.defi_positions} positions`} icon={Layers} />
          <SummaryCard label="Wallets" value={String(summary.wallet_count)} sub={`${summary.token_count} tokens`} icon={Wallet} />
          <SummaryCard label="Gas Fees" value={`${summary.total_gas_fees.toFixed(6)} ETH`} icon={Fuel} />
        </div>
      )}

      <div className="flex gap-1 border-b">
        {(['portfolio', 'defi', 'wallets'] as Tab[]).map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              tab === t ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            {t === 'portfolio' ? 'Portfolio' : t === 'defi' ? 'DeFi Positions' : 'Wallets'}
          </button>
        ))}
      </div>

      <Suspense fallback={<CardSkeleton />}>
        {tab === 'portfolio' && <CryptoPortfolio />}
        {tab === 'defi' && <CryptoDeFi />}
        {tab === 'wallets' && <CryptoWallets />}
      </Suspense>
    </div>
  );
}
