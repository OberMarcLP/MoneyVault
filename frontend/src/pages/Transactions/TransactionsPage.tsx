import { useState, useEffect, useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/ConfirmDialog';
import { ListSkeleton } from '@/components/ui/skeleton';
import { TransactionForm } from '@/components/forms/TransactionForm';
import { useTransactions, useDeleteTransaction, useAccounts, useCategories } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { formatCurrency, formatDate } from '@/lib/utils';
import { CategoryIcon } from '@/components/ui/CategoryIcon';
import { Plus, Trash2, Pencil, ChevronLeft, ChevronRight, ArrowLeftRight, Search, X } from 'lucide-react';
import type { Transaction, TransactionFilter, TransactionType } from '@/types';

export function TransactionsPage() {
  const [filter, setFilter] = useState<TransactionFilter>({ page: 1, per_page: 20 });
  const [searchInput, setSearchInput] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<Transaction | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => {
    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      setFilter((f) => ({ ...f, search: searchInput || undefined, page: 1 }));
    }, 300);
    return () => clearTimeout(debounceRef.current);
  }, [searchInput]);

  const { data, isLoading } = useTransactions(filter);
  const { data: accountsData } = useAccounts();
  const { data: categoriesData } = useCategories();
  const deleteTransaction = useDeleteTransaction();
  const { toast } = useToast();

  const transactions = data?.data ?? [];
  const accounts = accountsData?.accounts ?? [];
  const categories = categoriesData?.categories ?? [];

  function accountName(id: string) {
    return accounts.find((a) => a.id === id)?.name ?? 'Unknown';
  }

  function getCategory(id: string | null) {
    if (!id) return null;
    return categories.find((c) => c.id === id) ?? null;
  }

  function handleDeleteConfirm() {
    if (!deletingId) return;
    deleteTransaction.mutate(deletingId, {
      onSuccess: () => { toast('Transaction deleted', 'success'); setDeletingId(null); },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Transactions</h1>
          <p className="text-muted-foreground">Track income, expenses, and transfers</p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Transaction
        </Button>
      </div>

      <div className="flex flex-wrap gap-3">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="Search transactions..."
            className="w-64 pl-9 pr-8"
          />
          {searchInput && (
            <button
              onClick={() => setSearchInput('')}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>

        <Select
          value={filter.type ?? ''}
          onChange={(e) => setFilter((f) => ({ ...f, type: (e.target.value || undefined) as TransactionType | undefined, page: 1 }))}
          className="w-40"
        >
          <option value="">All Types</option>
          <option value="income">Income</option>
          <option value="expense">Expense</option>
          <option value="transfer">Transfer</option>
        </Select>

        <Select
          value={filter.account_id ?? ''}
          onChange={(e) => setFilter((f) => ({ ...f, account_id: e.target.value || undefined, page: 1 }))}
          className="w-48"
        >
          <option value="">All Accounts</option>
          {accounts.map((a) => (
            <option key={a.id} value={a.id}>{a.name}</option>
          ))}
        </Select>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">
            {data ? `${data.total} transaction${data.total !== 1 ? 's' : ''}` : 'Loading...'}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <ListSkeleton rows={5} />
          ) : transactions.length === 0 ? (
            <div className="flex flex-col items-center py-12">
              <ArrowLeftRight className="h-12 w-12 text-muted-foreground/50 mb-3" />
              <p className="text-lg font-medium text-muted-foreground">No transactions found</p>
              <p className="text-sm text-muted-foreground">Add your first transaction to get started</p>
            </div>
          ) : (
            <div className="space-y-2">
              {transactions.map((tx) => (
                <div key={tx.id} className="flex items-center justify-between rounded-lg border p-3 hover:bg-accent/50 transition-colors">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="font-medium truncate">{tx.description || 'No description'}</p>
                      <Badge variant={tx.type === 'income' ? 'success' : tx.type === 'expense' ? 'destructive' : 'secondary'}>
                        {tx.type}
                      </Badge>
                    </div>
                    <div className="flex items-center gap-3 text-sm text-muted-foreground mt-1">
                      <span>{formatDate(tx.date)}</span>
                      <span>{accountName(tx.account_id)}</span>
                      {(() => {
                        const cat = getCategory(tx.category_id);
                        return cat ? (
                          <span className="flex items-center gap-1">
                            <CategoryIcon icon={cat.icon} color={cat.color} size="sm" />
                            {cat.name}
                          </span>
                        ) : (
                          <span>Uncategorized</span>
                        );
                      })()}
                    </div>
                  </div>
                  <div className="flex items-center gap-3 ml-4">
                    <p className={`font-semibold whitespace-nowrap ${
                      tx.type === 'income' ? 'text-success' : tx.type === 'expense' ? 'text-destructive' : 'text-primary'
                    }`}>
                      {tx.type === 'income' ? '+' : tx.type === 'expense' ? '-' : ''}
                      {formatCurrency(tx.amount, tx.currency)}
                    </p>
                    <Button variant="ghost" size="icon" onClick={() => setEditing(tx)}>
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button variant="ghost" size="icon" onClick={() => setDeletingId(tx.id)}>
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {data && data.total_pages > 1 && (
            <div className="flex items-center justify-between mt-4 pt-4 border-t">
              <Button
                variant="outline"
                size="sm"
                disabled={filter.page === 1}
                onClick={() => setFilter((f) => ({ ...f, page: (f.page ?? 1) - 1 }))}
              >
                <ChevronLeft className="mr-1 h-4 w-4" /> Previous
              </Button>
              <span className="text-sm text-muted-foreground">
                Page {data.page} of {data.total_pages}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={data.page >= data.total_pages}
                onClick={() => setFilter((f) => ({ ...f, page: (f.page ?? 1) + 1 }))}
              >
                Next <ChevronRight className="ml-1 h-4 w-4" />
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={showCreate} onClose={() => setShowCreate(false)}>
        <DialogHeader>
          <DialogTitle>Add Transaction</DialogTitle>
        </DialogHeader>
        <TransactionForm onSuccess={() => setShowCreate(false)} />
      </Dialog>

      <Dialog open={!!editing} onClose={() => setEditing(null)}>
        <DialogHeader>
          <DialogTitle>Edit Transaction</DialogTitle>
        </DialogHeader>
        {editing && <TransactionForm transaction={editing} onSuccess={() => setEditing(null)} />}
      </Dialog>

      <ConfirmDialog
        open={!!deletingId}
        onClose={() => setDeletingId(null)}
        onConfirm={handleDeleteConfirm}
        title="Delete Transaction"
        message="Delete this transaction? This action cannot be undone."
        isPending={deleteTransaction.isPending}
      />
    </div>
  );
}
