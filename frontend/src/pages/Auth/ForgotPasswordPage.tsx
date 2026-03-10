import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useRequestPasswordReset } from '@/api/hooks';

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [resetToken, setResetToken] = useState<string | null>(null);
  const requestReset = useRequestPasswordReset();

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    requestReset.mutate(email, {
      onSuccess: (data) => {
        if (data.token) {
          setResetToken(data.token);
        }
      },
    });
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4 bg-background">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-primary text-primary-foreground font-bold text-lg">
            MV
          </div>
          <CardTitle>Reset Password</CardTitle>
          <CardDescription>Enter your email to receive a reset token</CardDescription>
        </CardHeader>
        <CardContent>
          {resetToken ? (
            <div className="space-y-4">
              <div className="rounded-lg border border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950 p-4">
                <p className="text-sm font-medium text-green-800 dark:text-green-200 mb-2">Reset token generated</p>
                <p className="text-xs text-muted-foreground mb-2">
                  Copy this token and use it on the reset password page. It expires in 1 hour.
                </p>
                <code className="block break-all rounded bg-muted p-2 text-xs font-mono">
                  {resetToken}
                </code>
              </div>
              <Link to={`/reset-password?token=${encodeURIComponent(resetToken)}`}>
                <Button className="w-full">Continue to Reset Password</Button>
              </Link>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="email">Email</label>
                <Input
                  id="email"
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>

              {requestReset.error && (
                <p className="text-sm text-destructive">{requestReset.error.message}</p>
              )}

              {requestReset.isSuccess && !resetToken && (
                <p className="text-sm text-green-600 dark:text-green-400">
                  If an account with that email exists, a reset token has been generated.
                </p>
              )}

              <Button type="submit" className="w-full" disabled={requestReset.isPending}>
                {requestReset.isPending ? 'Sending...' : 'Request Reset Token'}
              </Button>
            </form>
          )}

          <p className="mt-4 text-center text-sm text-muted-foreground">
            <Link to="/login" className="text-primary hover:underline">
              Back to Sign In
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
