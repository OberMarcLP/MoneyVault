import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select } from '@/components/ui/select';
import { CardSkeleton } from '@/components/ui/skeleton';
import { useSpendingBreakdown, useTopExpenses } from '@/api/hooks';
import { formatCurrency, formatDate } from '@/lib/utils';
import {
  PieChart, Pie, Cell, Tooltip, ResponsiveContainer,
} from 'recharts';
import { StatCard, tooltipFmt, CHART_COLORS } from './report-utils';

export default function SpendingTab() {
  const [period, setPeriod] = useState('month');
  const { data: breakdown, isLoading: breakdownLoading } = useSpendingBreakdown(period);
  const { data: topExpenses, isLoading: topLoading } = useTopExpenses(period, 10);

  if (breakdownLoading) return <CardSkeleton />;
  const categories = breakdown ?? [];
  const expenses = topExpenses ?? [];
  const totalSpending = categories.reduce((s, c) => s + c.total, 0);

  const pieData = categories.map((c, i) => ({
    name: c.category_name || 'Uncategorized',
    value: c.total,
    color: c.category_color || CHART_COLORS[i % CHART_COLORS.length],
  }));

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Select value={period} onChange={(e) => setPeriod(e.target.value)}>
          <option value="week">This Week</option>
          <option value="month">This Month</option>
          <option value="quarter">This Quarter</option>
          <option value="year">This Year</option>
          <option value="ytd">Year to Date</option>
        </Select>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label="Total Spending" value={formatCurrency(totalSpending)} />
        <StatCard label="Categories" value={String(categories.length)} />
        <StatCard label="Transactions" value={String(categories.reduce((s, c) => s + c.count, 0))} />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle className="text-lg">By Category</CardTitle></CardHeader>
          <CardContent>
            {pieData.length === 0 ? (
              <p className="text-center text-muted-foreground py-8">No spending data</p>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={100} label={({ name, percent }: { name?: string; percent?: number }) => `${name ?? ''} ${((percent ?? 0) * 100).toFixed(0)}%`}>
                    {pieData.map((entry, i) => <Cell key={i} fill={entry.color} />)}
                  </Pie>
                  <Tooltip formatter={tooltipFmt} />
                </PieChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle className="text-lg">Category Breakdown</CardTitle></CardHeader>
          <CardContent>
            {categories.length === 0 ? (
              <p className="text-center text-muted-foreground py-8">No data</p>
            ) : (
              <div className="space-y-3">
                {categories.map((c, i) => (
                  <div key={c.category_id} className="space-y-1.5">
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <div className="h-3 w-3 rounded-full" style={{ backgroundColor: c.category_color || CHART_COLORS[i % CHART_COLORS.length] }} />
                        <span className="font-medium">{c.category_name || 'Uncategorized'}</span>
                        <span className="text-xs text-muted-foreground">{c.count} txns</span>
                      </div>
                      <span className="font-medium">{formatCurrency(c.total)}</span>
                    </div>
                    <div className="h-2 w-full rounded-full bg-muted overflow-hidden">
                      <div
                        className="h-full rounded-full transition-all"
                        style={{ width: `${totalSpending > 0 ? (c.total / totalSpending) * 100 : 0}%`, backgroundColor: c.category_color || CHART_COLORS[i % CHART_COLORS.length] }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {!topLoading && expenses.length > 0 && (
        <Card>
          <CardHeader><CardTitle className="text-lg">Top Expenses</CardTitle></CardHeader>
          <CardContent>
            <div className="rounded-lg border overflow-hidden">
              <table className="w-full">
                <thead className="bg-muted/50">
                  <tr className="text-xs text-muted-foreground">
                    <th className="text-left p-3 font-medium">Description</th>
                    <th className="text-left p-3 font-medium">Category</th>
                    <th className="text-right p-3 font-medium">Amount</th>
                    <th className="text-right p-3 font-medium">Date</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {expenses.map((e, i) => (
                    <tr key={i} className="hover:bg-muted/30">
                      <td className="p-3 text-sm">{e.description || '\u2014'}</td>
                      <td className="p-3 text-sm text-muted-foreground">{e.category || '\u2014'}</td>
                      <td className="p-3 text-sm text-right font-medium text-destructive">{formatCurrency(e.amount)}</td>
                      <td className="p-3 text-sm text-right text-muted-foreground">{formatDate(e.date)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
