import { lazy, Suspense, useState } from 'react';
import {
  TrendingUp, PieChart as PieIcon, BarChart3,
  DollarSign, Target, Wallet,
} from 'lucide-react';

type Tab = 'networth' | 'spending' | 'trends' | 'budget' | 'cashflow' | 'allocation';

const NetWorthTab = lazy(() => import('./NetWorthTab'));
const SpendingTab = lazy(() => import('./SpendingTab'));
const TrendsTab = lazy(() => import('./TrendsTab'));
const BudgetVsActualTab = lazy(() => import('./BudgetVsActualTab'));
const CashFlowTab = lazy(() => import('./CashFlowTab'));
const AllocationTab = lazy(() => import('./AllocationTab'));

function TabFallback() {
  return (
    <div className="flex items-center justify-center py-12">
      <div className="text-muted-foreground text-sm">Loading...</div>
    </div>
  );
}

export function ReportsPage() {
  const [tab, setTab] = useState<Tab>('networth');

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Reports & Analytics</h1>
        <p className="text-muted-foreground">Comprehensive financial insights and trends</p>
      </div>

      <div className="flex gap-1 border-b overflow-x-auto">
        {([
          { id: 'networth', label: 'Net Worth', icon: TrendingUp },
          { id: 'spending', label: 'Spending', icon: PieIcon },
          { id: 'trends', label: 'Trends', icon: BarChart3 },
          { id: 'budget', label: 'Budget vs Actual', icon: Target },
          { id: 'cashflow', label: 'Cash Flow', icon: DollarSign },
          { id: 'allocation', label: 'Allocation', icon: Wallet },
        ] as { id: Tab; label: string; icon: React.ElementType }[]).map(t => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${
              tab === t.id ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <t.icon className="h-4 w-4" />
            {t.label}
          </button>
        ))}
      </div>

      <Suspense fallback={<TabFallback />}>
        {tab === 'networth' && <NetWorthTab />}
        {tab === 'spending' && <SpendingTab />}
        {tab === 'trends' && <TrendsTab />}
        {tab === 'budget' && <BudgetVsActualTab />}
        {tab === 'cashflow' && <CashFlowTab />}
        {tab === 'allocation' && <AllocationTab />}
      </Suspense>
    </div>
  );
}
