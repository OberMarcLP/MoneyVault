import { useMutation, useQuery } from '@tanstack/react-query';
import { apiFetch } from '../client';

export function useVAPIDKey() {
  return useQuery({
    queryKey: ['vapid-key'],
    queryFn: () => apiFetch<{ public_key: string }>('/push/vapid-key'),
    staleTime: Infinity,
  });
}

export function usePushSubscribe() {
  return useMutation({
    mutationFn: (sub: { endpoint: string; auth: string; p256dh: string }) =>
      apiFetch('/push/subscribe', { method: 'POST', body: JSON.stringify(sub) }),
  });
}

export function usePushUnsubscribe() {
  return useMutation({
    mutationFn: (endpoint: string) =>
      apiFetch('/push/unsubscribe', { method: 'POST', body: JSON.stringify({ endpoint }) }),
  });
}
