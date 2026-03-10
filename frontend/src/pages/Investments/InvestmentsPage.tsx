import { useState, type FormEvent } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/ConfirmDialog';
import { Badge } from '@/components/ui/badge';
import { ListSkeleton, CardSkeleton } from '@/components/ui/skeleton';
import {
  useHoldings, usePortfolioSummary, useCreateHolding, useUpdateHolding,
  useDeleteHolding, useSellHolding, useRefreshPrices, useAccounts,
  useDividends, useDividendSummary, useCreateDividend, useDeleteDividend,
} from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { formatCurrency, formatDate } from '@/lib/utils';
import {
  Plus, Trash2, TrendingUp, TrendingDown, RefreshCw,
  DollarSign, BarChart3, ArrowUpRight, ArrowDownRight, Briefcase, Coins,
} from 'lucide-react';
import type { Holding, AssetType, CreateDividendRequest, CostBasisMethod } from '@/types';

export function InvestmentsPage() {
  const { data: holdingsData, isLoading } = useHoldings();
  const { data: summary, isLoading: summaryLoading } = usePortfolioSummary();
  const { data: accountData } = useAccounts();
  const createHolding = useCreateHolding();
  const updateHolding = useUpdateHolding();
  const deleteHolding = useDeleteHolding();
  const sellHolding = useSellHolding();
  const refreshPrices = useRefreshPrices();
  const deleteDividend = useDeleteDividend();
  const { toast } = useToast();

  const [showAdd, setShowAdd] = useState(false);
  const [showSell, setShowSell] = useState<Holding | null>(null);
  const [editing, setEditing] = useState<Holding | null>(null);
  const [deletingHolding, setDeletingHolding] = useState<{ id: string; symbol: string } | null>(null);
  const [deletingDividend, setDeletingDividend] = useState<{ id: string; symbol: string } | null>(null);

  const holdings = holdingsData ?? [];
  const accounts = (accountData?.accounts ?? []).filter(a => a.type === 'investment' || a.type === 'crypto_wallet');

  const handleDeleteConfirm = () => {
    if (!deletingHolding) return;
    deleteHolding.mutate(deletingHolding.id, {
      onSuccess: () => { toast(`${deletingHolding.symbol} deleted`, 'success'); setDeletingHolding(null); },
      onError: (e) => toast(e.message, 'error'),
    });
  };

  const handleDelete = (id: string, symbol: string) => {
    setDeletingHolding({ id, symbol });
  };

  const handleRefresh = () => {
    refreshPrices.mutate(undefined, {
      onSuccess: () => toast('Prices refreshed', 'success'),
      onError: (e) => toast(e.message, 'error'),
    });
  };

  const stockHoldings = holdings.filter(h => h.asset_type === 'stock' || h.asset_type === 'etf');
  const cryptoHoldings = holdings.filter(h => h.asset_type === 'crypto');
  const otherHoldings = holdings.filter(h => h.asset_type === 'mutual_fund');

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Investments</h1>
          <p className="text-muted-foreground">Track your stock, ETF, and crypto portfolio</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleRefresh} loading={refreshPrices.isPending}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh Prices
          </Button>
          <Button onClick={() => setShowAdd(true)}>
            <Plus className="mr-2 h-4 w-4" /> Add Holding
          </Button>
        </div>
      </div>

      {/* Portfolio Summary */}
      {summaryLoading ? (
        <div className="grid gap-4 md:grid-cols-4">
          {[...Array(4)].map((_, i) => <CardSkeleton key={i} />)}
        </div>
      ) : summary ? (
        <div className="grid gap-4 md:grid-cols-4">
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Portfolio Value</p>
                  <p className="text-2xl font-bold">{formatCurrency(summary.total_value)}</p>
                </div>
                <div className="rounded-full bg-primary/10 p-3">
                  <Briefcase className="h-5 w-5 text-primary" />
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Total Cost</p>
                  <p className="text-2xl font-bold">{formatCurrency(summary.total_cost)}</p>
                </div>
                <div className="rounded-full bg-muted p-3">
                  <DollarSign className="h-5 w-5 text-muted-foreground" />
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Total Return</p>
                  <p className={`text-2xl font-bold ${summary.total_return >= 0 ? 'text-success' : 'text-destructive'}`}>
                    {summary.total_return >= 0 ? '+' : ''}{formatCurrency(summary.total_return)}
                  </p>
                  <p className={`text-xs ${summary.total_return_pct >= 0 ? 'text-success' : 'text-destructive'}`}>
                    {summary.total_return_pct >= 0 ? '+' : ''}{summary.total_return_pct.toFixed(2)}%
                  </p>
                </div>
                <div className={`rounded-full p-3 ${summary.total_return >= 0 ? 'bg-success/10' : 'bg-destructive/10'}`}>
                  {summary.total_return >= 0 ? (
                    <TrendingUp className="h-5 w-5 text-success" />
                  ) : (
                    <TrendingDown className="h-5 w-5 text-destructive" />
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Day Change</p>
                  <p className={`text-2xl font-bold ${summary.day_change >= 0 ? 'text-success' : 'text-destructive'}`}>
                    {summary.day_change >= 0 ? '+' : ''}{formatCurrency(summary.day_change)}
                  </p>
                  <p className={`text-xs ${summary.day_change_pct >= 0 ? 'text-success' : 'text-destructive'}`}>
                    {summary.day_change_pct >= 0 ? '+' : ''}{summary.day_change_pct.toFixed(2)}%
                  </p>
                </div>
                <div className={`rounded-full p-3 ${summary.day_change >= 0 ? 'bg-success/10' : 'bg-destructive/10'}`}>
                  <BarChart3 className="h-5 w-5 text-muted-foreground" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      ) : null}

      {/* Holdings */}
      {isLoading ? (
        <ListSkeleton rows={5} />
      ) : holdings.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <Briefcase className="h-12 w-12 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-semibold mb-1">No holdings yet</h3>
            <p className="text-muted-foreground mb-4">Add your first stock, ETF, or crypto holding</p>
            <Button onClick={() => setShowAdd(true)}>
              <Plus className="mr-2 h-4 w-4" /> Add Holding
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-6">
          {stockHoldings.length > 0 && (
            <HoldingsTable
              title="Stocks & ETFs"
              holdings={stockHoldings}
              onEdit={setEditing}
              onDelete={handleDelete}
              onSell={setShowSell}
            />
          )}
          {cryptoHoldings.length > 0 && (
            <HoldingsTable
              title="Crypto"
              holdings={cryptoHoldings}
              onEdit={setEditing}
              onDelete={handleDelete}
              onSell={setShowSell}
            />
          )}
          {otherHoldings.length > 0 && (
            <HoldingsTable
              title="Mutual Funds"
              holdings={otherHoldings}
              onEdit={setEditing}
              onDelete={handleDelete}
              onSell={setShowSell}
            />
          )}
        </div>
      )}

      <DividendsSection holdings={holdings} onDeleteDividend={setDeletingDividend} />

      {/* Add Holding Dialog */}
      <Dialog open={showAdd} onClose={() => setShowAdd(false)}>
        <DialogHeader>
          <DialogTitle>Add Holding</DialogTitle>
        </DialogHeader>
        <HoldingForm
          accounts={accounts}
          onSubmit={(data) => {
            createHolding.mutate(data, {
              onSuccess: () => {
                setShowAdd(false);
                toast(`${data.symbol} added`, 'success');
              },
              onError: (e) => toast(e.message, 'error'),
            });
          }}
          isLoading={createHolding.isPending}
        />
      </Dialog>

      {/* Edit Holding Dialog */}
      <Dialog open={!!editing} onClose={() => setEditing(null)}>
        <DialogHeader>
          <DialogTitle>Edit {editing?.symbol}</DialogTitle>
        </DialogHeader>
        {editing && (
          <EditHoldingForm
            holding={editing}
            onSubmit={(data) => {
              updateHolding.mutate({ id: editing.id, ...data }, {
                onSuccess: () => {
                  setEditing(null);
                  toast(`${editing.symbol} updated`, 'success');
                },
                onError: (e) => toast(e.message, 'error'),
              });
            }}
            isLoading={updateHolding.isPending}
          />
        )}
      </Dialog>

      {/* Sell Dialog */}
      <Dialog open={!!showSell} onClose={() => setShowSell(null)}>
        <DialogHeader>
          <DialogTitle>Sell {showSell?.symbol}</DialogTitle>
        </DialogHeader>
        {showSell && (
          <SellForm
            holding={showSell}
            onSubmit={(data) => {
              sellHolding.mutate({ id: showSell.id, ...data }, {
                onSuccess: () => {
                  setShowSell(null);
                  toast(`Sold ${data.quantity} ${showSell.symbol}`, 'success');
                },
                onError: (e) => toast(e.message, 'error'),
              });
            }}
            isLoading={sellHolding.isPending}
          />
        )}
      </Dialog>

      <ConfirmDialog
        open={!!deletingHolding}
        onClose={() => setDeletingHolding(null)}
        onConfirm={handleDeleteConfirm}
        title="Delete Holding"
        message={`Delete holding ${deletingHolding?.symbol ?? ''}? This action cannot be undone.`}
        isPending={deleteHolding.isPending}
      />

      <ConfirmDialog
        open={!!deletingDividend}
        onClose={() => setDeletingDividend(null)}
        onConfirm={() => {
          if (!deletingDividend) return;
          deleteDividend.mutate(deletingDividend.id, {
            onSuccess: () => { toast('Dividend removed', 'success'); setDeletingDividend(null); },
            onError: (err) => toast(err.message, 'error'),
          });
        }}
        title="Delete Dividend"
        message={`Delete dividend for ${deletingDividend?.symbol ?? ''}? This action cannot be undone.`}
        isPending={deleteDividend.isPending}
      />
    </div>
  );
}

