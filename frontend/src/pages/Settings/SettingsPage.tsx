import { useState, useEffect, useCallback, type FormEvent } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/ConfirmDialog';
import { useAuthStore } from '@/stores/auth';
import { useThemeStore } from '@/stores/theme';
import { useUpdatePreferences, useSetupTOTP, useVerifyTOTP, useDisableTOTP, useMe, useExportTransactions, useExportAccounts, useExportAll, useWebAuthnCredentials, useWebAuthnRegisterBegin, useWebAuthnRegisterFinish, useDeleteWebAuthnCredential } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { formatDate } from '@/lib/utils';
import { registerPasskey, isWebAuthnSupported } from '@/lib/webauthn';
import { Sun, Moon, Monitor, Shield, Lock, Palette, Globe, Download, Key, Trash2, Link2, RefreshCw, Power, Bell, ShieldCheck } from 'lucide-react';
import type { ExchangeName } from '@/types';
import { useExchangeConnections, useConnectExchange, useSyncExchange, useDeleteExchange, useToggleExchange, useVAPIDKey, usePushSubscribe, usePushUnsubscribe, useE2EExportData, useE2EEnable, useE2EDisable } from '@/api/hooks';
import { useCryptoStore } from '@/stores/crypto';
import { generateDEK, generateSalt, encryptDEK, encryptField, decryptField } from '@/utils/crypto';

