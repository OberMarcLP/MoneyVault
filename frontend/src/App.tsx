import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AppLayout } from '@/components/layout/AppLayout';
import { LoginPage } from '@/pages/Auth/LoginPage';
import { RegisterPage } from '@/pages/Auth/RegisterPage';
import { ForgotPasswordPage } from '@/pages/Auth/ForgotPasswordPage';
import { ResetPasswordPage } from '@/pages/Auth/ResetPasswordPage';
import { DashboardPage } from '@/pages/Dashboard/DashboardPage';
import { AccountsPage } from '@/pages/Accounts/AccountsPage';
import { TransactionsPage } from '@/pages/Transactions/TransactionsPage';
import { CategoriesPage } from '@/pages/Categories/CategoriesPage';
import { BudgetsPage } from '@/pages/Budgets/BudgetsPage';
import { RecurringPage } from '@/pages/Recurring/RecurringPage';
import { ImportPage } from '@/pages/Import/ImportPage';
import { InvestmentsPage } from '@/pages/Investments/InvestmentsPage';
import { CryptoPage } from '@/pages/Crypto/CryptoPage';
import { ReportsPage } from '@/pages/Reports/ReportsPage';
import AlertsPage from '@/pages/Alerts/AlertsPage';
import { SettingsPage } from '@/pages/Settings/SettingsPage';
import { AdminPage } from '@/pages/Admin/AdminPage';
import { ToastProvider } from '@/components/ui/toast';
import { PWAInstallPrompt } from '@/components/PWAInstallPrompt';
import { ErrorBoundary } from '@/components/ErrorBoundary';
import { useAuthStore } from '@/stores/auth';
import { initAuth, apiFetch } from '@/api/client';
import { useState, useEffect, type ReactNode } from 'react';
import type { User } from '@/types';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
});

function ProtectedRoute({ children }: { children: ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const clearUser = useAuthStore((s) => s.clearUser);
  const setUser = useAuthStore((s) => s.setUser);
  const [ready, setReady] = useState(!isAuthenticated);

  useEffect(() => {
    if (!isAuthenticated) return;
    initAuth().then(async (ok) => {
      if (!ok) {
        clearUser();
      } else {
        // Refresh user data (role, preferences) from server
        try {
          const data = await apiFetch<{ user: User }>('/auth/me');
          setUser(data.user);
        } catch {
          // Keep existing user data if /me fails
        }
      }
      setReady(true);
    });
  }, [isAuthenticated, clearUser, setUser]);

  if (!ready) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

function GuestRoute({ children }: { children: ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  if (isAuthenticated) return <Navigate to="/" replace />;
  return <>{children}</>;
}

function AdminRoute({ children }: { children: ReactNode }) {
  const user = useAuthStore((s) => s.user);
  if (user?.role !== 'admin') return <Navigate to="/" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<GuestRoute><LoginPage /></GuestRoute>} />
            <Route path="/register" element={<GuestRoute><RegisterPage /></GuestRoute>} />
            <Route path="/forgot-password" element={<GuestRoute><ForgotPasswordPage /></GuestRoute>} />
            <Route path="/reset-password" element={<GuestRoute><ResetPasswordPage /></GuestRoute>} />

            <Route element={<ProtectedRoute><ErrorBoundary><AppLayout /></ErrorBoundary></ProtectedRoute>}>
              <Route path="/" element={<DashboardPage />} />
              <Route path="/accounts" element={<AccountsPage />} />
              <Route path="/transactions" element={<TransactionsPage />} />
              <Route path="/categories" element={<CategoriesPage />} />
              <Route path="/budgets" element={<BudgetsPage />} />
              <Route path="/recurring" element={<RecurringPage />} />
              <Route path="/import" element={<ImportPage />} />
              <Route path="/investments" element={<InvestmentsPage />} />
              <Route path="/crypto" element={<CryptoPage />} />
              <Route path="/reports" element={<ReportsPage />} />
              <Route path="/alerts" element={<AlertsPage />} />
              <Route path="/settings" element={<SettingsPage />} />
              <Route path="/admin" element={<AdminRoute><AdminPage /></AdminRoute>} />
            </Route>

            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
        <PWAInstallPrompt />
      </ToastProvider>
    </QueryClientProvider>
  );
}
