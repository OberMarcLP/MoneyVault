import { useState, type FormEvent } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/ConfirmDialog';
import { ListSkeleton } from '@/components/ui/skeleton';
import {
  useRecurring, useCreateRecurring, useDeleteRecurring, useToggleRecurring,
  useAccounts, useCategories,
} from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { formatCurrency, formatDate } from '@/lib/utils';
import { Plus, Trash2, Repeat, Pause, Play } from 'lucide-react';
import type { Frequency, TransactionType } from '@/types';

const FREQ_LABELS: Record<Frequency, string> = {
  daily: 'Daily',
  weekly: 'Weekly',
  biweekly: 'Bi-weekly',
  monthly: 'Monthly',
  quarterly: 'Quarterly',
  yearly: 'Yearly',
};

export function RecurringPage() {
  const { data, isLoading } = useRecurring();
  const { data: accountsData } = useAccounts();
  const { data: catData } = useCategories();
  const createRecurring = useCreateRecurring();
  const deleteRecurring = useDeleteRecurring();
  const toggleRecurring = useToggleRecurring();
  const { toast } = useToast();
  const [showCreate, setShowCreate] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const items = data?.recurring ?? [];
  const accounts = accountsData?.accounts ?? [];
  const categories = catData?.categories ?? [];

  function accountName(id: string) {
    return accounts.find((a) => a.id === id)?.name ?? 'Unknown';
  }
  function categoryName(id: string | null) {
    if (!id) return 'Uncategorized';
    return categories.find((c) => c.id === id)?.name ?? 'Unknown';
  }

  function handleDeleteConfirm() {
    if (!deletingId) return;
    deleteRecurring.mutate(deletingId, {
      onSuccess: () => { toast('Recurring transaction deleted', 'success'); setDeletingId(null); },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  function handleToggle(id: string) {
    toggleRecurring.mutate(id, {
      onSuccess: () => toast('Status updated', 'success'),
      onError: (err) => toast(err.message, 'error'),
    });
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Recurring Transactions</h1>
          <p className="text-muted-foreground">Automate repeating income and expenses</p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="mr-2 h-4 w-4" /> New Template
        </Button>
      </div>

      {isLoading ? (
        <ListSkeleton rows={4} />
      ) : items.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Repeat className="h-12 w-12 text-muted-foreground/50 mb-3" />
            <p className="text-lg font-medium text-muted-foreground">No recurring transactions</p>
            <p className="text-sm text-muted-foreground mb-4">Set up templates for repeating transactions</p>
            <Button onClick={() => setShowCreate(true)}>
              <Plus className="mr-2 h-4 w-4" /> New Template
            </Button>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Templates</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {items.map((rt) => (
                <div key={rt.id} className={`flex items-center justify-between rounded-lg border p-4 ${!rt.is_active ? 'opacity-60' : ''}`}>
                  <div className="flex items-center gap-4">
                    <div className={`flex h-10 w-10 items-center justify-center rounded-full ${
                      rt.type === 'income' ? 'bg-success/10 text-success' : 'bg-destructive/10 text-destructive'
                    }`}>
                      <Repeat className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="font-medium">{rt.description || categoryName(rt.category_id)}</p>
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <span>{accountName(rt.account_id)}</span>
                        <span>&middot;</span>
                        <span>{FREQ_LABELS[rt.frequency]}</span>
                        <span>&middot;</span>
                        <span>Next: {formatDate(rt.next_date)}</span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <p className={`font-semibold ${rt.type === 'income' ? 'text-success' : 'text-destructive'}`}>
                      {rt.type === 'income' ? '+' : '-'}{formatCurrency(rt.amount, rt.currency)}
                    </p>
                    <Badge variant={rt.is_active ? 'success' : 'secondary'}>
                      {rt.is_active ? 'Active' : 'Paused'}
                    </Badge>
                    <Button variant="ghost" size="icon" onClick={() => handleToggle(rt.id)} title={rt.is_active ? 'Pause' : 'Resume'}>
                      {rt.is_active ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                    </Button>
                    <Button variant="ghost" size="icon" onClick={() => setDeletingId(rt.id)}>
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      <Dialog open={showCreate} onClose={() => setShowCreate(false)}>
        <DialogHeader>
          <DialogTitle>New Recurring Transaction</DialogTitle>
        </DialogHeader>
        <RecurringForm
          accounts={accounts}
          categories={categories}
          onSubmit={(data) => createRecurring.mutate(data, {
            onSuccess: () => { toast('Recurring transaction created', 'success'); setShowCreate(false); },
            onError: (err) => toast(err.message, 'error'),
          })}
          isPending={createRecurring.isPending}
        />
      </Dialog>

      <ConfirmDialog
        open={!!deletingId}
        onClose={() => setDeletingId(null)}
        onConfirm={handleDeleteConfirm}
        title="Delete Template"
        message="Delete this recurring transaction template? This action cannot be undone."
        isPending={deleteRecurring.isPending}
      />
    </div>
  );
}

function RecurringForm({
  accounts,
  categories,
  onSubmit,
  isPending,
}: {
  accounts: { id: string; name: string; currency: string }[];
  categories: { id: string; name: string; type: string }[];
  onSubmit: (data: any) => void;
  isPending: boolean;
}) {
  const [type, setType] = useState<TransactionType>('expense');
  const [accountId, setAccountId] = useState('');
  const [amount, setAmount] = useState('');
  const [currency, setCurrency] = useState('USD');
  const [categoryId, setCategoryId] = useState('');
  const [description, setDescription] = useState('');
  const [frequency, setFrequency] = useState<Frequency>('monthly');
  const [nextDate, setNextDate] = useState(new Date().toISOString().slice(0, 10));

  const filteredCategories = categories.filter((c) => c.type === type || type === 'transfer');

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    onSubmit({
      account_id: accountId,
      type,
      amount,
      currency,
      category_id: categoryId || undefined,
      description,
      frequency,
      next_date: nextDate,
    });
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Type</label>
          <Select value={type} onChange={(e) => setType(e.target.value as TransactionType)}>
            <option value="expense">Expense</option>
            <option value="income">Income</option>
            <option value="transfer">Transfer</option>
          </Select>
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Frequency</label>
          <Select value={frequency} onChange={(e) => setFrequency(e.target.value as Frequency)}>
            {Object.entries(FREQ_LABELS).map(([k, v]) => (
              <option key={k} value={k}>{v}</option>
            ))}
          </Select>
        </div>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Account</label>
        <Select value={accountId} onChange={(e) => {
          setAccountId(e.target.value);
          const acc = accounts.find((a) => a.id === e.target.value);
          if (acc) setCurrency(acc.currency);
        }} required>
          <option value="">Select account</option>
          {accounts.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}
        </Select>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Amount</label>
          <Input type="number" step="0.01" min="0" value={amount} onChange={(e) => setAmount(e.target.value)} required />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Next Date</label>
          <Input type="date" value={nextDate} onChange={(e) => setNextDate(e.target.value)} required />
        </div>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Category</label>
        <Select value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>
          <option value="">None</option>
          {filteredCategories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
        </Select>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Description</label>
        <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="e.g. Monthly rent" />
      </div>

      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Creating...' : 'Create Template'}
      </Button>
    </form>
  );
}
