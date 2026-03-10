import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type { AuditLog, User } from '@/types';

export function useAdminUsers() {
  return useQuery({
    queryKey: ['admin', 'users'],
    queryFn: () => apiFetch<{ users: User[] }>('/admin/users'),
  });
}

export function useAdminUpdateUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, role }: { id: string; role: string }) =>
      apiFetch(`/admin/users/${id}/role`, { method: 'PUT', body: JSON.stringify({ role }) }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'users'] }),
  });
}

export function useAdminDeleteUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/admin/users/${id}`, { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'users'] }),
  });
}

export function useAuditLogs(filter: { action?: string; page?: number; limit?: number } = {}) {
  const params = new URLSearchParams();
  if (filter.action) params.set('action', filter.action);
  if (filter.page) params.set('page', String(filter.page));
  params.set('limit', String(filter.limit ?? 20));
  return useQuery({
    queryKey: ['admin', 'audit-logs', filter],
    queryFn: () => apiFetch<{ logs: AuditLog[]; total: number; page: number; total_pages: number }>(
      `/admin/audit-logs?${params.toString()}`
    ),
  });
}
