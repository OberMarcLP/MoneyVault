import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type {
  Holding, PortfolioSummary, CreateHoldingRequest,
  UpdateHoldingRequest, SellHoldingRequest, TradeLot, PriceHistory,
} from '@/types';

export function useHoldings() {
  return useQuery({
    queryKey: ['holdings'],
    queryFn: () => apiFetch<Holding[]>('/investments'),
  });
}

export function useHolding(id: string) {
  return useQuery({
    queryKey: ['holdings', id],
    queryFn: () => apiFetch<Holding>(`/investments/${id}`),
    enabled: !!id,
  });
}

export function usePortfolioSummary() {
  return useQuery({
    queryKey: ['portfolio-summary'],
    queryFn: () => apiFetch<PortfolioSummary>('/investments/summary'),
  });
}

export function useCreateHolding() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateHoldingRequest) =>
      apiFetch<Holding>('/investments', { method: 'POST', body: JSON.stringify(req) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['holdings'] });
      queryClient.invalidateQueries({ queryKey: ['portfolio-summary'] });
    },
  });
}

export function useUpdateHolding() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: UpdateHoldingRequest & { id: string }) =>
      apiFetch<Holding>(`/investments/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['holdings'] });
      queryClient.invalidateQueries({ queryKey: ['portfolio-summary'] });
    },
  });
}

export function useDeleteHolding() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/investments/${id}`, { method: 'DELETE' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['holdings'] });
      queryClient.invalidateQueries({ queryKey: ['portfolio-summary'] });
    },
  });
}

export function useSellHolding() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: SellHoldingRequest & { id: string }) =>
      apiFetch(`/investments/${id}/sell`, { method: 'POST', body: JSON.stringify(data) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['holdings'] });
      queryClient.invalidateQueries({ queryKey: ['portfolio-summary'] });
    },
  });
}

export function useRealizedGains() {
  return useQuery({
    queryKey: ['realized-gains'],
    queryFn: () => apiFetch<TradeLot[]>('/investments/gains'),
  });
}

export function usePriceHistory(symbol: string, days = 30) {
  return useQuery({
    queryKey: ['price-history', symbol, days],
    queryFn: () => apiFetch<PriceHistory[]>(`/investments/price-history/${symbol}?days=${days}`),
    enabled: !!symbol,
  });
}

export function useRefreshPrices() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => apiFetch('/investments/refresh-prices', { method: 'POST' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['holdings'] });
      queryClient.invalidateQueries({ queryKey: ['portfolio-summary'] });
    },
  });
}
