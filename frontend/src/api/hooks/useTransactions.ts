import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type { CreateTransactionRequest, PaginatedResponse, Transaction, TransactionFilter } from '@/types';

export function useTransactions(filter: TransactionFilter = {}) {
  const params = new URLSearchParams();
  if (filter.account_id) params.set('account_id', filter.account_id);
  if (filter.type) params.set('type', filter.type);
  if (filter.category_id) params.set('category_id', filter.category_id);
  if (filter.date_from) params.set('date_from', filter.date_from);
  if (filter.date_to) params.set('date_to', filter.date_to);
  if (filter.search) params.set('search', filter.search);
  if (filter.page) params.set('page', String(filter.page));
  if (filter.per_page) params.set('per_page', String(filter.per_page));

  return useQuery({
    queryKey: ['transactions', filter],
    queryFn: () =>
      apiFetch<PaginatedResponse<Transaction>>(`/transactions?${params.toString()}`),
  });
}

export function useCreateTransaction() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateTransactionRequest) =>
      apiFetch<{ transaction: Transaction }>('/transactions', { method: 'POST', body: JSON.stringify(req) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
    },
  });
}

export function useUpdateTransaction() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: Partial<Transaction> & { id: string }) =>
      apiFetch<{ transaction: Transaction }>(`/transactions/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
    },
  });
}

export function useDeleteTransaction() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/transactions/${id}`, { method: 'DELETE' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
    },
  });
}
