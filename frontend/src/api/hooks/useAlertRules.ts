import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type { AlertRule, CreateAlertRuleRequest } from '@/types';

export function useAlertRules() {
  return useQuery({
    queryKey: ['alert-rules'],
    queryFn: () => apiFetch<{ rules: AlertRule[] }>('/alert-rules'),
  });
}

export function useCreateAlertRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateAlertRuleRequest) =>
      apiFetch<AlertRule>('/alert-rules', { method: 'POST', body: JSON.stringify(req) }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alert-rules'] }),
  });
}

export function useToggleAlertRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/alert-rules/${id}/toggle`, { method: 'POST' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alert-rules'] }),
  });
}

export function useDeleteAlertRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/alert-rules/${id}`, { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alert-rules'] }),
  });
}
