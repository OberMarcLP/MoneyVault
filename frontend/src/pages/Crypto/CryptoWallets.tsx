import { useState, type FormEvent } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/ConfirmDialog';
import { Badge } from '@/components/ui/badge';
import { CardSkeleton } from '@/components/ui/skeleton';
import {
  useWallets, useCreateWallet, useDeleteWallet, useSyncWallet,
  useWalletTransactions,
} from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { formatDate } from '@/lib/utils';
import {
  Plus, Trash2, RefreshCw, Wallet, ExternalLink,
  Fuel, Send, ArrowDown,
} from 'lucide-react';
import type { Wallet as WalletType, WalletTransaction } from '@/types';

function getExplorerURL(network: string, address: string, type: 'address' | 'tx' = 'address') {
  const explorers: Record<string, string> = {
    ethereum: 'https://etherscan.io',
    polygon: 'https://polygonscan.com',
    bsc: 'https://bscscan.com',
    arbitrum: 'https://arbiscan.io',
  };
  const base = explorers[network] || explorers.ethereum;
  return `${base}/${type}/${address}`;
}

function getExplorerName(network: string) {
  const names: Record<string, string> = {
    ethereum: 'Etherscan',
    polygon: 'Polygonscan',
    bsc: 'BscScan',
    arbitrum: 'Arbiscan',
  };
  return names[network] || 'Explorer';
}

function WalletsTab({ wallets, isLoading, onAdd, onDelete, onSync, syncPending, selectedWallet, onSelectWallet }: {
  wallets: WalletType[]; isLoading: boolean;
  onAdd: () => void; onDelete: (id: string) => void;
  onSync: (id: string) => void; syncPending: boolean;
  selectedWallet: string | null; onSelectWallet: (id: string | null) => void;
}) {
  if (isLoading) return <CardSkeleton />;

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button size="sm" onClick={onAdd}><Plus className="h-4 w-4 mr-2" />Track Wallet</Button>
      </div>

      {wallets.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Wallet className="h-12 w-12 mx-auto text-muted-foreground mb-3" />
            <p className="text-muted-foreground">No wallets being tracked</p>
            <p className="text-xs text-muted-foreground mt-1">Add an Ethereum wallet to track transactions and gas fees</p>
            <Button variant="outline" size="sm" className="mt-3" onClick={onAdd}>Add wallet</Button>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {wallets.map(w => (
              <Card key={w.id} className={`cursor-pointer transition-colors ${selectedWallet === w.id ? 'ring-2 ring-primary' : ''}`}
                onClick={() => onSelectWallet(selectedWallet === w.id ? null : w.id)}>
                <CardContent className="p-4">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center">
                        <Wallet className="h-4 w-4 text-primary" />
                      </div>
                      <div>
                        <p className="font-medium text-sm">{w.label || 'Unnamed'}</p>
                        <Badge variant="secondary" className="text-[10px]">{w.network}</Badge>
                      </div>
                    </div>
                    <div className="flex gap-1">
                      <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); onSync(w.id); }} disabled={syncPending}>
                        <RefreshCw className={`h-3.5 w-3.5 ${syncPending ? 'animate-spin' : ''}`} />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); onDelete(w.id); }} className="text-destructive">
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </div>
                  <p className="font-mono text-xs text-muted-foreground truncate">{w.address}</p>
                  {w.last_synced && (
                    <p className="text-[10px] text-muted-foreground mt-2">
                      Last synced: {new Date(w.last_synced).toLocaleString()}
                    </p>
                  )}
                  <a
                    href={getExplorerURL(w.network, w.address)}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs text-primary flex items-center gap-1 mt-2 hover:underline"
                    onClick={(e) => e.stopPropagation()}
                  >
                    View on {getExplorerName(w.network)} <ExternalLink className="h-3 w-3" />
                  </a>
                </CardContent>
              </Card>
            ))}
          </div>

          {selectedWallet && <WalletTransactionsList walletId={selectedWallet} />}
        </div>
      )}
    </div>
  );
}

