import { useState, type FormEvent } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { FormField } from '@/components/ui/FormField';
import { useCreateAccount, useUpdateAccount } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import type { Account, AccountType } from '@/types';

interface AccountFormProps {
  account?: Account;
  onSuccess: () => void;
}

export function AccountForm({ account, onSuccess }: AccountFormProps) {
  const [name, setName] = useState(account?.name ?? '');
  const [type, setType] = useState<AccountType>(account?.type ?? 'checking');
  const [currency, setCurrency] = useState(account?.currency ?? 'USD');
  const [balance, setBalance] = useState(account?.balance ?? '0');
  const [institution, setInstitution] = useState(account?.institution ?? '');
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const createAccount = useCreateAccount();
  const updateAccount = useUpdateAccount();
  const { toast } = useToast();
  const isEditing = !!account;

  const errors: Record<string, string> = {};
  if (!name.trim()) errors.name = 'Name is required';
  if (balance !== '' && isNaN(parseFloat(balance))) errors.balance = 'Balance must be a number';

  const isValid = Object.keys(errors).length === 0;

  function blur(field: string) {
    setTouched((t) => ({ ...t, [field]: true }));
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setTouched({ name: true, balance: true });
    if (!isValid) return;

    const data = { name, type, currency, balance, institution: institution || undefined };
    const opts = {
      onSuccess: () => { toast(isEditing ? 'Account updated' : 'Account created', 'success'); onSuccess(); },
      onError: (err: Error) => toast(err.message, 'error'),
    };
    if (isEditing) {
      updateAccount.mutate({ id: account.id, ...data }, opts);
    } else {
      createAccount.mutate(data, opts);
    }
  }

  const isPending = createAccount.isPending || updateAccount.isPending;
  const error = createAccount.error || updateAccount.error;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <FormField label="Name" error={touched.name ? errors.name : undefined}>
        <Input value={name} onChange={(e) => setName(e.target.value)} onBlur={() => blur('name')} placeholder="e.g. Main Checking" />
      </FormField>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Type</label>
          <Select value={type} onChange={(e) => setType(e.target.value as AccountType)}>
            <option value="checking">Checking</option>
            <option value="savings">Savings</option>
            <option value="credit">Credit Card</option>
            <option value="investment">Investment</option>
            <option value="crypto_wallet">Crypto Wallet</option>
          </Select>
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Currency</label>
          <Select value={currency} onChange={(e) => setCurrency(e.target.value)}>
            <option value="USD">USD</option>
            <option value="EUR">EUR</option>
            <option value="GBP">GBP</option>
            <option value="CHF">CHF</option>
            <option value="JPY">JPY</option>
            <option value="CAD">CAD</option>
            <option value="AUD">AUD</option>
            <option value="NOK">NOK</option>
            <option value="SEK">SEK</option>
          </Select>
        </div>
      </div>

      <FormField label="Balance" error={touched.balance ? errors.balance : undefined}>
        <Input type="number" step="0.01" value={balance} onChange={(e) => setBalance(e.target.value)} onBlur={() => blur('balance')} />
      </FormField>

      <div className="space-y-2">
        <label className="text-sm font-medium">Institution (optional)</label>
        <Input value={institution} onChange={(e) => setInstitution(e.target.value)} placeholder="e.g. Chase, Fidelity" />
      </div>

      {error && <p className="text-sm text-destructive">{error.message}</p>}

      <Button type="submit" className="w-full" loading={isPending} disabled={!isValid}>
        {isEditing ? 'Update Account' : 'Create Account'}
      </Button>
    </form>
  );
}