function HoldingsTable({
  title,
  holdings,
  onEdit,
  onDelete,
  onSell,
}: {
  title: string;
  holdings: Holding[];
  onEdit: (h: Holding) => void;
  onDelete: (id: string, symbol: string) => void;
  onSell: (h: Holding) => void;
}) {
  const totalValue = holdings.reduce((s, h) => s + h.market_value, 0);

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>{title}</CardTitle>
            <CardDescription>{holdings.length} holdings &middot; {formatCurrency(totalValue)}</CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b text-left text-sm text-muted-foreground">
                <th className="pb-3 pr-4 font-medium">Asset</th>
                <th className="pb-3 pr-4 font-medium text-right">Price</th>
                <th className="pb-3 pr-4 font-medium text-right">Qty</th>
                <th className="pb-3 pr-4 font-medium text-right">Cost Basis</th>
                <th className="pb-3 pr-4 font-medium text-right">Market Value</th>
                <th className="pb-3 pr-4 font-medium text-right">Return</th>
                <th className="pb-3 pr-4 font-medium text-right">Day</th>
                <th className="pb-3 font-medium text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {holdings.map((h) => (
                <tr key={h.id} className="border-b last:border-0 hover:bg-muted/50 transition-colors">
                  <td className="py-3 pr-4">
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 font-bold text-primary text-sm">
                        {h.symbol.slice(0, 3)}
                      </div>
                      <div>
                        <p className="font-medium">{h.symbol}</p>
                        <p className="text-sm text-muted-foreground truncate max-w-[200px]">{h.asset_name || h.name}</p>
                      </div>
                    </div>
                  </td>
                  <td className="py-3 pr-4 text-right">
                    {h.current_price > 0 ? formatCurrency(h.current_price) : '—'}
                  </td>
                  <td className="py-3 pr-4 text-right font-mono">
                    {h.quantity.toLocaleString(undefined, { maximumFractionDigits: 4 })}
                  </td>
                  <td className="py-3 pr-4 text-right">{formatCurrency(h.cost_basis)}</td>
                  <td className="py-3 pr-4 text-right font-semibold">
                    {h.market_value > 0 ? formatCurrency(h.market_value) : '—'}
                  </td>
                  <td className="py-3 pr-4 text-right">
                    <div className={`flex items-center justify-end gap-1 ${h.total_return >= 0 ? 'text-success' : 'text-destructive'}`}>
                      {h.total_return >= 0 ? <ArrowUpRight className="h-3 w-3" /> : <ArrowDownRight className="h-3 w-3" />}
                      <span className="font-medium">
                        {h.total_return >= 0 ? '+' : ''}{formatCurrency(h.total_return)}
                      </span>
                    </div>
                    <p className={`text-xs ${h.return_percent >= 0 ? 'text-success' : 'text-destructive'}`}>
                      {h.return_percent >= 0 ? '+' : ''}{h.return_percent.toFixed(2)}%
                    </p>
                  </td>
                  <td className="py-3 pr-4 text-right">
                    <Badge variant={h.day_change >= 0 ? 'success' : 'destructive'}>
                      {h.day_change >= 0 ? '+' : ''}{h.day_change.toFixed(2)}%
                    </Badge>
                  </td>
                  <td className="py-3 text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button variant="ghost" size="sm" onClick={() => onSell(h)} title="Sell">
                        <DollarSign className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => onEdit(h)} title="Edit">
                        <BarChart3 className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => onDelete(h.id, h.symbol)} title="Delete">
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  );
}