function WalletTransactionsList({ walletId }: { walletId: string }) {
  const { data, isLoading } = useWalletTransactions(walletId);
  const txs = data ?? [];

  if (isLoading) return <CardSkeleton />;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Recent Transactions</CardTitle>
      </CardHeader>
      <CardContent>
        {txs.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-4">No transactions found. Try syncing the wallet.</p>
        ) : (
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {txs.map(tx => (
              <WalletTxRow key={tx.id} tx={tx} />
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function WalletTxRow({ tx }: { tx: WalletTransaction }) {
  const isSend = tx.tx_type === 'send';
  const ethValue = parseFloat(tx.value) / 1e18;

  return (
    <div className="flex items-center justify-between p-3 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors">
      <div className="flex items-center gap-3">
        <div className={`h-8 w-8 rounded-full flex items-center justify-center ${isSend ? 'bg-destructive/10' : 'bg-success/10'}`}>
          {isSend ? <Send className="h-4 w-4 text-destructive" /> : <ArrowDown className="h-4 w-4 text-success" />}
        </div>
        <div>
          <p className="text-sm font-medium">{isSend ? 'Sent' : 'Received'} {tx.token_symbol}</p>
          <p className="font-mono text-[10px] text-muted-foreground">
            {tx.tx_hash.slice(0, 10)}...{tx.tx_hash.slice(-8)}
          </p>
        </div>
      </div>
      <div className="text-right">
        <p className={`text-sm font-medium ${isSend ? 'text-destructive' : 'text-success'}`}>
          {isSend ? '-' : '+'}{ethValue.toFixed(6)} {tx.token_symbol}
        </p>
        <div className="flex items-center gap-1 text-[10px] text-muted-foreground justify-end">
          <Fuel className="h-3 w-3" />
          {tx.gas_fee_eth.toFixed(6)} ETH
        </div>
        <p className="text-[10px] text-muted-foreground">{formatDate(tx.timestamp)}</p>
      </div>
    </div>
  );
}

function AddWalletForm({ onSubmit, isPending }: {
  onSubmit: (data: { address: string; network?: string; label?: string }) => void;
  isPending: boolean;
}) {
  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    onSubmit({
      address: fd.get('address') as string,
      network: (fd.get('network') as string) || 'ethereum',
      label: (fd.get('label') as string) || '',
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 p-4">
      <Input name="address" placeholder="0x... wallet address" required />
      <Select name="network">
        <option value="ethereum">Ethereum</option>
        <option value="polygon">Polygon</option>
        <option value="bsc">BSC</option>
        <option value="arbitrum">Arbitrum</option>
      </Select>
      <Input name="label" placeholder="Label (e.g. My Main Wallet)" />
      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Adding...' : 'Track Wallet'}
      </Button>
    </form>
  );
}

export default function CryptoWallets() {
  const { data: walletData, isLoading: walletsLoading } = useWallets();
  const createWallet = useCreateWallet();
  const deleteWallet = useDeleteWallet();
  const syncWallet = useSyncWallet();
  const { toast } = useToast();

  const [showAddWallet, setShowAddWallet] = useState(false);
  const [selectedWallet, setSelectedWallet] = useState<string | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<{ id: string } | null>(null);

  const wallets = walletData ?? [];

  return (
    <>
      <WalletsTab
        wallets={wallets}
        isLoading={walletsLoading}
        onAdd={() => setShowAddWallet(true)}
        onDelete={(id) => setDeleteConfirm({ id })}
        onSync={(id) => {
          syncWallet.mutate(id, {
            onSuccess: (data) => toast(`Synced ${data.synced} transactions`, 'success'),
            onError: (e) => toast(e.message, 'error'),
          });
        }}
        syncPending={syncWallet.isPending}
        selectedWallet={selectedWallet}
        onSelectWallet={setSelectedWallet}
      />

      {showAddWallet && (
        <Dialog open onClose={() => setShowAddWallet(false)}>
          <DialogHeader><DialogTitle>Track Wallet</DialogTitle></DialogHeader>
          <AddWalletForm
            onSubmit={(data) => {
              createWallet.mutate(data, {
                onSuccess: () => { toast('Wallet added', 'success'); setShowAddWallet(false); },
                onError: (e) => toast(e.message, 'error'),
              });
            }}
            isPending={createWallet.isPending}
          />
        </Dialog>
      )}

      <ConfirmDialog
        open={!!deleteConfirm}
        onClose={() => setDeleteConfirm(null)}
        onConfirm={() => {
          if (!deleteConfirm) return;
          deleteWallet.mutate(deleteConfirm.id, {
            onSuccess: () => { toast('Wallet deleted', 'success'); setDeleteConfirm(null); },
            onError: (e) => toast(e.message, 'error'),
          });
        }}
        title="Delete Wallet"
        message="Delete this wallet? This action cannot be undone."
        isPending={deleteWallet.isPending}
      />
    </>
  );
}
