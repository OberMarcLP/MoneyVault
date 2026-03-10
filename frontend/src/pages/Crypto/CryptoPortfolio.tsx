import { useState, type FormEvent } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/ConfirmDialog';
import { CardSkeleton } from '@/components/ui/skeleton';
import {
  useHoldings, useCreateHolding, useDeleteHolding, useSellHolding,
  useAccounts, useSearchTokens,
} from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { formatCurrency } from '@/lib/utils';
import {
  Plus, Trash2, Bitcoin, Search, Send,
  ArrowUpRight, ArrowDownRight,
} from 'lucide-react';
import type { Holding } from '@/types';

function PortfolioTab({ holdings, isLoading, onAdd, onSell, onDelete }: {
  holdings: Holding[]; isLoading: boolean;
  onAdd: () => void; onSell: (h: Holding) => void;
  onDelete: (id: string, symbol: string) => void;
}) {
  if (isLoading) return <CardSkeleton />;

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button size="sm" onClick={onAdd}><Plus className="h-4 w-4 mr-2" />Add Token</Button>
      </div>

      {holdings.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Bitcoin className="h-12 w-12 mx-auto text-muted-foreground mb-3" />
            <p className="text-muted-foreground">No crypto holdings yet</p>
            <Button variant="outline" size="sm" className="mt-3" onClick={onAdd}>Add your first token</Button>
          </CardContent>
        </Card>
      ) : (
        <div className="rounded-lg border overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted/50">
              <tr className="text-xs text-muted-foreground">
                <th className="text-left p-3 font-medium">Token</th>
                <th className="text-right p-3 font-medium">Price</th>
                <th className="text-right p-3 font-medium">Holdings</th>
                <th className="text-right p-3 font-medium">Value</th>
                <th className="text-right p-3 font-medium">Return</th>
                <th className="text-right p-3 font-medium">24h</th>
                <th className="text-right p-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {holdings.map(h => (
                <tr key={h.id} className="hover:bg-muted/30 transition-colors">
                  <td className="p-3">
                    <div className="flex items-center gap-2">
                      <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-xs font-bold text-primary">
                        {h.symbol.slice(0, 3)}
                      </div>
                      <div>
                        <p className="font-medium text-sm">{h.symbol}</p>
                        <p className="text-xs text-muted-foreground">{h.asset_name || h.name}</p>
                      </div>
                    </div>
                  </td>
                  <td className="p-3 text-right text-sm">{formatCurrency(h.current_price)}</td>
                  <td className="p-3 text-right text-sm">{h.quantity.toLocaleString(undefined, { maximumFractionDigits: 8 })}</td>
                  <td className="p-3 text-right text-sm font-medium">{formatCurrency(h.market_value)}</td>
                  <td className="p-3 text-right text-sm">
                    <span className={h.total_return >= 0 ? 'text-success' : 'text-destructive'}>
                      {h.total_return >= 0 ? '+' : ''}{formatCurrency(h.total_return)}
                    </span>
                    <p className="text-xs text-muted-foreground">
                      {h.return_percent >= 0 ? '+' : ''}{h.return_percent.toFixed(2)}%
                    </p>
                  </td>
                  <td className="p-3 text-right text-sm">
                    <span className={h.day_change >= 0 ? 'text-success' : 'text-destructive'}>
                      {h.day_change >= 0 ? <ArrowUpRight className="inline h-3 w-3" /> : <ArrowDownRight className="inline h-3 w-3" />}
                      {Math.abs(h.day_change).toFixed(2)}%
                    </span>
                  </td>
                  <td className="p-3 text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button variant="ghost" size="sm" onClick={() => onSell(h)} title="Sell">
                        <Send className="h-3.5 w-3.5" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => onDelete(h.id, h.symbol)} className="text-destructive">
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function AddTokenForm({ accounts, onSubmit, isPending }: {
  accounts: { id: string; name: string }[];
  onSubmit: (data: any) => void;
  isPending: boolean;
}) {
  const [symbol, setSymbol] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const { data: searchResults } = useSearchTokens(searchQuery);

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    onSubmit({
      account_id: fd.get('account_id'),
      asset_type: 'crypto',
      symbol: (fd.get('symbol') as string).toUpperCase(),
      name: fd.get('name') || '',
      quantity: parseFloat(fd.get('quantity') as string),
      cost_basis: parseFloat(fd.get('cost_basis') as string),
      acquired_at: fd.get('acquired_at'),
      network: fd.get('network') || '',
      token_address: fd.get('token_address') || '',
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 p-4">
      <div>
        <label className="text-sm font-medium mb-1 block">Search Token</label>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by name or symbol..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        {searchResults && searchResults.length > 0 && searchQuery.length >= 2 && (
          <div className="border rounded-lg mt-1 max-h-32 overflow-y-auto bg-popover">
            {searchResults.map(t => (
              <button
                key={t.id}
                type="button"
                className="w-full text-left px-3 py-2 text-sm hover:bg-muted/50 flex justify-between"
                onClick={() => { setSymbol(t.symbol.toUpperCase()); setSearchQuery(''); }}
              >
                <span className="font-medium">{t.symbol.toUpperCase()}</span>
                <span className="text-muted-foreground text-xs">{t.name}</span>
              </button>
            ))}
          </div>
        )}
      </div>
      <Select name="account_id" required>
        <option value="">Select account</option>
        {accounts.map(a => <option key={a.id} value={a.id}>{a.name}</option>)}
      </Select>
      <Input name="symbol" placeholder="Symbol (e.g. BTC)" required value={symbol} onChange={(e) => setSymbol(e.target.value)} />
      <Input name="name" placeholder="Name (optional)" />
      <div className="grid grid-cols-2 gap-3">
        <Input name="quantity" type="number" step="any" placeholder="Quantity" required />
        <Input name="cost_basis" type="number" step="any" placeholder="Total cost" required />
      </div>
      <Input name="acquired_at" type="date" required />
      <Input name="network" placeholder="Network (e.g. ethereum)" />
      <Input name="token_address" placeholder="Token address (optional)" />
      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Adding...' : 'Add Token'}
      </Button>
    </form>
  );
}

function SellForm({ holding, onSubmit, isPending }: {
  holding: Holding;
  onSubmit: (data: { quantity: number; price: number; sold_at: string }) => void;
  isPending: boolean;
}) {
  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    onSubmit({
      quantity: parseFloat(fd.get('quantity') as string),
      price: parseFloat(fd.get('price') as string),
      sold_at: fd.get('sold_at') as string,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 p-4">
      <p className="text-sm text-muted-foreground">
        Available: {holding.quantity.toLocaleString(undefined, { maximumFractionDigits: 8 })} {holding.symbol}
      </p>
      <Input name="quantity" type="number" step="any" placeholder="Quantity to sell" required max={holding.quantity} />
      <Input name="price" type="number" step="any" placeholder="Price per unit (USD)" required defaultValue={holding.current_price || ''} />
      <Input name="sold_at" type="date" required defaultValue={new Date().toISOString().split('T')[0]} />
      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Selling...' : 'Record Sale'}
      </Button>
    </form>
  );
}

export default function CryptoPortfolio() {
  const { data: holdingsData, isLoading } = useHoldings();
  const { data: accountData } = useAccounts();
  const createHolding = useCreateHolding();
  const deleteHolding = useDeleteHolding();
  const sellHolding = useSellHolding();
  const { toast } = useToast();

  const [showAddToken, setShowAddToken] = useState(false);
  const [showSell, setShowSell] = useState<Holding | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<{ id: string; label: string } | null>(null);

  const cryptoHoldings = (holdingsData ?? []).filter(h => h.asset_type === 'crypto');
  const accounts = (accountData?.accounts ?? []).filter(a => a.type === 'crypto_wallet' || a.type === 'investment');

  return (
    <>
      <PortfolioTab
        holdings={cryptoHoldings}
        isLoading={isLoading}
        onAdd={() => setShowAddToken(true)}
        onSell={setShowSell}
        onDelete={(id, sym) => setDeleteConfirm({ id, label: sym })}
      />

      {showAddToken && (
        <Dialog open onClose={() => setShowAddToken(false)}>
          <DialogHeader><DialogTitle>Add Crypto Holding</DialogTitle></DialogHeader>
          <AddTokenForm
            accounts={accounts}
            onSubmit={(data) => {
              createHolding.mutate(data, {
                onSuccess: () => { toast('Token added', 'success'); setShowAddToken(false); },
                onError: (e) => toast(e.message, 'error'),
              });
            }}
            isPending={createHolding.isPending}
          />
        </Dialog>
      )}

      {showSell && (
        <Dialog open onClose={() => setShowSell(null)}>
          <DialogHeader><DialogTitle>Sell {showSell.symbol}</DialogTitle></DialogHeader>
          <SellForm
            holding={showSell}
            onSubmit={(data) => {
              sellHolding.mutate({ id: showSell.id, ...data }, {
                onSuccess: () => { toast('Sale recorded', 'success'); setShowSell(null); },
                onError: (e) => toast(e.message, 'error'),
              });
            }}
            isPending={sellHolding.isPending}
          />
        </Dialog>
      )}

      <ConfirmDialog
        open={!!deleteConfirm}
        onClose={() => setDeleteConfirm(null)}
        onConfirm={() => {
          if (!deleteConfirm) return;
          deleteHolding.mutate(deleteConfirm.id, {
            onSuccess: () => { toast(`${deleteConfirm.label} deleted`, 'success'); setDeleteConfirm(null); },
            onError: (e) => toast(e.message, 'error'),
          });
        }}
        title="Delete Holding"
        message={`Delete ${deleteConfirm?.label ?? ''}? This action cannot be undone.`}
        isPending={deleteHolding.isPending}
      />
    </>
  );
}
