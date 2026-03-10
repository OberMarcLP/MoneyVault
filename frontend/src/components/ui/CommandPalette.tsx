import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Input } from '@/components/ui/input';
import {
  LayoutDashboard, ArrowLeftRight, Wallet, Tag, PieChart,
  Repeat, TrendingUp, Bitcoin, BarChart3, Settings, Upload,
  Bell, Search,
} from 'lucide-react';

interface CommandPaletteProps {
  open: boolean;
  onClose: () => void;
}

const PAGES = [
  { name: 'Dashboard', path: '/', icon: LayoutDashboard, keywords: 'home overview' },
  { name: 'Transactions', path: '/transactions', icon: ArrowLeftRight, keywords: 'payments expenses income' },
  { name: 'Accounts', path: '/accounts', icon: Wallet, keywords: 'bank checking savings' },
  { name: 'Categories', path: '/categories', icon: Tag, keywords: 'labels tags' },
  { name: 'Budgets', path: '/budgets', icon: PieChart, keywords: 'spending limits' },
  { name: 'Recurring', path: '/recurring', icon: Repeat, keywords: 'subscriptions bills' },
  { name: 'Investments', path: '/investments', icon: TrendingUp, keywords: 'stocks etf portfolio' },
  { name: 'Crypto', path: '/crypto', icon: Bitcoin, keywords: 'bitcoin ethereum wallet defi' },
  { name: 'Reports', path: '/reports', icon: BarChart3, keywords: 'analytics charts trends' },
  { name: 'Import', path: '/import', icon: Upload, keywords: 'csv upload' },
  { name: 'Notifications', path: '/notifications', icon: Bell, keywords: 'alerts' },
  { name: 'Settings', path: '/settings', icon: Settings, keywords: 'preferences profile' },
];

export function CommandPalette({ open, onClose }: CommandPaletteProps) {
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const filtered = query
    ? PAGES.filter((p) => {
        const q = query.toLowerCase();
        return p.name.toLowerCase().includes(q) || p.keywords.includes(q);
      })
    : PAGES;

  useEffect(() => {
    if (open) {
      setQuery('');
      setSelectedIndex(0);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [open]);

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((i) => Math.min(i + 1, filtered.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === 'Enter' && filtered[selectedIndex]) {
      navigate(filtered[selectedIndex].path);
      onClose();
    } else if (e.key === 'Escape') {
      onClose();
    }
  }

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]" onClick={onClose}>
      <div className="fixed inset-0 bg-black/50" />
      <div
        className="relative w-full max-w-lg rounded-xl border bg-card shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-2 border-b px-4">
          <Search className="h-4 w-4 text-muted-foreground shrink-0" />
          <Input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search pages..."
            className="border-0 focus-visible:ring-0 focus-visible:ring-offset-0"
          />
          <kbd className="hidden sm:inline-flex items-center gap-1 rounded border bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
            ESC
          </kbd>
        </div>
        <div className="max-h-80 overflow-y-auto p-2">
          {filtered.length === 0 ? (
            <p className="py-6 text-center text-sm text-muted-foreground">No results found</p>
          ) : (
            filtered.map((page, i) => {
              const Icon = page.icon;
              return (
                <button
                  key={page.path}
                  className={`flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm transition-colors ${
                    i === selectedIndex ? 'bg-accent text-accent-foreground' : 'text-foreground hover:bg-accent/50'
                  }`}
                  onClick={() => { navigate(page.path); onClose(); }}
                  onMouseEnter={() => setSelectedIndex(i)}
                >
                  <Icon className="h-4 w-4 text-muted-foreground" />
                  <span>{page.name}</span>
                </button>
              );
            })
          )}
        </div>
        <div className="border-t px-4 py-2 text-xs text-muted-foreground flex gap-3">
          <span><kbd className="rounded border bg-muted px-1">↑↓</kbd> Navigate</span>
          <span><kbd className="rounded border bg-muted px-1">↵</kbd> Open</span>
          <span><kbd className="rounded border bg-muted px-1">Esc</kbd> Close</span>
        </div>
      </div>
    </div>
  );
}
