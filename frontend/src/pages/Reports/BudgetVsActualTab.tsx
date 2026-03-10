import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select } from '@/components/ui/select';
import { CardSkeleton } from '@/components/ui/skeleton';
import { useBudgetVsActual } from '@/api/hooks';
import { formatCurrency } from '@/lib/utils';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, Legend,
} from 'recharts';
import { Target } from 'lucide-react';
import { StatCard, tooltipFmt } from './report-utils';

export default function BudgetVsActualTab() {
  const [period, setPeriod] = useState('month');
  const { data, isLoading } = useBudgetVsActual(period);

  if (isLoading) return <CardSkeleton />;
  const history = data;

  if (!history || history.budgets.length === 0) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <Target className="h-12 w-12 mx-auto text-muted-foreground mb-3" />
          <p className="text-muted-foreground">No budgets configured yet</p>
          <p className="text-xs text-muted-foreground mt-1">Create budgets in the Budgets page to see comparisons here</p>
        </CardContent>
      </Card>
    );
  }

  const chartData = history.budgets.map(b => ({
    name: b.category_name || 'Unknown',
    budget: b.budget_amount,
    actual: b.actual_amount,
    color: b.category_color,
  }));

  const overBudget = history.budgets.filter(b => b.percentage >= 100);
  const onTrack = history.budgets.filter(b => b.percentage < 80);

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Select value={period} onChange={(e) => setPeriod(e.target.value)}>
          <option value="week">This Week</option>
          <option value="month">This Month</option>
          <option value="quarter">This Quarter</option>
          <option value="year">This Year</option>
        </Select>
      </div>

      <div className="grid gap-4 sm:grid-cols-4">
        <StatCard label="Total Budget" value={formatCurrency(history.total_budget)} />
        <StatCard label="Total Spent" value={formatCurrency(history.total_actual)} />
        <StatCard label="On Track" value={String(onTrack.length)} positive />
        <StatCard label="Over Budget" value={String(overBudget.length)} positive={overBudget.length === 0} />
      </div>

      <Card>
        <CardHeader><CardTitle className="text-lg">Budget vs Actual</CardTitle></CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={350}>
            <BarChart data={chartData} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
              <XAxis type="number" tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} tickFormatter={(v) => `$${v}`} />
              <YAxis type="category" dataKey="name" tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} width={120} />
              <Tooltip
                contentStyle={{ background: 'hsl(var(--card))', border: '1px solid hsl(var(--border))', borderRadius: 8 }}
                formatter={tooltipFmt}
              />
              <Legend />
              <Bar dataKey="budget" fill="#6366f1" name="Budget" radius={[0, 4, 4, 0]} />
              <Bar dataKey="actual" fill="#f59e0b" name="Actual" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {history.budgets.map(b => (
          <Card key={b.category_id}>
            <CardContent className="p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="font-medium text-sm">{b.category_name}</span>
                <span className={`text-xs font-medium ${
                  b.percentage >= 100 ? 'text-destructive' : b.percentage >= 80 ? 'text-amber-500' : 'text-success'
                }`}>
                  {b.percentage.toFixed(0)}%
                </span>
              </div>
              <div className="h-2 w-full rounded-full bg-muted overflow-hidden mb-2">
                <div
                  className={`h-full rounded-full transition-all ${
                    b.percentage >= 100 ? 'bg-destructive' : b.percentage >= 80 ? 'bg-amber-500' : 'bg-success'
                  }`}
                  style={{ width: `${Math.min(b.percentage, 100)}%` }}
                />
              </div>
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>Spent: {formatCurrency(b.actual_amount)}</span>
                <span>Budget: {formatCurrency(b.budget_amount)}</span>
              </div>
              {b.difference < 0 && (
                <p className="text-xs text-destructive mt-1">Over by {formatCurrency(Math.abs(b.difference))}</p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {overBudget.length > 0 && (
        <Card className="border-amber-500/30 bg-amber-500/5">
          <CardContent className="p-4">
            <p className="text-sm font-medium text-amber-500 mb-2">Insights</p>
            {overBudget.map(b => (
              <p key={b.category_id} className="text-sm text-muted-foreground">
                You overspent on <span className="font-medium">{b.category_name}</span> by{' '}
                <span className="text-destructive font-medium">{formatCurrency(Math.abs(b.difference))}</span>{' '}
                ({b.percentage.toFixed(0)}% of budget)
              </p>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