function HoldingForm({
  accounts,
  onSubmit,
  isLoading,
}: {
  accounts: { id: string; name: string }[];
  onSubmit: (data: {
    account_id: string;
    asset_type: AssetType;
    symbol: string;
    name: string;
    quantity: number;
    cost_basis: number;
    currency: string;
    acquired_at: string;
    notes: string;
  }) => void;
  isLoading: boolean;
}) {
  const [accountId, setAccountId] = useState(accounts[0]?.id ?? '');
  const [assetType, setAssetType] = useState<AssetType>('stock');
  const [symbol, setSymbol] = useState('');
  const [name, setName] = useState('');
  const [quantity, setQuantity] = useState('');
  const [costBasis, setCostBasis] = useState('');
  const [currency, setCurrency] = useState('USD');
  const [acquiredAt, setAcquiredAt] = useState(new Date().toISOString().split('T')[0]);
  const [notes, setNotes] = useState('');

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    onSubmit({
      account_id: accountId,
      asset_type: assetType,
      symbol: symbol.toUpperCase(),
      name,
      quantity: parseFloat(quantity),
      cost_basis: parseFloat(costBasis),
      currency,
      acquired_at: acquiredAt,
      notes,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Account</label>
          <Select value={accountId} onChange={(e) => setAccountId(e.target.value)}>
            {accounts.length === 0 && <option value="">No investment accounts</option>}
            {accounts.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}
          </Select>
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Asset Type</label>
          <Select value={assetType} onChange={(e) => setAssetType(e.target.value as AssetType)}>
            <option value="stock">Stock</option>
            <option value="etf">ETF</option>
            <option value="crypto">Crypto</option>
            <option value="mutual_fund">Mutual Fund</option>
          </Select>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Symbol / Ticker</label>
          <Input
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            placeholder="e.g. AAPL, BTC-USD"
            required
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Name (optional)</label>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Auto-fetched if empty"
          />
        </div>
      </div>
      <div className="grid grid-cols-3 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Quantity</label>
          <Input
            type="number"
            step="any"
            min="0"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            placeholder="10"
            required
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Total Cost Basis</label>
          <Input
            type="number"
            step="0.01"
            min="0"
            value={costBasis}
            onChange={(e) => setCostBasis(e.target.value)}
            placeholder="1500.00"
            required
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Currency</label>
          <Select value={currency} onChange={(e) => setCurrency(e.target.value)}>
            <option value="USD">USD</option>
            <option value="EUR">EUR</option>
            <option value="GBP">GBP</option>
            <option value="ISK">ISK</option>
          </Select>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Acquired Date</label>
          <Input
            type="date"
            value={acquiredAt}
            onChange={(e) => setAcquiredAt(e.target.value)}
            required
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Notes</label>
          <Input
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Optional notes"
          />
        </div>
      </div>
      <div className="flex justify-end gap-2">
        <Button type="submit" disabled={isLoading || !accountId}>
          {isLoading ? 'Adding...' : 'Add Holding'}
        </Button>
      </div>
    </form>
  );
}

function EditHoldingForm({
  holding,
  onSubmit,
  isLoading,
}: {
  holding: Holding;
  onSubmit: (data: { quantity?: number; cost_basis?: number; notes?: string }) => void;
  isLoading: boolean;
}) {
  const [quantity, setQuantity] = useState(String(holding.quantity));
  const [costBasis, setCostBasis] = useState(String(holding.cost_basis));
  const [notes, setNotes] = useState(holding.notes);

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    onSubmit({
      quantity: parseFloat(quantity),
      cost_basis: parseFloat(costBasis),
      notes,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Quantity</label>
          <Input
            type="number"
            step="any"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            required
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Cost Basis</label>
          <Input
            type="number"
            step="0.01"
            value={costBasis}
            onChange={(e) => setCostBasis(e.target.value)}
            required
          />
        </div>
      </div>
      <div className="space-y-2">
        <label className="text-sm font-medium">Notes</label>
        <Input
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          placeholder="Optional notes"
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="submit" disabled={isLoading}>
          {isLoading ? 'Saving...' : 'Save Changes'}
        </Button>
      </div>
    </form>
  );
}

function SellForm({
  holding,
  onSubmit,
  isLoading,
}: {
  holding: Holding;
  onSubmit: (data: { quantity: number; price: number; sold_at: string; method: CostBasisMethod }) => void;
  isLoading: boolean;
}) {
  const [quantity, setQuantity] = useState('');
  const [price, setPrice] = useState(String(holding.current_price || ''));
  const [soldAt, setSoldAt] = useState(new Date().toISOString().split('T')[0]);
  const [method, setMethod] = useState<CostBasisMethod>('fifo');

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    onSubmit({
      quantity: parseFloat(quantity),
      price: parseFloat(price),
      sold_at: soldAt,
      method,
    });
  };

  const sellValue = parseFloat(quantity || '0') * parseFloat(price || '0');
  const costBasisPortion = holding.quantity > 0
    ? (holding.cost_basis / holding.quantity) * parseFloat(quantity || '0')
    : 0;
  const gain = sellValue - costBasisPortion;

  const methodDescriptions: Record<CostBasisMethod, string> = {
    fifo: 'Sells oldest lots first (First In, First Out)',
    lifo: 'Sells newest lots first (Last In, First Out)',
    average: 'Uses average cost per unit across all lots',
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="rounded-lg bg-muted/50 p-3 text-sm">
        <p>Current holding: <span className="font-medium">{holding.quantity}</span> {holding.symbol}</p>
        <p>Avg cost: <span className="font-medium">{formatCurrency(holding.quantity > 0 ? holding.cost_basis / holding.quantity : 0)}</span>/share</p>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Quantity to Sell</label>
          <Input
            type="number"
            step="any"
            min="0"
            max={holding.quantity}
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            placeholder={`Max ${holding.quantity}`}
            required
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Price per Unit</label>
          <Input
            type="number"
            step="0.01"
            min="0"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            required
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Sell Date</label>
          <Input
            type="date"
            value={soldAt}
            onChange={(e) => setSoldAt(e.target.value)}
            required
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Cost Basis Method</label>
          <Select value={method} onChange={(e) => setMethod(e.target.value as CostBasisMethod)}>
            <option value="fifo">FIFO</option>
            <option value="lifo">LIFO</option>
            <option value="average">Average Cost</option>
          </Select>
        </div>
      </div>
      <p className="text-xs text-muted-foreground">{methodDescriptions[method]}</p>
      {parseFloat(quantity) > 0 && parseFloat(price) > 0 && (
        <div className="rounded-lg border p-3 text-sm space-y-1">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Sell Value</span>
            <span className="font-medium">{formatCurrency(sellValue)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Cost Basis ({method.toUpperCase()})</span>
            <span className="font-medium">{formatCurrency(costBasisPortion)}</span>
          </div>
          <div className="flex justify-between border-t pt-1">
            <span className="text-muted-foreground">Estimated Gain/Loss</span>
            <span className={`font-bold ${gain >= 0 ? 'text-success' : 'text-destructive'}`}>
              {gain >= 0 ? '+' : ''}{formatCurrency(gain)}
            </span>
          </div>
        </div>
      )}
      <div className="flex justify-end gap-2">
        <Button type="submit" disabled={isLoading || !quantity}>
          {isLoading ? 'Selling...' : 'Record Sale'}
        </Button>
      </div>
    </form>
  );
}

function DividendsSection({ holdings, onDeleteDividend }: { holdings: Holding[]; onDeleteDividend: (info: { id: string; symbol: string }) => void }) {
  const { data: divData } = useDividends();
  const { data: summary } = useDividendSummary();
  const createDividend = useCreateDividend();
  const { toast } = useToast();

  const [showForm, setShowForm] = useState(false);
  const [holdingId, setHoldingId] = useState('');
  const [amount, setAmount] = useState('');
  const [exDate, setExDate] = useState('');
  const [payDate, setPayDate] = useState('');
  const [notes, setNotes] = useState('');

  const dividends = divData?.dividends ?? [];

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const req: CreateDividendRequest = {
      holding_id: holdingId,
      amount,
      ex_date: exDate,
      pay_date: payDate || undefined,
      notes: notes || undefined,
    };
    createDividend.mutate(req, {
      onSuccess: () => {
        setShowForm(false);
        setAmount(''); setExDate(''); setPayDate(''); setNotes('');
        toast('Dividend recorded', 'success');
      },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  const holdingMap = Object.fromEntries(holdings.map(h => [h.id, h]));

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Coins className="h-5 w-5" />
              Dividends
            </CardTitle>
            <CardDescription>Track dividend income from your holdings</CardDescription>
          </div>
          <Button size="sm" onClick={() => setShowForm(true)} disabled={holdings.length === 0}>
            <Plus className="mr-1 h-4 w-4" /> Record Dividend
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {summary && (
          <div className="grid grid-cols-3 gap-4">
            <div className="rounded-lg border p-3 text-center">
              <p className="text-xs text-muted-foreground">Total Dividends</p>
              <p className="text-lg font-bold text-success">{formatCurrency(summary.total_dividends)}</p>
            </div>
            <div className="rounded-lg border p-3 text-center">
              <p className="text-xs text-muted-foreground">YTD</p>
              <p className="text-lg font-bold">{formatCurrency(summary.dividends_ytd)}</p>
            </div>
            <div className="rounded-lg border p-3 text-center">
              <p className="text-xs text-muted-foreground">Payments</p>
              <p className="text-lg font-bold">{summary.dividend_count}</p>
            </div>
          </div>
        )}

        {dividends.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-4">No dividends recorded yet.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-2 font-medium">Holding</th>
                  <th className="pb-2 font-medium">Amount</th>
                  <th className="pb-2 font-medium">Ex-Date</th>
                  <th className="pb-2 font-medium">Pay Date</th>
                  <th className="pb-2 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {dividends.slice(0, 10).map((d) => {
                  const h = holdingMap[d.holding_id];
                  return (
                    <tr key={d.id} className="border-b last:border-0">
                      <td className="py-2 font-medium">{h?.symbol ?? '—'}</td>
                      <td className="py-2 text-success">{formatCurrency(parseFloat(d.amount), d.currency)}</td>
                      <td className="py-2">{formatDate(d.ex_date)}</td>
                      <td className="py-2">{d.pay_date ? formatDate(d.pay_date) : '—'}</td>
                      <td className="py-2">
                        <Button variant="ghost" size="sm" onClick={() =>
                          onDeleteDividend({ id: d.id, symbol: h?.symbol ?? 'Unknown' })
                        }>
                          <Trash2 className="h-3 w-3 text-destructive" />
                        </Button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        <Dialog open={showForm} onClose={() => setShowForm(false)}>
          <DialogHeader>
            <DialogTitle>Record Dividend</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-3">
            <div>
              <label className="text-sm font-medium">Holding</label>
              <Select value={holdingId} onChange={(e) => setHoldingId(e.target.value)} required>
                <option value="">Select holding...</option>
                {holdings.map(h => (
                  <option key={h.id} value={h.id}>{h.symbol} — {h.asset_name || h.name}</option>
                ))}
              </Select>
            </div>
            <div>
              <label className="text-sm font-medium">Amount</label>
              <Input type="number" step="0.01" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="0.00" required />
            </div>
            <div className="grid grid-cols-2 gap-2">
              <div>
                <label className="text-sm font-medium">Ex-Date</label>
                <Input type="date" value={exDate} onChange={(e) => setExDate(e.target.value)} required />
              </div>
              <div>
                <label className="text-sm font-medium">Pay Date (optional)</label>
                <Input type="date" value={payDate} onChange={(e) => setPayDate(e.target.value)} />
              </div>
            </div>
            <div>
              <label className="text-sm font-medium">Notes (optional)</label>
              <Input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="e.g. Quarterly dividend" />
            </div>
            <Button type="submit" className="w-full" loading={createDividend.isPending}>
              Record Dividend
            </Button>
          </form>
        </Dialog>
      </CardContent>
    </Card>
  );
}
