import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Download, X } from 'lucide-react';

interface BeforeInstallPromptEvent extends Event {
  prompt(): Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

const PWA_DISMISSED_KEY = 'pwa-dismissed-at';
const PWA_INSTALLED_KEY = 'pwa-installed';
const DISMISS_DAYS = 30;

function isDismissed(): boolean {
  if (localStorage.getItem(PWA_INSTALLED_KEY)) return true;
  const dismissed = localStorage.getItem(PWA_DISMISSED_KEY);
  if (!dismissed) return false;
  const elapsed = Date.now() - parseInt(dismissed, 10);
  return elapsed < DISMISS_DAYS * 24 * 60 * 60 * 1000;
}

export function PWAInstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [hidden, setHidden] = useState(isDismissed());

  useEffect(() => {
    const handler = (e: Event) => {
      e.preventDefault();
      setDeferredPrompt(e as BeforeInstallPromptEvent);
    };
    window.addEventListener('beforeinstallprompt', handler);
    return () => window.removeEventListener('beforeinstallprompt', handler);
  }, []);

  if (!deferredPrompt || hidden) return null;

  function dismiss() {
    localStorage.setItem(PWA_DISMISSED_KEY, String(Date.now()));
    setHidden(true);
  }

  async function handleInstall() {
    if (!deferredPrompt) return;
    await deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    if (outcome === 'accepted') {
      localStorage.setItem(PWA_INSTALLED_KEY, 'true');
      setDeferredPrompt(null);
      setHidden(true);
    }
  }

  return (
    <div className="fixed bottom-4 left-4 right-4 sm:left-auto sm:right-4 sm:w-80 z-50 rounded-lg border bg-card p-4 shadow-xl animate-in slide-in-from-bottom">
      <div className="flex items-start gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold text-sm">
          MV
        </div>
        <div className="flex-1">
          <p className="text-sm font-semibold">Install MoneyVault</p>
          <p className="text-xs text-muted-foreground mt-0.5">Add to your home screen for quick access and offline support.</p>
          <div className="flex gap-2 mt-3">
            <Button size="sm" onClick={handleInstall}>
              <Download className="h-3 w-3 mr-1" /> Install
            </Button>
            <Button variant="ghost" size="sm" onClick={dismiss}>
              Not now
            </Button>
          </div>
        </div>
        <button onClick={dismiss} className="text-muted-foreground hover:text-foreground">
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}
