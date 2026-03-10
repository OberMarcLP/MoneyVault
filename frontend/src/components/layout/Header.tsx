import { useState, useRef, useEffect } from 'react';
import { Menu, Moon, Sun, Monitor, LogOut, Bell, Check, X, AlertTriangle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useThemeStore } from '@/stores/theme';
import { useLogout, useUnreadCount, useNotifications, useMarkRead, useMarkAllRead, useDeleteNotification, useVerifyEmail } from '@/api/hooks';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';
import type { Notification } from '@/types';

interface HeaderProps {
  onMenuClick: () => void;
}

export function Header({ onMenuClick }: HeaderProps) {
  const { theme, setTheme } = useThemeStore();
  const logout = useLogout();
  const navigate = useNavigate();
  const { data: countData } = useUnreadCount();
  const [showNotifs, setShowNotifs] = useState(false);
  const user = useAuthStore((s) => s.user);
  const setUser = useAuthStore((s) => s.setUser);
  const verifyEmail = useVerifyEmail();

  const unread = countData?.count ?? 0;

  function cycleTheme() {
    const next = theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light';
    setTheme(next);
  }

  function handleLogout() {
    logout.mutate(undefined, {
      onSettled: () => navigate('/login'),
    });
  }

  const ThemeIcon = theme === 'light' ? Sun : theme === 'dark' ? Moon : Monitor;

  return (
    <>
      <header className="sticky top-0 z-30 flex h-16 items-center gap-4 border-b bg-background/95 backdrop-blur px-4 lg:px-6">
        <Button variant="ghost" size="icon" onClick={onMenuClick} className="lg:hidden">
          <Menu className="h-5 w-5" />
        </Button>

        <div className="flex-1" />

        <div className="relative">
          <Button variant="ghost" size="icon" onClick={() => setShowNotifs(!showNotifs)} title="Notifications">
            <Bell className="h-5 w-5" />
            {unread > 0 && (
              <span className="absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-destructive text-[10px] font-bold text-destructive-foreground px-1">
                {unread > 99 ? '99+' : unread}
              </span>
            )}
          </Button>
          {showNotifs && <NotificationDropdown onClose={() => setShowNotifs(false)} />}
        </div>

        <Button variant="ghost" size="icon" onClick={cycleTheme} title={`Theme: ${theme}`}>
          <ThemeIcon className="h-5 w-5" />
        </Button>

        <Button variant="ghost" size="icon" onClick={handleLogout} title="Logout">
          <LogOut className="h-5 w-5" />
        </Button>
      </header>

      {user && !user.email_verified && (
        <div className="flex items-center gap-2 bg-amber-50 dark:bg-amber-950 border-b border-amber-200 dark:border-amber-800 px-4 py-2 text-sm text-amber-800 dark:text-amber-200">
          <AlertTriangle className="h-4 w-4 shrink-0" />
          <span>Your email is not verified.</span>
          <Button
            variant="ghost"
            size="sm"
            className="h-6 text-xs underline px-1"
            onClick={() => verifyEmail.mutate(undefined, {
              onSuccess: () => {
                if (user) setUser({ ...user, email_verified: true });
              },
            })}
            disabled={verifyEmail.isPending}
          >
            {verifyEmail.isPending ? 'Verifying...' : 'Verify now'}
          </Button>
        </div>
      )}
    </>
  );
}

function NotificationDropdown({ onClose }: { onClose: () => void }) {
  const { data } = useNotifications();
  const markRead = useMarkRead();
  const markAllRead = useMarkAllRead();
  const deleteNotif = useDeleteNotification();
  const navigate = useNavigate();
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose();
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [onClose]);

  const notifications = data?.notifications ?? [];

  const typeIcon: Record<string, string> = {
    budget_alert: '💰',
    price_alert: '📈',
    milestone: '🎯',
    info: 'ℹ️',
    import_complete: '✅',
    summary: '📊',
  };

  return (
    <div ref={ref} className="absolute right-0 top-full mt-2 w-96 max-h-[28rem] rounded-lg border bg-card shadow-xl overflow-hidden z-50">
      <div className="flex items-center justify-between border-b px-4 py-3">
        <h3 className="font-semibold text-sm">Notifications</h3>
        <div className="flex items-center gap-1">
          {notifications.some(n => !n.is_read) && (
            <Button variant="ghost" size="sm" className="text-xs h-7" onClick={() => markAllRead.mutate()}>
              <Check className="h-3 w-3 mr-1" /> Read All
            </Button>
          )}
          <Button variant="ghost" size="sm" className="h-7" onClick={() => { onClose(); navigate('/alerts'); }}>
            Settings
          </Button>
        </div>
      </div>
      <div className="overflow-y-auto max-h-80">
        {notifications.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">No notifications</div>
        ) : (
          notifications.map((n: Notification) => (
            <div
              key={n.id}
              className={`flex items-start gap-3 px-4 py-3 border-b last:border-0 hover:bg-muted/30 transition-colors cursor-pointer ${
                !n.is_read ? 'bg-primary/5' : ''
              }`}
              onClick={() => {
                if (!n.is_read) markRead.mutate(n.id);
                if (n.link) { navigate(n.link); onClose(); }
              }}
            >
              <span className="text-lg mt-0.5 shrink-0">{typeIcon[n.type] || 'ℹ️'}</span>
              <div className="flex-1 min-w-0">
                <p className={`text-sm ${!n.is_read ? 'font-semibold' : 'font-medium'}`}>{n.title}</p>
                {n.message && <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">{n.message}</p>}
                <p className="text-[10px] text-muted-foreground mt-1">{new Date(n.created_at).toLocaleString()}</p>
              </div>
              <Button
                variant="ghost" size="sm" className="shrink-0 h-6 w-6 p-0 opacity-50 hover:opacity-100"
                onClick={(e) => { e.stopPropagation(); deleteNotif.mutate(n.id); }}
              >
                <X className="h-3 w-3" />
              </Button>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
