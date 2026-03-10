import { useQuery } from '@tanstack/react-query';
import { apiFetch } from '../client';

export function useExchangeRates(base?: string) {
  return useQuery({
    queryKey: ['exchange-rates', base],
    queryFn: () => apiFetch<{ base: string; rates: Record<string, number> }>(`/exchange-rates?base=${base || 'USD'}`),
    staleTime: 60 * 60 * 1000, // 1 hour
  });
}
