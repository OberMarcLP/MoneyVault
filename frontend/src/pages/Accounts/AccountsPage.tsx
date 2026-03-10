import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/ConfirmDialog';
import { CardSkeleton } from '@/components/ui/skeleton';
import { AccountForm } from '@/components/forms/AccountForm';
import { useAccounts, useDeleteAccount } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { formatCurrency, getAccountTypeLabel } from '@/lib/utils';
import { Plus, Pencil, Trash2, Wallet } from 'lucide-react';
import type { Account } from '@/types';

export function AccountsPage() {
  const { data, isLoading } = useAccounts();
  const deleteAccount = useDeleteAccount();
  const { toast } = useToast();
  const [editingAccount, setEditingAccount] = useState<Account | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [deletingAccount, setDeletingAccount] = useState<Account | null>(null);

  const accounts = data?.accounts ?? [];

  const grouped = accounts.reduce<Record<string, Account[]>>((acc, account) => {
    const group = account.type;
    if (!acc[group]) acc[group] = [];
    acc[group].push(account);
    return acc;
  }, {});

  function handleDeleteConfirm() {
    if (!deletingAccount) return;
    deleteAccount.mutate(deletingAccount.id, {
      onSuccess: () => { toast('Account deleted', 'success'); setDeletingAccount(null); },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Accounts</h1>
          <p className="text-muted-foreground">Manage your financial accounts</p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Account
        </Button>
      </div>

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => <CardSkeleton key={i} />)}
        </div>
      ) : accounts.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Wallet className="h-12 w-12 text-muted-foreground/50 mb-3" />
            <p className="text-lg font-medium text-muted-foreground">No accounts yet</p>
            <p className="text-sm text-muted-foreground mb-4">Create your first account to start tracking</p>
            <Button onClick={() => setShowCreate(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Add Account
            </Button>
          </CardContent>
        </Card>
      ) : (
        Object.entries(grouped).map(([type, accts]) => (
          <div key={type}>
            <h2 className="text-lg font-semibold mb-3">{getAccountTypeLabel(type)}</h2>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {accts.map((account) => (
                <Card key={account.id}>
                  <CardHeader className="flex flex-row items-start justify-between pb-2">
                    <div>
                      <CardTitle className="text-base">{account.name}</CardTitle>
                      {account.institution && (
                        <p className="text-sm text-muted-foreground">{account.institution}</p>
                      )}
                    </div>
                    <Badge variant={account.is_active ? 'default' : 'secondary'}>
                      {account.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                  </CardHeader>
                  <CardContent>
                    <p className={`text-2xl font-bold ${parseFloat(account.balance) >= 0 ? 'text-success' : 'text-destructive'}`}>
                      {formatCurrency(account.balance, account.currency)}
                    </p>
                    <div className="mt-4 flex gap-2">
                      <Button variant="outline" size="sm" onClick={() => setEditingAccount(account)}>
                        <Pencil className="mr-1 h-3 w-3" /> Edit
                      </Button>
                      <Button variant="outline" size="sm" onClick={() => setDeletingAccount(account)} className="text-destructive">
                        <Trash2 className="mr-1 h-3 w-3" /> Delete
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        ))
      )}

      <Dialog open={showCreate} onClose={() => setShowCreate(false)}>
        <DialogHeader>
          <DialogTitle>Create Account</DialogTitle>
        </DialogHeader>
        <AccountForm onSuccess={() => setShowCreate(false)} />
      </Dialog>

      <Dialog open={!!editingAccount} onClose={() => setEditingAccount(null)}>
        <DialogHeader>
          <DialogTitle>Edit Account</DialogTitle>
        </DialogHeader>
        {editingAccount && (
          <AccountForm
            account={editingAccount}
            onSuccess={() => setEditingAccount(null)}
          />
        )}
      </Dialog>

      <ConfirmDialog
        open={!!deletingAccount}
        onClose={() => setDeletingAccount(null)}
        onConfirm={handleDeleteConfirm}
        title="Delete Account"
        message={`Delete account "${deletingAccount?.name}"? This will also delete all associated transactions.`}
        isPending={deleteAccount.isPending}
      />
    </div>
  );
}
