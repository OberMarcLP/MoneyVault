import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type { ExchangeConnection, CreateExchangeConnectionRequest, ExchangeSyncResult } from '@/types';

export function useExchangeConnections() {
  return useQuery({
    queryKey: ['exchange-connections'],
    queryFn: () => apiFetch<ExchangeConnection[]>('/exchanges'),
  });
}

export function useConnectExchange() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateExchangeConnectionRequest) =>
      apiFetch<ExchangeConnection>('/exchanges/connect', { method: 'POST', body: JSON.stringify(req) }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['exchange-connections'] }),
  });
}

export function useSyncExchange() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiFetch<ExchangeSyncResult>(`/exchanges/${id}/sync`, { method: 'POST' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['exchange-connections'] }),
  });
}

export function useToggleExchange() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/exchanges/${id}/toggle`, { method: 'POST' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['exchange-connections'] }),
  });
}

export function useDeleteExchange() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/exchanges/${id}`, { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['exchange-connections'] }),
  });
}
