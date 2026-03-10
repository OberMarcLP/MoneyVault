import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type { Dividend, DividendSummary, CreateDividendRequest } from '@/types';

export function useDividends(holdingId?: string) {
  return useQuery({
    queryKey: ['dividends', holdingId],
    queryFn: () => apiFetch<{ dividends: Dividend[] }>(`/investments/dividends${holdingId ? `?holding_id=${holdingId}` : ''}`),
  });
}

export function useDividendSummary() {
  return useQuery({
    queryKey: ['dividend-summary'],
    queryFn: () => apiFetch<DividendSummary>('/investments/dividends/summary'),
  });
}

export function useCreateDividend() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateDividendRequest) => apiFetch<Dividend>('/investments/dividends', {
      method: 'POST',
      body: JSON.stringify(req),
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['dividends'] });
      qc.invalidateQueries({ queryKey: ['dividend-summary'] });
    },
  });
}

export function useDeleteDividend() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/investments/dividends/${id}`, { method: 'DELETE' }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['dividends'] });
      qc.invalidateQueries({ queryKey: ['dividend-summary'] });
    },
  });
}
