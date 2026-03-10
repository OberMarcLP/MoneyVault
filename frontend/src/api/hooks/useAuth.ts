import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch, setAccessToken } from '../client';
import type { LoginRequest, LoginResponse, User } from '@/types';
import { useAuthStore } from '@/stores/auth';
import { useCryptoStore } from '@/stores/crypto';
import { decryptDEK } from '@/utils/crypto';

export function useLogin() {
  const setUser = useAuthStore((s) => s.setUser);
  const { setDEK, setE2EEnabled } = useCryptoStore.getState();
  return useMutation({
    mutationFn: async (req: LoginRequest) => {
      const data = await apiFetch<LoginResponse>('/auth/login', {
        method: 'POST',
        body: JSON.stringify(req),
      });
      setAccessToken(data.access_token);
      setUser(data.user);

      if (data.user.e2e_enabled && data.e2e_encrypted_dek && data.e2e_kek_salt) {
        const dek = await decryptDEK(data.e2e_encrypted_dek, req.password, data.e2e_kek_salt);
        setDEK(dek);
      } else {
        setE2EEnabled(false);
      }

      return data;
    },
  });
}

export function useRegister() {
  return useMutation({
    mutationFn: (req: { email: string; password: string }) =>
      apiFetch<{ user: User }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify(req),
      }),
  });
}

export function useLogout() {
  const clearUser = useAuthStore((s) => s.clearUser);
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => apiFetch('/auth/logout', { method: 'POST' }),
    onSettled: () => {
      setAccessToken(null);
      clearUser();
      useCryptoStore.getState().clearDEK();
      queryClient.clear();
    },
  });
}

export function useMe() {
  return useQuery({
    queryKey: ['me'],
    queryFn: () => apiFetch<{ user: User }>('/auth/me'),
    retry: false,
  });
}

export function useUpdatePreferences() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (prefs: { theme?: string; currency?: string; locale?: string; onboarding_dismissed?: boolean }) =>
      apiFetch('/auth/preferences', { method: 'PUT', body: JSON.stringify(prefs) }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['me'] }),
  });
}

export function useRequestPasswordReset() {
  return useMutation({
    mutationFn: (email: string) =>
      apiFetch<{ message: string; token?: string }>('/auth/password-reset/request', {
        method: 'POST',
        body: JSON.stringify({ email }),
      }),
  });
}

export function useResetPassword() {
  return useMutation({
    mutationFn: (req: { token: string; new_password: string }) =>
      apiFetch<{ message: string }>('/auth/password-reset/confirm', {
        method: 'POST',
        body: JSON.stringify(req),
      }),
  });
}

export function useVerifyEmail() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => apiFetch('/auth/verify-email', { method: 'POST' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['me'] }),
  });
}
