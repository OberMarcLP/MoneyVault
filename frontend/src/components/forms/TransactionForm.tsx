import { useState, type FormEvent } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { FormField } from '@/components/ui/FormField';
import { useCreateTransaction, useUpdateTransaction, useAccounts, useCategories } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import type { Transaction, TransactionType } from '@/types';

interface TransactionFormProps {
  transaction?: Transaction;
  onSuccess: () => void;
}

export function TransactionForm({ transaction, onSuccess }: TransactionFormProps) {
  const { data: accountsData } = useAccounts();
  const { data: categoriesData } = useCategories();

  const accounts = accountsData?.accounts ?? [];
  const categories = categoriesData?.categories ?? [];

  const [type, setType] = useState<TransactionType>(transaction?.type ?? 'expense');
  const [accountId, setAccountId] = useState(transaction?.account_id ?? accounts[0]?.id ?? '');
  const [amount, setAmount] = useState(transaction?.amount ?? '');
  const [currency, setCurrency] = useState(transaction?.currency ?? 'USD');
  const [categoryId, setCategoryId] = useState(transaction?.category_id ?? '');
  const [description, setDescription] = useState(transaction?.description ?? '');
  const [date, setDate] = useState(transaction ? transaction.date.split('T')[0] : new Date().toISOString().split('T')[0]);
  const [transferAccountId, setTransferAccountId] = useState(transaction?.transfer_account_id ?? '');
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const createTx = useCreateTransaction();
  const updateTx = useUpdateTransaction();
  const { toast } = useToast();
  const isEditing = !!transaction;

  const filteredCategories = categories.filter((c) =>
    type === 'transfer' ? true : c.type === type || (type === 'expense' && c.type === 'expense') || (type === 'income' && c.type === 'income')
  );

  const errors: Record<string, string> = {};
  if (!accountId) errors.account = 'Account is required';
  if (!amount || parseFloat(amount) <= 0) errors.amount = 'Amount must be greater than 0';
  if (!date) errors.date = 'Date is required';
  if (type === 'transfer' && !transferAccountId) errors.transfer = 'Destination account is required';

  const isValid = Object.keys(errors).length === 0;

  function blur(field: string) {
    setTouched((t) => ({ ...t, [field]: true }));
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setTouched({ account: true, amount: true, date: true, transfer: true });
    if (!isValid) return;

    const payload = {
      type,
      account_id: accountId,
      amount,
      currency,
      category_id: categoryId || undefined,
      description,
      date,
      transfer_account_id: type === 'transfer' ? transferAccountId : undefined,
    };

    const opts = {
      onSuccess: () => { toast(isEditing ? 'Transaction updated' : 'Transaction added', 'success'); onSuccess(); },
      onError: (err: Error) => toast(err.message, 'error'),
    };
    if (isEditing) {
      updateTx.mutate({ id: transaction.id, ...payload }, opts);
    } else {
      createTx.mutate(payload, opts);
    }
  }

  const isPending = createTx.isPending || updateTx.isPending;
  const error = createTx.error || updateTx.error;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <label className="text-sm font-medium">Type</label>
        <Select value={type} onChange={(e) => setType(e.target.value as TransactionType)}>
          <option value="expense">Expense</option>
          <option value="income">Income</option>
          <option value="transfer">Transfer</option>
        </Select>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <FormField label={type === 'transfer' ? 'From Account' : 'Account'} error={touched.account ? errors.account : undefined}>
          <Select value={accountId} onChange={(e) => setAccountId(e.target.value)} onBlur={() => blur('account')}>
            <option value="">Select account</option>
            {accounts.map((a) => (
              <option key={a.id} value={a.id}>{a.name}</option>
            ))}
          </Select>
        </FormField>
        <FormField label="Date" error={touched.date ? errors.date : undefined}>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} onBlur={() => blur('date')} />
        </FormField>
      </div>

      {type === 'transfer' && (
        <FormField label="To Account" error={touched.transfer ? errors.transfer : undefined}>
          <Select value={transferAccountId} onChange={(e) => setTransferAccountId(e.target.value)} onBlur={() => blur('transfer')}>
            <option value="">Select destination</option>
            {accounts.filter((a) => a.id !== accountId).map((a) => (
              <option key={a.id} value={a.id}>{a.name}</option>
            ))}
          </Select>
        </FormField>
      )}

      <div className="grid grid-cols-2 gap-4">
        <FormField label="Amount" error={touched.amount ? errors.amount : undefined}>
          <Input type="number" step="0.01" min="0" value={amount} onChange={(e) => setAmount(e.target.value)} onBlur={() => blur('amount')} />
        </FormField>
        <div className="space-y-2">
          <label className="text-sm font-medium">Currency</label>
          <Select value={currency} onChange={(e) => setCurrency(e.target.value)}>
            <option value="USD">USD</option>
            <option value="EUR">EUR</option>
            <option value="GBP">GBP</option>
            <option value="CHF">CHF</option>
            <option value="NOK">NOK</option>
          </Select>
        </div>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Category</label>
        <Select value={categoryId ?? ''} onChange={(e) => setCategoryId(e.target.value)}>
          <option value="">No category</option>
          {filteredCategories.map((c) => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </Select>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Description</label>
        <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="What was this for?" />
      </div>

      {error && <p className="text-sm text-destructive">{error.message}</p>}

      <Button type="submit" className="w-full" loading={isPending} disabled={!isValid}>
        {isEditing ? 'Update Transaction' : 'Add Transaction'}
      </Button>
    </form>
  );
}
