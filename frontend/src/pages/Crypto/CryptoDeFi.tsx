import { useState, type FormEvent } from 'react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/ConfirmDialog';
import { Badge } from '@/components/ui/badge';
import { CardSkeleton } from '@/components/ui/skeleton';
import {
  useHoldings, useCreateHolding, useDeleteHolding, useAccounts,
} from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { formatCurrency } from '@/lib/utils';
import { Plus, Trash2, Layers } from 'lucide-react';
import type { Holding, CreateHoldingRequest } from '@/types';

function DeFiTab({ holdings, isLoading, onAdd, onDelete }: {
  holdings: Holding[]; isLoading: boolean;
  onAdd: () => void; onDelete: (id: string, name: string) => void;
}) {
  if (isLoading) return <CardSkeleton />;

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button size="sm" onClick={onAdd}><Plus className="h-4 w-4 mr-2" />Add Position</Button>
      </div>

      {holdings.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Layers className="h-12 w-12 mx-auto text-muted-foreground mb-3" />
            <p className="text-muted-foreground">No DeFi positions tracked yet</p>
            <p className="text-xs text-muted-foreground mt-1">Track LP positions, staking, and yield farming</p>
            <Button variant="outline" size="sm" className="mt-3" onClick={onAdd}>Add a position</Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {holdings.map(h => (
            <Card key={h.id}>
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary">{h.symbol}</Badge>
                    <span className="text-sm font-medium">{h.name || h.asset_name}</span>
                  </div>
                  <Button variant="ghost" size="sm" onClick={() => onDelete(h.id, h.name || h.symbol)} className="text-destructive">
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Value</span>
                  <span className="font-medium">{formatCurrency(h.market_value || h.cost_basis)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Quantity</span>
                  <span>{h.quantity.toLocaleString(undefined, { maximumFractionDigits: 8 })}</span>
                </div>
                {h.notes && (
                  <p className="text-xs text-muted-foreground border-t pt-2 mt-2">{h.notes}</p>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

function AddDeFiForm({ accounts, onSubmit, isPending }: {
  accounts: { id: string; name: string }[];
  onSubmit: (data: CreateHoldingRequest) => void;
  isPending: boolean;
}) {
  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const metadata = JSON.stringify({
      protocol: fd.get('protocol') || '',
      pool_name: fd.get('pool_name') || '',
      position_type: fd.get('position_type') || 'liquidity',
      apy: parseFloat(fd.get('apy') as string) || 0,
    });
    onSubmit({
      account_id: fd.get('account_id') as string,
      asset_type: 'defi_position',
      symbol: (fd.get('symbol') as string).toUpperCase(),
      name: (fd.get('name') as string) || '',
      quantity: parseFloat(fd.get('quantity') as string),
      cost_basis: parseFloat(fd.get('cost_basis') as string),
      acquired_at: fd.get('acquired_at') as string,
      metadata,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 p-4">
      <Select name="account_id" required>
        <option value="">Select account</option>
        {accounts.map(a => <option key={a.id} value={a.id}>{a.name}</option>)}
      </Select>
      <Select name="position_type">
        <option value="liquidity">Liquidity Pool</option>
        <option value="staking">Staking</option>
        <option value="farming">Yield Farming</option>
        <option value="lending">Lending</option>
      </Select>
      <Input name="protocol" placeholder="Protocol (e.g. Uniswap, Aave)" />
      <Input name="pool_name" placeholder="Pool name (e.g. ETH/USDC)" />
      <Input name="symbol" placeholder="Token symbol" required />
      <Input name="name" placeholder="Position name" />
      <div className="grid grid-cols-2 gap-3">
        <Input name="quantity" type="number" step="any" placeholder="Quantity" required />
        <Input name="cost_basis" type="number" step="any" placeholder="Value (USD)" required />
      </div>
      <Input name="apy" type="number" step="any" placeholder="APY %" />
      <Input name="acquired_at" type="date" required />
      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Adding...' : 'Add Position'}
      </Button>
    </form>
  );
}

export default function CryptoDeFi() {
  const { data: holdingsData, isLoading } = useHoldings();
  const { data: accountData } = useAccounts();
  const createHolding = useCreateHolding();
  const deleteHolding = useDeleteHolding();
  const { toast } = useToast();

  const [showAddDefi, setShowAddDefi] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState<{ id: string; label: string } | null>(null);

  const defiHoldings = (holdingsData ?? []).filter(h => h.asset_type === 'defi_position');
  const accounts = (accountData?.accounts ?? []).filter(a => a.type === 'crypto_wallet' || a.type === 'investment');

  return (
    <>
      <DeFiTab
        holdings={defiHoldings}
        isLoading={isLoading}
        onAdd={() => setShowAddDefi(true)}
        onDelete={(id, name) => setDeleteConfirm({ id, label: name })}
      />

      {showAddDefi && (
        <Dialog open onClose={() => setShowAddDefi(false)}>
          <DialogHeader><DialogTitle>Add DeFi Position</DialogTitle></DialogHeader>
          <AddDeFiForm
            accounts={accounts}
            onSubmit={(data) => {
              createHolding.mutate(data, {
                onSuccess: () => { toast('DeFi position added', 'success'); setShowAddDefi(false); },
                onError: (e) => toast(e.message, 'error'),
              });
            }}
            isPending={createHolding.isPending}
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
