import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { CheckCircle, Circle, Wallet, ArrowLeftRight, PieChart, BarChart3, X } from 'lucide-react';
import { useUpdatePreferences } from '@/api/hooks';

const DISMISSED_KEY = 'onboarding-dismissed';

interface OnboardingChecklistProps {
  hasAccounts: boolean;
  hasTransactions: boolean;
  hasBudgets: boolean;
  serverDismissed?: boolean;
}

const steps = [
  { key: 'accounts', label: 'Create your first account', description: 'Add a checking, savings, or investment account', link: '/accounts', icon: Wallet },
  { key: 'transactions', label: 'Add a transaction', description: 'Record an income or expense', link: '/transactions', icon: ArrowLeftRight },
  { key: 'budgets', label: 'Set up a budget', description: 'Track your spending by category', link: '/budgets', icon: PieChart },
  { key: 'reports', label: 'Explore reports', description: 'See your financial analytics', link: '/reports', icon: BarChart3 },
] as const;

export function OnboardingChecklist({ hasAccounts, hasTransactions, hasBudgets, serverDismissed }: OnboardingChecklistProps) {
  const [dismissed, setDismissed] = useState(() => serverDismissed || localStorage.getItem(DISMISSED_KEY) === 'true');
  const updatePrefs = useUpdatePreferences();

  if (dismissed) return null;

  const completed: Record<string, boolean> = {
    accounts: hasAccounts,
    transactions: hasTransactions,
    budgets: hasBudgets,
    reports: false,
  };

  const completedCount = Object.values(completed).filter(Boolean).length;

  function handleDismiss() {
    localStorage.setItem(DISMISSED_KEY, 'true');
    setDismissed(true);
    updatePrefs.mutate({ onboarding_dismissed: true });
  }

  return (
    <Card className="border-primary/20 bg-primary/5">
      <CardHeader className="flex flex-row items-center justify-between pb-3">
        <div>
          <CardTitle className="text-lg">Getting Started</CardTitle>
          <p className="text-sm text-muted-foreground mt-1">
            {completedCount}/{steps.length} steps completed
          </p>
        </div>
        <Button variant="ghost" size="icon" onClick={handleDismiss} title="Dismiss">
          <X className="h-4 w-4" />
        </Button>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {steps.map((step) => {
            const done = completed[step.key];
            const Icon = step.icon;
            return (
              <Link
                key={step.key}
                to={step.link}
                className={`flex items-center gap-3 rounded-lg border p-3 transition-colors ${
                  done ? 'bg-muted/50 opacity-70' : 'hover:bg-accent/50'
                }`}
              >
                {done ? (
                  <CheckCircle className="h-5 w-5 shrink-0 text-success" />
                ) : (
                  <Circle className="h-5 w-5 shrink-0 text-muted-foreground" />
                )}
                <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
                <div className="min-w-0">
                  <p className={`text-sm font-medium ${done ? 'line-through text-muted-foreground' : ''}`}>{step.label}</p>
                  <p className="text-xs text-muted-foreground">{step.description}</p>
                </div>
              </Link>
            );
          })}
        </div>
        <Button variant="ghost" size="sm" className="mt-3 w-full" onClick={handleDismiss}>
          Got it, dismiss
        </Button>
      </CardContent>
    </Card>
  );
}
