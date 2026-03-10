import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { CardSkeleton } from '@/components/ui/skeleton';
import { useAssetAllocation } from '@/api/hooks';
import { formatCurrency } from '@/lib/utils';
import {
  PieChart, Pie, Cell, Tooltip, ResponsiveContainer,
} from 'recharts';
import { PieChart as PieIcon } from 'lucide-react';
import { StatCard, tooltipFmt, CHART_COLORS } from './report-utils';

export default function AllocationTab() {
  const { data, isLoading } = useAssetAllocation();

  if (isLoading) return <CardSkeleton />;
  const allocations = data ?? [];

  if (allocations.length === 0) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <PieIcon className="h-12 w-12 mx-auto text-muted-foreground mb-3" />
          <p className="text-muted-foreground">No investment holdings to analyze</p>
        </CardContent>
      </Card>
    );
  }

  const totalValue = allocations.reduce((s, a) => s + a.value, 0);

  const typeLabels: Record<string, string> = {
    stock: 'Stocks', etf: 'ETFs', crypto: 'Crypto',
    mutual_fund: 'Mutual Funds', defi_position: 'DeFi',
  };

  const pieData = allocations.map((a, i) => ({
    name: typeLabels[a.asset_type] || a.asset_type,
    value: a.value,
    color: CHART_COLORS[i % CHART_COLORS.length],
    count: a.count,
    percentage: a.percentage,
  }));

  return (
    <div className="space-y-4">
      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label="Total Portfolio" value={formatCurrency(totalValue)} />
        <StatCard label="Asset Types" value={String(allocations.length)} />
        <StatCard label="Total Holdings" value={String(allocations.reduce((s, a) => s + a.count, 0))} />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle className="text-lg">Asset Allocation</CardTitle></CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={100} innerRadius={50}
                  label={({ name, percent }: { name?: string; percent?: number }) => `${name ?? ''} ${((percent ?? 0) * 100).toFixed(1)}%`}>
                  {pieData.map((entry, i) => <Cell key={i} fill={entry.color} />)}
                </Pie>
                <Tooltip formatter={tooltipFmt} />
              </PieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle className="text-lg">Breakdown</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            {pieData.map((a, i) => (
              <div key={i} className="space-y-1.5">
                <div className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-2">
                    <div className="h-3 w-3 rounded-full" style={{ backgroundColor: a.color }} />
                    <span className="font-medium">{a.name}</span>
                    <span className="text-xs text-muted-foreground">{a.count} holdings</span>
                  </div>
                  <div className="text-right">
                    <span className="font-medium">{formatCurrency(a.value)}</span>
                    <span className="text-xs text-muted-foreground ml-2">{a.percentage.toFixed(1)}%</span>
                  </div>
                </div>
                <div className="h-2 w-full rounded-full bg-muted overflow-hidden">
                  <div className="h-full rounded-full" style={{ width: `${a.percentage}%`, backgroundColor: a.color }} />
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
