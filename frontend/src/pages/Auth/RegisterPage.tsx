import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FormField } from '@/components/ui/FormField';
import { PasswordStrength } from '@/components/ui/PasswordStrength';
import { isPasswordStrong } from '@/components/ui/password-utils';
import { useRegister } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function RegisterPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const register = useRegister();
  const navigate = useNavigate();
  const { toast } = useToast();

  const errors: Record<string, string> = {};
  if (!email) errors.email = 'Email is required';
  else if (!EMAIL_RE.test(email)) errors.email = 'Invalid email format';
  if (password.length > 0 && !isPasswordStrong(password)) errors.password = 'Password is too weak';
  if (confirmPassword.length > 0 && password !== confirmPassword) errors.confirm = 'Passwords do not match';

  const isValid = !errors.email && !errors.password && !errors.confirm && password.length > 0 && confirmPassword.length > 0;

  function blur(field: string) {
    setTouched((t) => ({ ...t, [field]: true }));
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setTouched({ email: true, password: true, confirm: true });
    if (!isValid) return;
    register.mutate(
      { email, password },
      {
        onSuccess: () => {
          toast('Account created! Please sign in.', 'success');
          navigate('/login');
        },
      },
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4 bg-background">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-primary text-primary-foreground font-bold text-lg">
            MV
          </div>
          <CardTitle>Create your account</CardTitle>
          <CardDescription>Get started with MoneyVault</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <FormField label="Email" error={touched.email ? errors.email : undefined}>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onBlur={() => blur('email')}
              />
            </FormField>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="password">Password</label>
              <Input
                id="password"
                type="password"
                placeholder="Create a strong password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onBlur={() => blur('password')}
                minLength={8}
              />
              <PasswordStrength password={password} />
              {touched.password && errors.password && (
                <p className="text-xs text-destructive">{errors.password}</p>
              )}
            </div>
            <FormField label="Confirm Password" error={touched.confirm ? errors.confirm : undefined}>
              <Input
                id="confirm"
                type="password"
                placeholder="Repeat your password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                onBlur={() => blur('confirm')}
              />
            </FormField>

            {register.error && (
              <p className="text-sm text-destructive">{register.error.message}</p>
            )}

            <Button type="submit" className="w-full" loading={register.isPending} disabled={!isValid}>
              Create account
            </Button>
          </form>

          <p className="mt-4 text-center text-sm text-muted-foreground">
            Already have an account?{' '}
            <Link to="/login" className="text-primary hover:underline">
              Sign in
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
