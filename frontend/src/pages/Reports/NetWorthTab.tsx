import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select } from '@/components/ui/select';
import { CardSkeleton } from '@/components/ui/skeleton';
import { useNetWorthHistory } from '@/api/hooks';
import { formatCurrency, formatDate } from '@/lib/utils';
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, Legend,
} from 'recharts';
import { StatCard, tooltipFmt } from './report-utils';

export default function NetWorthTab() {
  const [days, setDays] = useState(90);
  const { data, isLoading } = useNetWorthHistory(days);

  if (isLoading) return <CardSkeleton />;
  const snapshots = data ?? [];

  const latest = snapshots.length > 0 ? snapshots[snapshots.length - 1] : null;
  const earliest = snapshots.length > 1 ? snapshots[0] : null;
  const change = latest && earliest ? latest.total_value - earliest.total_value : 0;
  const changePct = earliest && earliest.total_value > 0
    ? (change / earliest.total_value) * 100 : 0;

  const chartData = snapshots.map(s => ({
    date: formatDate(s.date),
    total: s.total_value,
    accounts: s.accounts_value,
    investments: s.investments_value,
    crypto: s.crypto_value,
  }));

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Select value={String(days)} onChange={(e) => setDays(Number(e.target.value))}>
          <option value="30">30 days</option>
          <option value="90">90 days</option>
          <option value="180">6 months</option>
          <option value="365">1 year</option>
        </Select>
      </div>

      {latest && (
        <div className="grid gap-4 sm:grid-cols-4">
          <StatCard label="Current Net Worth" value={formatCurrency(latest.total_value)} />
          <StatCard
            label={`${days}-Day Change`}
            value={formatCurrency(change)}
            sub={`${changePct >= 0 ? '+' : ''}${changePct.toFixed(2)}%`}
            positive={change >= 0}
          />
          <StatCard label="Accounts" value={formatCurrency(latest.accounts_value)} />
          <StatCard label="Investments + Crypto" value={formatCurrency(latest.investments_value + latest.crypto_value)} />
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Net Worth Over Time</CardTitle>
        </CardHeader>
        <CardContent>
          {chartData.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">No data yet. Snapshots are taken daily.</p>
          ) : (
            <ResponsiveContainer width="100%" height={350}>
              <AreaChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                <XAxis dataKey="date" tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} />
                <YAxis tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`} />
                <Tooltip
                  contentStyle={{ background: 'hsl(var(--card))', border: '1px solid hsl(var(--border))', borderRadius: 8 }}
                  formatter={tooltipFmt}
                />
                <Legend />
                <Area type="monotone" dataKey="accounts" stackId="1" stroke="#3b82f6" fill="#3b82f6" fillOpacity={0.3} name="Accounts" />
                <Area type="monotone" dataKey="investments" stackId="1" stroke="#8b5cf6" fill="#8b5cf6" fillOpacity={0.3} name="Investments" />
                <Area type="monotone" dataKey="crypto" stackId="1" stroke="#f59e0b" fill="#f59e0b" fillOpacity={0.3} name="Crypto" />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
