import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';

export function useSetupTOTP() {
  return useMutation({
    mutationFn: () =>
      apiFetch<{ secret: string; url: string }>('/auth/totp/setup', { method: 'POST' }),
  });
}

export function useVerifyTOTP() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (code: string) =>
      apiFetch('/auth/totp/verify', { method: 'POST', body: JSON.stringify({ code }) }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['me'] }),
  });
}

export function useDisableTOTP() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => apiFetch('/auth/totp', { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['me'] }),
  });
}
