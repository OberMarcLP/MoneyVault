import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select } from '@/components/ui/select';
import { CardSkeleton } from '@/components/ui/skeleton';
import { useSpendingTrends } from '@/api/hooks';
import { formatCurrency } from '@/lib/utils';
import {
  LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer, Legend,
} from 'recharts';
import { StatCard, tooltipFmt } from './report-utils';

export default function TrendsTab() {
  const [months, setMonths] = useState(12);
  const { data, isLoading } = useSpendingTrends(months);

  if (isLoading) return <CardSkeleton />;
  const trends = (data ?? []).sort((a, b) => a.period.localeCompare(b.period));

  const chartData = trends.map(t => ({
    period: t.period,
    income: t.income,
    expense: t.expense,
    net: t.net,
  }));

  const avgIncome = trends.length > 0 ? trends.reduce((s, t) => s + t.income, 0) / trends.length : 0;
  const avgExpense = trends.length > 0 ? trends.reduce((s, t) => s + t.expense, 0) / trends.length : 0;

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Select value={String(months)} onChange={(e) => setMonths(Number(e.target.value))}>
          <option value="6">6 months</option>
          <option value="12">12 months</option>
          <option value="24">24 months</option>
        </Select>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label="Avg Monthly Income" value={formatCurrency(avgIncome)} positive />
        <StatCard label="Avg Monthly Expenses" value={formatCurrency(avgExpense)} />
        <StatCard label="Avg Monthly Savings" value={formatCurrency(avgIncome - avgExpense)} positive={avgIncome >= avgExpense} />
      </div>

      <Card>
        <CardHeader><CardTitle className="text-lg">Income vs Expenses</CardTitle></CardHeader>
        <CardContent>
          {chartData.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">No trend data available</p>
          ) : (
            <ResponsiveContainer width="100%" height={350}>
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                <XAxis dataKey="period" tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} />
                <YAxis tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`} />
                <Tooltip
                  contentStyle={{ background: 'hsl(var(--card))', border: '1px solid hsl(var(--border))', borderRadius: 8 }}
                  formatter={tooltipFmt}
                />
                <Legend />
                <Bar dataKey="income" fill="#22c55e" name="Income" radius={[4, 4, 0, 0]} />
                <Bar dataKey="expense" fill="#ef4444" name="Expenses" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </CardContent>
      </Card>

      {chartData.length > 0 && (
        <Card>
          <CardHeader><CardTitle className="text-lg">Net Savings Trend</CardTitle></CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                <XAxis dataKey="period" tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} />
                <YAxis tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} tickFormatter={(v) => `$${v.toFixed(0)}`} />
                <Tooltip
                  contentStyle={{ background: 'hsl(var(--card))', border: '1px solid hsl(var(--border))', borderRadius: 8 }}
                  formatter={tooltipFmt}
                />
                <Line type="monotone" dataKey="net" stroke="#6366f1" strokeWidth={2} dot={{ r: 4 }} name="Net Savings" />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
