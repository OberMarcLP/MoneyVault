import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useAlertRules, useCreateAlertRule, useToggleAlertRule, useDeleteAlertRule, useNotifications, useMarkAllRead, useClearAllNotifications } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { Bell, BellOff, Plus, Trash2, Check, AlertTriangle, TrendingUp, TrendingDown, DollarSign, Target } from 'lucide-react';
import type { AlertRuleType, Notification } from '@/types';
import { useNavigate } from 'react-router-dom';

type Tab = 'notifications' | 'rules';

export default function AlertsPage() {
  const [tab, setTab] = useState<Tab>('notifications');

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Notifications & Alerts</h1>
      </div>

      <div className="flex gap-2 border-b">
        {([
          { id: 'notifications', label: 'Notifications', icon: Bell },
          { id: 'rules', label: 'Alert Rules', icon: AlertTriangle },
        ] as const).map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            onClick={() => setTab(id)}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 transition-colors -mb-px ${
              tab === id ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <Icon className="h-4 w-4" /> {label}
          </button>
        ))}
      </div>

      {tab === 'notifications' && <NotificationsTab />}
      {tab === 'rules' && <AlertRulesTab />}
    </div>
  );
}

function NotificationsTab() {
  const { data } = useNotifications();
  const markAllRead = useMarkAllRead();
  const clearAll = useClearAllNotifications();
  const navigate = useNavigate();

  const notifications = data?.notifications ?? [];

  const typeIcon: Record<string, React.ReactNode> = {
    budget_alert: <DollarSign className="h-5 w-5 text-warning" />,
    price_alert: <TrendingUp className="h-5 w-5 text-primary" />,
    milestone: <Target className="h-5 w-5 text-success" />,
    info: <Bell className="h-5 w-5 text-muted-foreground" />,
    import_complete: <Check className="h-5 w-5 text-success" />,
    summary: <Bell className="h-5 w-5 text-primary" />,
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>All Notifications</CardTitle>
        <div className="flex gap-2">
          {notifications.some(n => !n.is_read) && (
            <Button variant="outline" size="sm" onClick={() => markAllRead.mutate()}>
              <Check className="h-3 w-3 mr-1" /> Mark All Read
            </Button>
          )}
          {notifications.length > 0 && (
            <Button variant="outline" size="sm" onClick={() => clearAll.mutate()}>
              <Trash2 className="h-3 w-3 mr-1" /> Clear All
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {notifications.length === 0 ? (
          <div className="py-12 text-center">
            <BellOff className="h-12 w-12 mx-auto text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">No notifications yet</p>
            <p className="text-sm text-muted-foreground mt-1">Set up alert rules to get notified about budget limits, price changes, and milestones.</p>
          </div>
        ) : (
          <div className="divide-y">
            {notifications.map((n: Notification) => (
              <div
                key={n.id}
                className={`flex items-start gap-4 py-4 cursor-pointer hover:bg-muted/30 rounded-lg px-3 -mx-3 transition-colors ${
                  !n.is_read ? 'bg-primary/5' : ''
                }`}
                onClick={() => n.link && navigate(n.link)}
              >
                <div className="shrink-0 mt-0.5">{typeIcon[n.type] || <Bell className="h-5 w-5" />}</div>
                <div className="flex-1 min-w-0">
                  <p className={`text-sm ${!n.is_read ? 'font-semibold' : ''}`}>{n.title}</p>
                  {n.message && <p className="text-sm text-muted-foreground mt-0.5">{n.message}</p>}
                  <p className="text-xs text-muted-foreground mt-1">{new Date(n.created_at).toLocaleString()}</p>
                </div>
                {!n.is_read && <div className="h-2 w-2 rounded-full bg-primary shrink-0 mt-2" />}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function AlertRulesTab() {
  const { data } = useAlertRules();
  const toggleRule = useToggleAlertRule();
  const deleteRule = useDeleteAlertRule();
  const [showForm, setShowForm] = useState(false);

  const rules = data?.rules ?? [];

  const ruleTypeLabel: Record<string, string> = {
    budget_overspend: 'Budget Overspend',
    price_drop: 'Price Drop',
    price_rise: 'Price Rise',
    net_worth_milestone: 'Net Worth Milestone',
  };

  const ruleTypeIcon: Record<string, React.ReactNode> = {
    budget_overspend: <DollarSign className="h-5 w-5 text-warning" />,
    price_drop: <TrendingDown className="h-5 w-5 text-destructive" />,
    price_rise: <TrendingUp className="h-5 w-5 text-success" />,
    net_worth_milestone: <Target className="h-5 w-5 text-primary" />,
  };

  function describeCondition(type: string, condition: Record<string, unknown>) {
    switch (type) {
      case 'budget_overspend':
        return `Alert when spending reaches ${condition.threshold ?? 100}% of budget`;
      case 'price_drop':
        return `Alert when ${condition.symbol ?? '?'} drops below $${condition.price ?? 0}`;
      case 'price_rise':
        return `Alert when ${condition.symbol ?? '?'} rises above $${condition.price ?? 0}`;
      case 'net_worth_milestone':
        return `Alert when net worth reaches $${Number(condition.amount ?? 0).toLocaleString()}`;
      default:
        return JSON.stringify(condition);
    }
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Alert Rules</CardTitle>
        <Button size="sm" onClick={() => setShowForm(!showForm)}>
          <Plus className="h-4 w-4 mr-1" /> New Rule
        </Button>
      </CardHeader>
      <CardContent>
        {showForm && <CreateRuleForm onClose={() => setShowForm(false)} />}

        {rules.length === 0 && !showForm ? (
          <div className="py-12 text-center">
            <AlertTriangle className="h-12 w-12 mx-auto text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">No alert rules configured</p>
            <p className="text-sm text-muted-foreground mt-1">Create rules to get notified about budget limits, price changes, and net worth milestones.</p>
          </div>
        ) : (
          <div className="divide-y mt-4">
            {rules.map(rule => (
              <div key={rule.id} className="flex items-center gap-4 py-4">
                <div className="shrink-0">{ruleTypeIcon[rule.type] || <Bell className="h-5 w-5" />}</div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium">{ruleTypeLabel[rule.type] || rule.type}</p>
                  <p className="text-sm text-muted-foreground">{describeCondition(rule.type, rule.condition)}</p>
                  {rule.last_triggered && (
                    <p className="text-xs text-muted-foreground mt-1">
                      Last triggered: {new Date(rule.last_triggered).toLocaleString()}
                    </p>
                  )}
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <Button
                    variant={rule.is_active ? 'outline' : 'secondary'}
                    size="sm"
                    onClick={() => toggleRule.mutate(rule.id)}
                  >
                    {rule.is_active ? <Bell className="h-3 w-3" /> : <BellOff className="h-3 w-3" />}
                    <span className="ml-1">{rule.is_active ? 'Active' : 'Paused'}</span>
                  </Button>
                  <Button variant="ghost" size="sm" onClick={() => deleteRule.mutate(rule.id)}>
                    <Trash2 className="h-3 w-3 text-destructive" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function CreateRuleForm({ onClose }: { onClose: () => void }) {
  const createRule = useCreateAlertRule();
  const { toast } = useToast();
  const [type, setType] = useState<AlertRuleType>('budget_overspend');
  const [threshold, setThreshold] = useState('80');
  const [symbol, setSymbol] = useState('');
  const [price, setPrice] = useState('');
  const [amount, setAmount] = useState('');

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    let condition: Record<string, unknown> = {};
    switch (type) {
      case 'budget_overspend':
        condition = { threshold: parseFloat(threshold) || 100 };
        break;
      case 'price_drop':
      case 'price_rise':
        if (!symbol || !price) { toast('Symbol and price are required'); return; }
        condition = { symbol: symbol.toUpperCase(), direction: type === 'price_rise' ? 'above' : 'below', price: parseFloat(price) };
        break;
      case 'net_worth_milestone':
        if (!amount) { toast('Amount is required'); return; }
        condition = { amount: parseFloat(amount) };
        break;
    }
    createRule.mutate({ type, condition }, {
      onSuccess: () => { toast('Alert rule created', 'success'); onClose(); },
      onError: () => toast('Failed to create rule', 'error'),
    });
  }

  return (
    <form onSubmit={handleSubmit} className="border rounded-lg p-4 mb-4 space-y-4 bg-muted/30">
      <div>
        <label className="block text-sm font-medium mb-1">Rule Type</label>
        <select
          className="w-full rounded-md border bg-background px-3 py-2 text-sm"
          value={type}
          onChange={e => setType(e.target.value as AlertRuleType)}
        >
          <option value="budget_overspend">Budget Overspend</option>
          <option value="price_drop">Price Drop</option>
          <option value="price_rise">Price Rise</option>
          <option value="net_worth_milestone">Net Worth Milestone</option>
        </select>
      </div>

      {type === 'budget_overspend' && (
        <div>
          <label className="block text-sm font-medium mb-1">Threshold (%)</label>
          <input
            type="number" min="1" max="200" step="1"
            className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            value={threshold} onChange={e => setThreshold(e.target.value)}
            placeholder="80"
          />
          <p className="text-xs text-muted-foreground mt-1">Get alerted when any budget is this % spent</p>
        </div>
      )}

      {(type === 'price_drop' || type === 'price_rise') && (
        <>
          <div>
            <label className="block text-sm font-medium mb-1">Symbol</label>
            <input
              type="text"
              className="w-full rounded-md border bg-background px-3 py-2 text-sm"
              value={symbol} onChange={e => setSymbol(e.target.value)}
              placeholder="AAPL, BTC, etc."
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Target Price ($)</label>
            <input
              type="number" min="0" step="0.01"
              className="w-full rounded-md border bg-background px-3 py-2 text-sm"
              value={price} onChange={e => setPrice(e.target.value)}
              placeholder="150.00"
            />
            <p className="text-xs text-muted-foreground mt-1">
              {type === 'price_rise' ? 'Alert when price rises above this' : 'Alert when price drops below this'}
            </p>
          </div>
        </>
      )}

      {type === 'net_worth_milestone' && (
        <div>
          <label className="block text-sm font-medium mb-1">Target Amount ($)</label>
          <input
            type="number" min="0" step="1000"
            className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            value={amount} onChange={e => setAmount(e.target.value)}
            placeholder="100000"
          />
          <p className="text-xs text-muted-foreground mt-1">Get alerted when your net worth reaches this amount</p>
        </div>
      )}

      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" size="sm" onClick={onClose}>Cancel</Button>
        <Button type="submit" size="sm" disabled={createRule.isPending}>Create Rule</Button>
      </div>
    </form>
  );
}
