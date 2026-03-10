import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FormField } from '@/components/ui/FormField';
import { useLogin, useWebAuthnLoginBegin, useWebAuthnLoginFinish } from '@/api/hooks';
import { authenticatePasskey, isWebAuthnSupported } from '@/lib/webauthn';
import { Key } from 'lucide-react';

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [totpCode, setTotpCode] = useState('');
  const [needsTOTP, setNeedsTOTP] = useState(false);
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [passkeyError, setPasskeyError] = useState('');
  const [isPasskeyLoading, setIsPasskeyLoading] = useState(false);
  const login = useLogin();
  const passkeyBegin = useWebAuthnLoginBegin();
  const passkeyFinish = useWebAuthnLoginFinish();
  const navigate = useNavigate();

  const emailError = touched.email && email.length > 0 && !EMAIL_RE.test(email) ? 'Invalid email format' : undefined;
  const supportsPasskeys = isWebAuthnSupported();

  function blur(field: string) {
    setTouched((t) => ({ ...t, [field]: true }));
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setPasskeyError('');
    login.mutate(
      { email, password, totp_code: totpCode || undefined },
      {
        onSuccess: () => navigate('/'),
        onError: (err) => {
          if (err.message === 'totp_required') {
            setNeedsTOTP(true);
          }
        },
      },
    );
  }

  async function handlePasskeyLogin() {
    setPasskeyError('');
    setIsPasskeyLoading(true);
    try {
      const { options } = await passkeyBegin.mutateAsync();
      const credential = await authenticatePasskey(options as unknown as Record<string, unknown>);
      await passkeyFinish.mutateAsync(credential);
      navigate('/');
    } catch (err) {
      setPasskeyError(err instanceof Error ? err.message : 'Passkey authentication failed');
    } finally {
      setIsPasskeyLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4 bg-background">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-primary text-primary-foreground font-bold text-lg">
            MV
          </div>
          <CardTitle>Welcome back</CardTitle>
          <CardDescription>Sign in to your MoneyVault account</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <FormField label="Email" error={emailError}>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onBlur={() => blur('email')}
                required
              />
            </FormField>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="password">Password</label>
              <Input
                id="password"
                type="password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>

            {needsTOTP && (
              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="totp">2FA Code</label>
                <Input
                  id="totp"
                  type="text"
                  placeholder="Enter 6-digit code"
                  value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value)}
                  maxLength={6}
                  autoFocus
                />
              </div>
            )}

            {login.error && (
              <p className="text-sm text-destructive">
                {login.error.message === 'totp_required' ? 'Please enter your 2FA code' : login.error.message}
              </p>
            )}

            <Button type="submit" className="w-full" loading={login.isPending}>
              Sign in
            </Button>
          </form>

          {supportsPasskeys && (
            <>
              <div className="relative my-4">
                <div className="absolute inset-0 flex items-center">
                  <span className="w-full border-t" />
                </div>
                <div className="relative flex justify-center text-xs uppercase">
                  <span className="bg-card px-2 text-muted-foreground">or</span>
                </div>
              </div>

              <Button
                variant="outline"
                className="w-full"
                onClick={handlePasskeyLogin}
                disabled={isPasskeyLoading}
                loading={isPasskeyLoading}
              >
                <Key className="mr-2 h-4 w-4" />
                Sign in with Passkey
              </Button>

              {passkeyError && (
                <p className="mt-2 text-sm text-destructive">{passkeyError}</p>
              )}
            </>
          )}

          <div className="mt-4 space-y-2 text-center text-sm text-muted-foreground">
            <p>
              <Link to="/forgot-password" className="text-primary hover:underline">
                Forgot your password?
              </Link>
            </p>
            <p>
              Don't have an account?{' '}
              <Link to="/register" className="text-primary hover:underline">
                Create one
              </Link>
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
