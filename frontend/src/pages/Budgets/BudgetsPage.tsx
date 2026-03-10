import { useState, type FormEvent } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/ConfirmDialog';
import { ListSkeleton } from '@/components/ui/skeleton';
import { useBudgets, useCreateBudget, useUpdateBudget, useDeleteBudget, useCategories } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { formatCurrency } from '@/lib/utils';
import { Plus, Trash2, Pencil, PieChart, AlertTriangle } from 'lucide-react';
import type { Budget, BudgetPeriod } from '@/types';

export function BudgetsPage() {
  const { data, isLoading } = useBudgets();
  const { data: catData } = useCategories();
  const createBudget = useCreateBudget();
  const updateBudget = useUpdateBudget();
  const deleteBudget = useDeleteBudget();
  const { toast } = useToast();
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<Budget | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const budgets = data?.budgets ?? [];
  const categories = (catData?.categories ?? []).filter((c) => c.type === 'expense');

  const totalBudgeted = budgets.reduce((s, b) => s + b.amount, 0);
  const totalSpent = budgets.reduce((s, b) => s + b.spent, 0);
  const overBudget = budgets.filter((b) => b.percentage >= 100);

  function handleDeleteConfirm() {
    if (!deletingId) return;
    deleteBudget.mutate(deletingId, {
      onSuccess: () => { toast('Budget deleted', 'success'); setDeletingId(null); },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Budgets</h1>
          <p className="text-muted-foreground">Track spending against your budget goals</p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="mr-2 h-4 w-4" /> New Budget
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Budgeted</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(totalBudgeted)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Spent</CardTitle>
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${totalSpent > totalBudgeted ? 'text-destructive' : 'text-success'}`}>
              {formatCurrency(totalSpent)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Over Budget</CardTitle>
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${overBudget.length > 0 ? 'text-destructive' : ''}`}>
              {overBudget.length} {overBudget.length === 1 ? 'category' : 'categories'}
            </div>
          </CardContent>
        </Card>
      </div>

      {isLoading ? (
        <ListSkeleton rows={4} />
      ) : budgets.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <PieChart className="h-12 w-12 text-muted-foreground/50 mb-3" />
            <p className="text-lg font-medium text-muted-foreground">No budgets yet</p>
            <p className="text-sm text-muted-foreground mb-4">Create your first budget to start tracking spending</p>
            <Button onClick={() => setShowCreate(true)}>
              <Plus className="mr-2 h-4 w-4" /> New Budget
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {budgets.map((budget) => (
            <Card key={budget.id} className={budget.percentage >= 100 ? 'border-destructive/50' : ''}>
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div
                      className="h-3 w-3 rounded-full"
                      style={{ backgroundColor: budget.category_color || '#94a3b8' }}
                    />
                    <CardTitle className="text-base">{budget.category_name || 'Uncategorized'}</CardTitle>
                  </div>
                  <div className="flex gap-1">
                    <Button variant="ghost" size="icon" onClick={() => setEditing(budget)}>
                      <Pencil className="h-3.5 w-3.5" />
                    </Button>
                    <Button variant="ghost" size="icon" onClick={() => setDeletingId(budget.id)}>
                      <Trash2 className="h-3.5 w-3.5 text-destructive" />
                    </Button>
                  </div>
                </div>
                <CardDescription className="capitalize">{budget.period}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex items-end justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">Spent</p>
                    <p className={`text-xl font-bold ${budget.percentage >= 100 ? 'text-destructive' : budget.percentage >= 80 ? 'text-amber-500' : 'text-foreground'}`}>
                      {formatCurrency(budget.spent)}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground">of {formatCurrency(budget.amount)}</p>
                    <p className="text-sm font-medium">{formatCurrency(budget.remaining)} left</p>
                  </div>
                </div>

                <div className="space-y-1">
                  <div className="h-3 w-full rounded-full bg-muted overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all ${
                        budget.percentage >= 100 ? 'bg-destructive' : budget.percentage >= 80 ? 'bg-amber-500' : 'bg-primary'
                      }`}
                      style={{ width: `${Math.min(budget.percentage, 100)}%` }}
                    />
                  </div>
                  <div className="flex justify-between text-xs text-muted-foreground">
                    <span>{Math.round(budget.percentage)}%</span>
                    {budget.percentage >= 100 && (
                      <span className="flex items-center gap-1 text-destructive">
                        <AlertTriangle className="h-3 w-3" /> Over budget!
                      </span>
                    )}
                    {budget.rollover && <span>Rollover enabled</span>}
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={showCreate} onClose={() => setShowCreate(false)}>
        <DialogHeader>
          <DialogTitle>Create Budget</DialogTitle>
        </DialogHeader>
        <BudgetForm
          categories={categories}
          onSuccess={() => setShowCreate(false)}
          onSubmit={(data) => createBudget.mutate(data, {
            onSuccess: () => { toast('Budget created', 'success'); setShowCreate(false); },
            onError: (err) => toast(err.message, 'error'),
          })}
          isPending={createBudget.isPending}
        />
      </Dialog>

      <Dialog open={!!editing} onClose={() => setEditing(null)}>
        <DialogHeader>
          <DialogTitle>Edit Budget</DialogTitle>
        </DialogHeader>
        {editing && (
          <BudgetForm
            categories={categories}
            budget={editing}
            onSuccess={() => setEditing(null)}
            onSubmit={(data) => updateBudget.mutate({ id: editing.id, ...data }, {
              onSuccess: () => { toast('Budget updated', 'success'); setEditing(null); },
              onError: (err) => toast(err.message, 'error'),
            })}
            isPending={updateBudget.isPending}
          />
        )}
      </Dialog>

      <ConfirmDialog
        open={!!deletingId}
        onClose={() => setDeletingId(null)}
        onConfirm={handleDeleteConfirm}
        title="Delete Budget"
        message="Delete this budget? This action cannot be undone."
        isPending={deleteBudget.isPending}
      />
    </div>
  );
}

function BudgetForm({
  categories,
  budget,
  onSubmit,
  onSuccess: _onSuccess,
  isPending,
}: {
  categories: { id: string; name: string }[];
  budget?: Budget;
  onSubmit: (data: any) => void;
  onSuccess: () => void;
  isPending: boolean;
}) {
  const [categoryId, setCategoryId] = useState(budget?.category_id ?? '');
  const [amount, setAmount] = useState(String(budget?.amount ?? ''));
  const [period, setPeriod] = useState<BudgetPeriod>(budget?.period ?? 'monthly');
  const [startDate, setStartDate] = useState(
    budget?.start_date ? new Date(budget.start_date).toISOString().slice(0, 10) : new Date().toISOString().slice(0, 10),
  );
  const [rollover, setRollover] = useState(budget?.rollover ?? false);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    onSubmit({
      category_id: categoryId,
      amount: parseFloat(amount),
      period,
      start_date: startDate,
      rollover,
    });
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <label className="text-sm font-medium">Category</label>
        <Select value={categoryId} onChange={(e) => setCategoryId(e.target.value)} required disabled={!!budget}>
          <option value="">Select a category</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </Select>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Amount</label>
          <Input type="number" step="0.01" min="0" value={amount} onChange={(e) => setAmount(e.target.value)} required />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Period</label>
          <Select value={period} onChange={(e) => setPeriod(e.target.value as BudgetPeriod)}>
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
            <option value="yearly">Yearly</option>
          </Select>
        </div>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Start Date</label>
        <Input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} required />
      </div>

      <label className="flex items-center gap-2 cursor-pointer">
        <input type="checkbox" checked={rollover} onChange={(e) => setRollover(e.target.checked)} className="rounded" />
        <span className="text-sm font-medium">Enable rollover (carry unused budget to next period)</span>
      </label>

      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Saving...' : budget ? 'Update Budget' : 'Create Budget'}
      </Button>
    </form>
  );
}