export function SettingsPage() {
  const user = useAuthStore((s) => s.user);
  const setUser = useAuthStore((s) => s.setUser);
  const { theme, setTheme } = useThemeStore();
  const updatePreferences = useUpdatePreferences();
  const { toast } = useToast();
  const { refetch: refetchMe } = useMe();

  const [showTOTPSetup, setShowTOTPSetup] = useState(false);
  const [showDisable2FA, setShowDisable2FA] = useState(false);
  const [totpSecret, setTotpSecret] = useState('');
  const [totpUrl, setTotpUrl] = useState('');
  const [verifyCode, setVerifyCode] = useState('');

  const setupTOTP = useSetupTOTP();
  const verifyTOTP = useVerifyTOTP();
  const disableTOTP = useDisableTOTP();

  const themes = [
    { value: 'light' as const, label: 'Light', icon: Sun },
    { value: 'dark' as const, label: 'Dark', icon: Moon },
    { value: 'system' as const, label: 'System', icon: Monitor },
  ];

  function handleThemeChange(newTheme: 'light' | 'dark' | 'system') {
    setTheme(newTheme);
    updatePreferences.mutate({ theme: newTheme });
  }

  function handleCurrencyChange(currency: string) {
    updatePreferences.mutate(
      { currency },
      { onSuccess: () => toast('Currency updated', 'success') },
    );
  }

  function handleEnableTOTP() {
    setupTOTP.mutate(undefined, {
      onSuccess: (data) => {
        setTotpSecret(data.secret);
        setTotpUrl(data.url);
        setShowTOTPSetup(true);
      },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  function handleVerifyTOTP(e: FormEvent) {
    e.preventDefault();
    verifyTOTP.mutate(verifyCode, {
      onSuccess: async () => {
        setShowTOTPSetup(false);
        setVerifyCode('');
        toast('Two-factor authentication enabled', 'success');
        const res = await refetchMe();
        if (res.data?.user) setUser(res.data.user);
      },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  function handleDisableTOTP() {
    disableTOTP.mutate(undefined, {
      onSuccess: async () => {
        setShowDisable2FA(false);
        toast('Two-factor authentication disabled', 'success');
        const res = await refetchMe();
        if (res.data?.user) setUser(res.data.user);
      },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground">Manage your account and preferences</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Lock className="h-5 w-5" />
            Profile
          </CardTitle>
          <CardDescription>Your account information</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary/10 text-primary text-2xl font-bold">
              {user?.email?.charAt(0).toUpperCase()}
            </div>
            <div>
              <p className="font-medium">{user?.email}</p>
              <p className="text-sm text-muted-foreground capitalize">Role: {user?.role}</p>
              <p className="text-sm text-muted-foreground">
                Member since {user?.created_at ? formatDate(user.created_at) : '...'}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Palette className="h-5 w-5" />
            Appearance
          </CardTitle>
          <CardDescription>Choose your preferred theme</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-3">
            {themes.map(({ value, label, icon: Icon }) => (
              <button
                key={value}
                onClick={() => handleThemeChange(value)}
                className={`flex flex-col items-center gap-2 rounded-lg border-2 p-4 transition-colors ${
                  theme === value ? 'border-primary bg-primary/5' : 'border-border hover:bg-accent'
                }`}
              >
                <Icon className="h-6 w-6" />
                <span className="text-sm font-medium">{label}</span>
              </button>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Globe className="h-5 w-5" />
            Default Currency
          </CardTitle>
          <CardDescription>Used when creating new accounts and transactions</CardDescription>
        </CardHeader>
        <CardContent>
          <Select
            defaultValue={user?.preferences?.currency ?? 'USD'}
            onChange={(e) => handleCurrencyChange(e.target.value)}
            className="w-48"
          >
            <option value="USD">USD - US Dollar</option>
            <option value="EUR">EUR - Euro</option>
            <option value="GBP">GBP - British Pound</option>
            <option value="CHF">CHF - Swiss Franc</option>
            <option value="JPY">JPY - Japanese Yen</option>
            <option value="CAD">CAD - Canadian Dollar</option>
            <option value="AUD">AUD - Australian Dollar</option>
            <option value="NOK">NOK - Norwegian Krone</option>
            <option value="SEK">SEK - Swedish Krona</option>
          </Select>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Security
          </CardTitle>
          <CardDescription>Two-factor authentication</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div>
              <p className="font-medium">Two-Factor Authentication (TOTP)</p>
              <p className="text-sm text-muted-foreground">
                {user?.totp_enabled
                  ? 'Your account is protected with 2FA'
                  : 'Add an extra layer of security with an authenticator app'}
              </p>
            </div>
            {user?.totp_enabled ? (
              <Button
                variant="destructive"
                size="sm"
                onClick={() => setShowDisable2FA(true)}
                disabled={disableTOTP.isPending}
              >
                {disableTOTP.isPending ? 'Disabling...' : 'Disable 2FA'}
              </Button>
            ) : (
              <Button
                size="sm"
                onClick={handleEnableTOTP}
                disabled={setupTOTP.isPending}
              >
                {setupTOTP.isPending ? 'Setting up...' : 'Enable 2FA'}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      <PasskeysSection />

      <ExchangeConnectionsSection />

      <PushNotificationsSection />

      <E2EEncryptionSection />

      <Dialog open={showTOTPSetup} onClose={() => setShowTOTPSetup(false)}>
        <DialogHeader>
          <DialogTitle>Set Up Two-Factor Authentication</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Scan the QR code below with your authenticator app (Google Authenticator, Authy, etc.),
            or manually enter the secret key.
          </p>

          <div className="flex justify-center p-4 bg-white rounded-lg">
            <img
              src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(totpUrl)}`}
              alt="TOTP QR Code"
              className="w-48 h-48"
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium">Secret Key</label>
            <div className="flex gap-2">
              <code className="flex-1 rounded-md border bg-muted px-3 py-2 text-sm font-mono break-all">
                {totpSecret}
              </code>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  navigator.clipboard.writeText(totpSecret);
                  toast('Copied to clipboard', 'success');
                }}
              >
                Copy
              </Button>
            </div>
          </div>

          <form onSubmit={handleVerifyTOTP} className="space-y-3">
            <div className="space-y-1">
              <label className="text-sm font-medium">Verification Code</label>
              <Input
                type="text"
                placeholder="Enter 6-digit code"
                value={verifyCode}
                onChange={(e) => setVerifyCode(e.target.value)}
                maxLength={6}
                autoFocus
              />
              <p className="text-xs text-muted-foreground">
                Enter the 6-digit code from your authenticator app to verify setup
              </p>
            </div>
            <Button type="submit" className="w-full" disabled={verifyTOTP.isPending || verifyCode.length !== 6}>
              {verifyTOTP.isPending ? 'Verifying...' : 'Verify & Enable'}
            </Button>
          </form>
        </div>
      </Dialog>

      <DataExportSection />

      <ConfirmDialog
        open={showDisable2FA}
        onClose={() => setShowDisable2FA(false)}
        onConfirm={handleDisableTOTP}
        title="Disable Two-Factor Authentication"
        message="Are you sure you want to disable 2FA? This will reduce the security of your account."
        confirmLabel="Disable 2FA"
        variant="warning"
        isPending={disableTOTP.isPending}
      />
    </div>
  );
}

function PasskeysSection() {
  const { toast } = useToast();
  const { data, isLoading } = useWebAuthnCredentials();
  const registerBegin = useWebAuthnRegisterBegin();
  const registerFinish = useWebAuthnRegisterFinish();
  const deleteCredential = useDeleteWebAuthnCredential();
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);
  const [passkeyName, setPasskeyName] = useState('');
  const [isRegistering, setIsRegistering] = useState(false);

  const supported = isWebAuthnSupported();
  const credentials = data?.credentials ?? [];

  async function handleRegister() {
    setIsRegistering(true);
    try {
      const { options } = await registerBegin.mutateAsync();
      const credential = await registerPasskey(options as unknown as Record<string, unknown>);
      await registerFinish.mutateAsync({ credential, name: passkeyName || undefined });
      setPasskeyName('');
      toast('Passkey registered successfully', 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to register passkey', 'error');
    } finally {
      setIsRegistering(false);
    }
  }

  function handleDelete(id: string) {
    deleteCredential.mutate(id, {
      onSuccess: () => {
        setShowDeleteConfirm(null);
        toast('Passkey removed', 'success');
      },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Key className="h-5 w-5" />
          Passkeys
        </CardTitle>
        <CardDescription>Sign in without a password using biometrics or security keys</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {!supported && (
          <p className="text-sm text-muted-foreground">
            Your browser does not support passkeys.
          </p>
        )}

        {supported && (
          <>
            {isLoading ? (
              <p className="text-sm text-muted-foreground">Loading...</p>
            ) : credentials.length === 0 ? (
              <p className="text-sm text-muted-foreground">No passkeys registered yet.</p>
            ) : (
              <div className="space-y-2">
                {credentials.map((cred) => (
                  <div key={cred.id} className="flex items-center justify-between rounded-lg border p-3">
                    <div>
                      <p className="font-medium text-sm">{cred.name}</p>
                      <p className="text-xs text-muted-foreground">
                        Added {formatDate(cred.created_at)}
                        {cred.last_used_at && ` · Last used ${formatDate(cred.last_used_at)}`}
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setShowDeleteConfirm(cred.id)}
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                ))}
              </div>
            )}

            <div className="flex gap-2 items-end">
              <div className="flex-1">
                <label className="text-sm font-medium">Passkey Name</label>
                <Input
                  placeholder="e.g. MacBook Touch ID"
                  value={passkeyName}
                  onChange={(e) => setPasskeyName(e.target.value)}
                />
              </div>
              <Button
                onClick={handleRegister}
                disabled={isRegistering}
                loading={isRegistering}
              >
                {isRegistering ? 'Registering...' : 'Add Passkey'}
              </Button>
            </div>
          </>
        )}

        <ConfirmDialog
          open={!!showDeleteConfirm}
          onClose={() => setShowDeleteConfirm(null)}
          onConfirm={() => showDeleteConfirm && handleDelete(showDeleteConfirm)}
          title="Remove Passkey"
          message="Are you sure you want to remove this passkey? You won't be able to use it to sign in anymore."
          confirmLabel="Remove"
          variant="warning"
          isPending={deleteCredential.isPending}
        />
      </CardContent>
    </Card>
  );
}

function urlBase64ToUint8Array(base64String: string) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const raw = atob(base64);
  const arr = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i++) arr[i] = raw.charCodeAt(i);
  return arr;
}

function PushNotificationsSection() {
  const { toast } = useToast();
  const { data: vapidData } = useVAPIDKey();
  const pushSubscribe = usePushSubscribe();
  const pushUnsubscribe = usePushUnsubscribe();
  const [isSubscribed, setIsSubscribed] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [supported, setSupported] = useState(false);

  const checkSubscription = useCallback(async () => {
    if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
      setSupported(false);
      setIsLoading(false);
      return;
    }
    setSupported(true);
    try {
      const reg = await navigator.serviceWorker.ready;
      const sub = await reg.pushManager.getSubscription();
      setIsSubscribed(!!sub);
    } catch {
      // ignore
    }
    setIsLoading(false);
  }, []);

  useEffect(() => { checkSubscription(); }, [checkSubscription]);

  async function handleEnable() {
    if (!vapidData?.public_key) {
      toast('Push notifications not configured on server', 'error');
      return;
    }
    try {
      const permission = await Notification.requestPermission();
      if (permission !== 'granted') {
        toast('Notification permission denied', 'error');
        return;
      }
      const reg = await navigator.serviceWorker.ready;
      const sub = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(vapidData.public_key),
      });
      const json = sub.toJSON();
      await pushSubscribe.mutateAsync({
        endpoint: sub.endpoint,
        auth: json.keys?.auth ?? '',
        p256dh: json.keys?.p256dh ?? '',
      });
      setIsSubscribed(true);
      toast('Push notifications enabled', 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to enable push', 'error');
    }
  }

  async function handleDisable() {
    try {
      const reg = await navigator.serviceWorker.ready;
      const sub = await reg.pushManager.getSubscription();
      if (sub) {
        await pushUnsubscribe.mutateAsync(sub.endpoint);
        await sub.unsubscribe();
      }
      setIsSubscribed(false);
      toast('Push notifications disabled', 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to disable push', 'error');
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bell className="h-5 w-5" />
          Push Notifications
        </CardTitle>
        <CardDescription>Receive browser notifications for alerts even when the app is closed</CardDescription>
      </CardHeader>
      <CardContent>
        {!supported ? (
          <p className="text-sm text-muted-foreground">Your browser does not support push notifications.</p>
        ) : isLoading ? (
          <p className="text-sm text-muted-foreground">Checking...</p>
        ) : (
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div>
              <p className="font-medium">Browser Push Notifications</p>
              <p className="text-sm text-muted-foreground">
                {isSubscribed
                  ? 'You will receive push notifications for budget alerts, price alerts, and milestones'
                  : 'Enable to receive alerts even when MoneyVault is not open'}
              </p>
            </div>
            {isSubscribed ? (
              <Button variant="destructive" size="sm" onClick={handleDisable}>
                Disable
              </Button>
            ) : (
              <Button size="sm" onClick={handleEnable}>
                Enable
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

const EXCHANGE_INFO: Record<ExchangeName, { name: string; color: string }> = {
  binance: { name: 'Binance', color: 'bg-yellow-500' },
  coinbase: { name: 'Coinbase', color: 'bg-blue-500' },
  kraken: { name: 'Kraken', color: 'bg-purple-500' },
};

function ExchangeConnectionsSection() {
  const { toast } = useToast();
  const { data: connections = [], isLoading } = useExchangeConnections();
  const connectExchange = useConnectExchange();
  const syncExchange = useSyncExchange();
  const deleteExchange = useDeleteExchange();
  const toggleExchange = useToggleExchange();
  const [showConnect, setShowConnect] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [connectForm, setConnectForm] = useState({
    exchange: 'binance' as ExchangeName,
    api_key: '',
    api_secret: '',
    label: '',
  });

  function handleConnect(e: FormEvent) {
    e.preventDefault();
    connectExchange.mutate(connectForm, {
      onSuccess: () => {
        setShowConnect(false);
        setConnectForm({ exchange: 'binance', api_key: '', api_secret: '', label: '' });
        toast('Exchange connected successfully', 'success');
      },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  function handleSync(id: string) {
    setSyncingId(id);
    syncExchange.mutate(id, {
      onSuccess: (result) => {
        setSyncingId(null);
        const count = result.balances?.length ?? 0;
        toast(`Synced ${count} balance${count !== 1 ? 's' : ''} from ${result.exchange}`, 'success');
      },
      onError: (err) => {
        setSyncingId(null);
        toast(err.message, 'error');
      },
    });
  }

  function handleDelete(id: string) {
    deleteExchange.mutate(id, {
      onSuccess: () => {
        setShowDeleteConfirm(null);
        toast('Exchange disconnected', 'success');
      },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  function handleToggle(id: string) {
    toggleExchange.mutate(id, {
      onSuccess: () => toast('Exchange connection toggled', 'success'),
      onError: (err) => toast(err.message, 'error'),
    });
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Link2 className="h-5 w-5" />
              Connected Exchanges
            </CardTitle>
            <CardDescription>Connect crypto exchanges for read-only balance syncing</CardDescription>
          </div>
          <Button size="sm" onClick={() => setShowConnect(true)}>
            Connect Exchange
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {isLoading ? (
          <p className="text-sm text-muted-foreground">Loading...</p>
        ) : connections.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No exchanges connected yet. Connect an exchange to automatically sync your crypto balances.
          </p>
        ) : (
          connections.map((conn) => {
            const info = EXCHANGE_INFO[conn.exchange];
            return (
              <div key={conn.id} className="flex items-center justify-between rounded-lg border p-4">
                <div className="flex items-center gap-3">
                  <div className={`h-8 w-8 rounded-full ${info.color} flex items-center justify-center text-white text-xs font-bold`}>
                    {info.name.charAt(0)}
                  </div>
                  <div>
                    <p className="font-medium text-sm">
                      {info.name}
                      {conn.label && <span className="text-muted-foreground ml-1">({conn.label})</span>}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {conn.is_active ? (
                        conn.last_synced
                          ? `Last synced ${formatDate(conn.last_synced)}`
                          : 'Never synced'
                      ) : (
                        'Disabled'
                      )}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleSync(conn.id)}
                    disabled={syncingId === conn.id || !conn.is_active}
                    title="Sync balances"
                  >
                    <RefreshCw className={`h-4 w-4 ${syncingId === conn.id ? 'animate-spin' : ''}`} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleToggle(conn.id)}
                    title={conn.is_active ? 'Disable' : 'Enable'}
                  >
                    <Power className={`h-4 w-4 ${conn.is_active ? 'text-green-500' : 'text-muted-foreground'}`} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowDeleteConfirm(conn.id)}
                    title="Remove"
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </div>
            );
          })
        )}

        <Dialog open={showConnect} onClose={() => setShowConnect(false)}>
          <DialogHeader>
            <DialogTitle>Connect Exchange</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleConnect} className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Connect your exchange using read-only API keys. Your keys are encrypted and stored securely.
            </p>

            <div className="space-y-1">
              <label className="text-sm font-medium">Exchange</label>
              <Select
                value={connectForm.exchange}
                onChange={(e) => setConnectForm({ ...connectForm, exchange: e.target.value as ExchangeName })}
              >
                <option value="binance">Binance</option>
                <option value="coinbase">Coinbase</option>
                <option value="kraken">Kraken</option>
              </Select>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium">API Key</label>
              <Input
                type="password"
                placeholder="Enter your API key"
                value={connectForm.api_key}
                onChange={(e) => setConnectForm({ ...connectForm, api_key: e.target.value })}
                required
              />
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium">API Secret</label>
              <Input
                type="password"
                placeholder="Enter your API secret"
                value={connectForm.api_secret}
                onChange={(e) => setConnectForm({ ...connectForm, api_secret: e.target.value })}
                required
              />
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium">Label (optional)</label>
              <Input
                placeholder="e.g. Main account"
                value={connectForm.label}
                onChange={(e) => setConnectForm({ ...connectForm, label: e.target.value })}
              />
            </div>

            <div className="rounded-lg bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-800 p-3">
              <p className="text-sm text-amber-800 dark:text-amber-200">
                <strong>Security tip:</strong> Only use read-only API keys. Never use keys with withdrawal permissions.
              </p>
            </div>

            <Button
              type="submit"
              className="w-full"
              disabled={connectExchange.isPending || !connectForm.api_key || !connectForm.api_secret}
              loading={connectExchange.isPending}
            >
              {connectExchange.isPending ? 'Connecting...' : 'Connect Exchange'}
            </Button>
          </form>
        </Dialog>

        <ConfirmDialog
          open={!!showDeleteConfirm}
          onClose={() => setShowDeleteConfirm(null)}
          onConfirm={() => showDeleteConfirm && handleDelete(showDeleteConfirm)}
          title="Disconnect Exchange"
          message="Are you sure you want to disconnect this exchange? Your synced balances will be preserved but no longer auto-updated."
          confirmLabel="Disconnect"
          variant="warning"
          isPending={deleteExchange.isPending}
        />
      </CardContent>
    </Card>
  );
}

function E2EEncryptionSection() {
  const user = useAuthStore((s) => s.user);
  const setUser = useAuthStore((s) => s.setUser);
  const { dek, setDEK, clearDEK } = useCryptoStore();
  const { toast } = useToast();
  const { refetch: refetchMe } = useMe();
  const exportData = useE2EExportData();
  const enableE2E = useE2EEnable();
  const disableE2E = useE2EDisable();

  const [showEnableDialog, setShowEnableDialog] = useState(false);
  const [showDisableDialog, setShowDisableDialog] = useState(false);
  const [password, setPassword] = useState('');
  const [migrating, setMigrating] = useState(false);

  async function handleEnable(e: FormEvent) {
    e.preventDefault();
    if (!password) return;
    setMigrating(true);

    try {
      // 1. Get all plaintext data from server
      const data = await exportData.mutateAsync();

      // 2. Generate new DEK and salt
      const newDEK = await generateDEK();
      const salt = generateSalt();
      const encDEK = await encryptDEK(newDEK, password, salt);

      // 3. Encrypt all data with new DEK
      const encryptedAccounts = await Promise.all(
        (data.accounts || []).map(async (a) => ({
          id: a.id,
          name: await encryptField(a.name, newDEK),
          balance: await encryptField(a.balance, newDEK),
        })),
      );

      const encryptedTransactions = await Promise.all(
        (data.transactions || []).map(async (t) => ({
          id: t.id,
          amount: await encryptField(t.amount, newDEK),
          description: await encryptField(t.description, newDEK),
        })),
      );

      // 4. Send everything to server atomically
      await enableE2E.mutateAsync({
        password,
        e2e_encrypted_dek: encDEK,
        e2e_kek_salt: salt,
        data: {
          accounts: encryptedAccounts,
          transactions: encryptedTransactions,
        },
      });

      // 5. Store DEK in memory
      setDEK(newDEK);

      // 6. Refresh user data
      const res = await refetchMe();
      if (res.data?.user) setUser(res.data.user);

      setShowEnableDialog(false);
      setPassword('');
      toast('End-to-end encryption enabled', 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to enable E2E', 'error');
    } finally {
      setMigrating(false);
    }
  }

  async function handleDisable(e: FormEvent) {
    e.preventDefault();
    if (!password || !dek) return;
    setMigrating(true);

    try {
      // 1. Get encrypted data from server (we need to decrypt it with our DEK)
      const data = await exportData.mutateAsync();

      // 2. Decrypt with client DEK and send plaintext for server to re-encrypt
      const decryptedAccounts = await Promise.all(
        (data.accounts || []).map(async (a) => ({
          id: a.id,
          name: await decryptField(a.name, dek),
          balance: await decryptField(a.balance, dek),
        })),
      );

      const decryptedTransactions = await Promise.all(
        (data.transactions || []).map(async (t) => ({
          id: t.id,
          amount: await decryptField(t.amount, dek),
          description: await decryptField(t.description, dek),
        })),
      );

      // 3. Send plaintext to server for re-encryption with server-side DEK
      await disableE2E.mutateAsync({
        password,
        data: {
          accounts: decryptedAccounts,
          transactions: decryptedTransactions,
        },
      });

      clearDEK();

      const res = await refetchMe();
      if (res.data?.user) setUser(res.data.user);

      setShowDisableDialog(false);
      setPassword('');
      toast('End-to-end encryption disabled', 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to disable E2E', 'error');
    } finally {
      setMigrating(false);
    }
  }

  const isEnabled = user?.e2e_enabled || false;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <ShieldCheck className="h-5 w-5" />
          End-to-End Encryption
        </CardTitle>
        <CardDescription>
          Encrypt your data in the browser before it reaches the server
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <p className="font-medium">
              {isEnabled ? 'E2E Encryption Active' : 'E2E Encryption Disabled'}
            </p>
            <p className="text-sm text-muted-foreground">
              {isEnabled
                ? 'Your data is encrypted in the browser. The server never sees plaintext.'
                : 'Data is encrypted server-side. Enable E2E for maximum privacy.'}
            </p>
          </div>
          {isEnabled ? (
            <Button variant="outline" onClick={() => setShowDisableDialog(true)}>
              Disable
            </Button>
          ) : (
            <Button onClick={() => setShowEnableDialog(true)}>
              Enable
            </Button>
          )}
        </div>

        {isEnabled && (
          <div className="rounded-lg bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 p-3">
            <p className="text-sm text-green-800 dark:text-green-200">
              <strong>Active:</strong> All sensitive fields (account names, balances, transaction amounts, descriptions) are encrypted in your browser using AES-256-GCM before being sent to the server.
            </p>
          </div>
        )}

        <div className="rounded-lg bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-800 p-3">
          <p className="text-sm text-amber-800 dark:text-amber-200">
            <strong>Important:</strong> With E2E encryption enabled, some server-side features (CSV export, analytics) may show encrypted values. You must be logged in to view your data.
          </p>
        </div>

        <Dialog open={showEnableDialog} onClose={() => { setShowEnableDialog(false); setPassword(''); }}>
          <DialogHeader>
            <DialogTitle>Enable End-to-End Encryption</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleEnable} className="space-y-4 p-4">
            <p className="text-sm text-muted-foreground">
              Enter your password to generate encryption keys. All existing data will be re-encrypted client-side.
            </p>
            <Input
              type="password"
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              autoFocus
            />
            <Button type="submit" className="w-full" disabled={migrating || !password} loading={migrating}>
              {migrating ? 'Migrating data...' : 'Enable E2E Encryption'}
            </Button>
          </form>
        </Dialog>

        <Dialog open={showDisableDialog} onClose={() => { setShowDisableDialog(false); setPassword(''); }}>
          <DialogHeader>
            <DialogTitle>Disable End-to-End Encryption</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleDisable} className="space-y-4 p-4">
            <p className="text-sm text-muted-foreground">
              Your data will be decrypted and re-encrypted server-side. Enter your password to confirm.
            </p>
            <Input
              type="password"
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              autoFocus
            />
            <Button type="submit" variant="destructive" className="w-full" disabled={migrating || !password} loading={migrating}>
              {migrating ? 'Migrating data...' : 'Disable E2E Encryption'}
            </Button>
          </form>
        </Dialog>
      </CardContent>
    </Card>
  );
}

function DataExportSection() {
  const { toast } = useToast();
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');

  const exportTx = useExportTransactions();
  const exportAcc = useExportAccounts();
  const exportAll = useExportAll();

  const anyPending = exportTx.isPending || exportAcc.isPending || exportAll.isPending;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Download className="h-5 w-5" />
          Data Export
        </CardTitle>
        <CardDescription>Download your financial data in CSV or JSON format</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Transaction Date Range (optional)</label>
          <div className="flex gap-2">
            <Input type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} placeholder="From" />
            <Input type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} placeholder="To" />
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium">Transactions</label>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={anyPending}
              loading={exportTx.isPending}
              onClick={() => exportTx.mutate(
                { format: 'csv', from: dateFrom || undefined, to: dateTo || undefined },
                { onSuccess: () => toast('Transactions exported as CSV', 'success'), onError: (e) => toast(e.message, 'error') },
              )}
            >
              Export CSV
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={anyPending}
              loading={exportTx.isPending}
              onClick={() => exportTx.mutate(
                { format: 'json', from: dateFrom || undefined, to: dateTo || undefined },
                { onSuccess: () => toast('Transactions exported as JSON', 'success'), onError: (e) => toast(e.message, 'error') },
              )}
            >
              Export JSON
            </Button>
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium">Accounts</label>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={anyPending}
              loading={exportAcc.isPending}
              onClick={() => exportAcc.mutate(
                { format: 'csv' },
                { onSuccess: () => toast('Accounts exported as CSV', 'success'), onError: (e) => toast(e.message, 'error') },
              )}
            >
              Export CSV
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={anyPending}
              loading={exportAcc.isPending}
              onClick={() => exportAcc.mutate(
                { format: 'json' },
                { onSuccess: () => toast('Accounts exported as JSON', 'success'), onError: (e) => toast(e.message, 'error') },
              )}
            >
              Export JSON
            </Button>
          </div>
        </div>

        <div className="border-t pt-4">
          <Button
            variant="outline"
            disabled={anyPending}
            loading={exportAll.isPending}
            onClick={() => exportAll.mutate(undefined, {
              onSuccess: () => toast('All data exported', 'success'),
              onError: (e) => toast(e.message, 'error'),
            })}
          >
            Export All Data (JSON)
          </Button>
          <p className="text-xs text-muted-foreground mt-1">Downloads all accounts and transactions in a single JSON file</p>
        </div>
      </CardContent>
    </Card>
  );
}
