import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type { Budget, CreateBudgetRequest } from '@/types';

export function useBudgets() {
  return useQuery({
    queryKey: ['budgets'],
    queryFn: () => apiFetch<{ budgets: Budget[] }>('/budgets'),
  });
}

export function useCreateBudget() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateBudgetRequest) =>
      apiFetch<{ budget: Budget }>('/budgets', { method: 'POST', body: JSON.stringify(req) }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['budgets'] }),
  });
}

export function useUpdateBudget() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: Partial<Budget> & { id: string }) =>
      apiFetch<{ budget: Budget }>(`/budgets/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['budgets'] }),
  });
}

export function useDeleteBudget() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/budgets/${id}`, { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['budgets'] }),
  });
}
