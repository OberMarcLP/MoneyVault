import { useMemo } from 'react';
import { Check, X } from 'lucide-react';

interface PasswordStrengthProps {
  password: string;
}

interface Requirement {
  label: string;
  met: boolean;
}

export function PasswordStrength({ password }: PasswordStrengthProps) {
  const requirements = useMemo((): Requirement[] => [
    { label: 'At least 8 characters', met: password.length >= 8 },
    { label: 'One uppercase letter', met: /[A-Z]/.test(password) },
    { label: 'One lowercase letter', met: /[a-z]/.test(password) },
    { label: 'One digit', met: /\d/.test(password) },
    { label: 'One special character', met: /[^a-zA-Z0-9]/.test(password) },
  ], [password]);

  const metCount = requirements.filter((r) => r.met).length;
  const strength = metCount <= 2 ? 'weak' : metCount <= 4 ? 'medium' : 'strong';
  const strengthPercent = (metCount / requirements.length) * 100;

  const strengthColors = {
    weak: 'bg-destructive',
    medium: 'bg-amber-500',
    strong: 'bg-success',
  };

  const strengthLabels = {
    weak: 'Weak',
    medium: 'Medium',
    strong: 'Strong',
  };

  if (!password) return null;

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">Password strength</span>
        <span className={`font-medium ${
          strength === 'weak' ? 'text-destructive' :
          strength === 'medium' ? 'text-amber-500' :
          'text-success'
        }`}>
          {strengthLabels[strength]}
        </span>
      </div>
      <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
        <div
          className={`h-full rounded-full transition-all duration-300 ${strengthColors[strength]}`}
          style={{ width: `${strengthPercent}%` }}
        />
      </div>
      <ul className="space-y-1">
        {requirements.map((req) => (
          <li key={req.label} className="flex items-center gap-1.5 text-xs">
            {req.met ? (
              <Check className="h-3 w-3 text-success" />
            ) : (
              <X className="h-3 w-3 text-muted-foreground" />
            )}
            <span className={req.met ? 'text-success' : 'text-muted-foreground'}>
              {req.label}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}

export function isPasswordStrong(password: string): boolean {
  return (
    password.length >= 8 &&
    /[A-Z]/.test(password) &&
    /[a-z]/.test(password) &&
    /\d/.test(password) &&
    /[^a-zA-Z0-9]/.test(password)
  );
}
