import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select } from '@/components/ui/select';
import { CardSkeleton } from '@/components/ui/skeleton';
import { useCashFlowForecast } from '@/api/hooks';
import { formatCurrency, formatDate } from '@/lib/utils';
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, Legend,
} from 'recharts';
import { Calendar } from 'lucide-react';
import { StatCard, tooltipFmt } from './report-utils';

export default function CashFlowTab() {
  const [months, setMonths] = useState(6);
  const { data, isLoading } = useCashFlowForecast(months);

  if (isLoading) return <CardSkeleton />;
  if (!data) return <p className="text-muted-foreground text-center py-8">No data available</p>;

  const { forecast, runway, upcoming_bills } = data;

  const chartData = forecast.map(f => ({
    period: f.period,
    income: f.projected_income,
    expense: f.projected_expense,
    balance: f.running_balance,
  }));

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Select value={String(months)} onChange={(e) => setMonths(Number(e.target.value))}>
          <option value="3">3 months</option>
          <option value="6">6 months</option>
          <option value="12">12 months</option>
        </Select>
      </div>

      <div className="grid gap-4 sm:grid-cols-4">
        <StatCard label="Current Balance" value={formatCurrency(runway.current_balance)} />
        <StatCard label="Monthly Income" value={formatCurrency(runway.monthly_income)} positive />
        <StatCard label="Monthly Expenses" value={formatCurrency(runway.monthly_expenses)} />
        <StatCard
          label="Runway"
          value={runway.runway_months < 0 ? '\u221E' : `${runway.runway_months.toFixed(1)} mo`}
          sub={runway.monthly_savings >= 0 ? 'Savings positive' : 'Burning savings'}
          positive={runway.monthly_savings >= 0}
        />
      </div>

      <Card>
        <CardHeader><CardTitle className="text-lg">Projected Balance</CardTitle></CardHeader>
        <CardContent>
          {chartData.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">No forecast data</p>
          ) : (
            <ResponsiveContainer width="100%" height={350}>
              <AreaChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                <XAxis dataKey="period" tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} />
                <YAxis tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`} />
                <Tooltip
                  contentStyle={{ background: 'hsl(var(--card))', border: '1px solid hsl(var(--border))', borderRadius: 8 }}
                  formatter={tooltipFmt}
                />
                <Legend />
                <Area type="monotone" dataKey="balance" stroke="#6366f1" fill="#6366f1" fillOpacity={0.2} name="Balance" />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </CardContent>
      </Card>

      {upcoming_bills.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Calendar className="h-5 w-5" /> Upcoming Bills
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {upcoming_bills.map((bill, i) => (
                <div key={i} className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <p className="font-medium text-sm">{bill.description || 'Unnamed bill'}</p>
                    <p className="text-xs text-muted-foreground">{bill.frequency} &middot; {bill.account_name}</p>
                  </div>
                  <div className="text-right">
                    <p className="font-medium text-sm text-destructive">{formatCurrency(bill.amount)}</p>
                    <p className="text-xs text-muted-foreground">Due {formatDate(bill.due_date)}</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
