import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type { RecurringTransaction, CreateRecurringRequest } from '@/types';

export function useRecurring() {
  return useQuery({
    queryKey: ['recurring'],
    queryFn: () => apiFetch<{ recurring: RecurringTransaction[] }>('/recurring'),
  });
}

export function useCreateRecurring() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateRecurringRequest) =>
      apiFetch<{ recurring: RecurringTransaction }>('/recurring', { method: 'POST', body: JSON.stringify(req) }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['recurring'] }),
  });
}

export function useDeleteRecurring() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/recurring/${id}`, { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['recurring'] }),
  });
}

export function useToggleRecurring() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/recurring/${id}/toggle`, { method: 'POST' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['recurring'] }),
  });
}
