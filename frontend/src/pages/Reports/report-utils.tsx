import { Card, CardContent } from '@/components/ui/card';
import { formatCurrency } from '@/lib/utils';

// eslint-disable-next-line react-refresh/only-export-components
export const tooltipFmt = (value: unknown) => [formatCurrency(Number(value ?? 0)), ''];

// eslint-disable-next-line react-refresh/only-export-components
export const CHART_COLORS = [
  '#6366f1', '#8b5cf6', '#a855f7', '#d946ef', '#ec4899',
  '#f43f5e', '#ef4444', '#f97316', '#eab308', '#22c55e',
  '#14b8a6', '#06b6d4', '#3b82f6', '#6366f1',
];

export function StatCard({ label, value, sub, positive }: {
  label: string; value: string; sub?: string; positive?: boolean;
}) {
  return (
    <Card>
      <CardContent className="p-4">
        <p className="text-xs text-muted-foreground mb-1">{label}</p>
        <p className="text-xl font-bold">{value}</p>
        {sub && (
          <p className={`text-xs mt-1 ${positive !== undefined ? (positive ? 'text-success' : 'text-destructive') : 'text-muted-foreground'}`}>
            {sub}
          </p>
        )}
      </CardContent>
    </Card>
  );
}
