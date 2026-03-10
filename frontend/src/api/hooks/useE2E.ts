import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type { E2EMigrateData } from '@/types';
import { useCryptoStore } from '@/stores/crypto';

export function useE2EExportData() {
  return useMutation({
    mutationFn: () => apiFetch<E2EMigrateData>('/e2e/export-data'),
  });
}

export function useE2EEnable() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: {
      password: string;
      e2e_encrypted_dek: string;
      e2e_kek_salt: string;
      data: E2EMigrateData;
    }) => apiFetch('/e2e/enable', { method: 'POST', body: JSON.stringify(req) }),
    onSuccess: () => {
      queryClient.invalidateQueries();
    },
  });
}

export function useE2EDisable() {
  const queryClient = useQueryClient();
  const { clearDEK } = useCryptoStore.getState();
  return useMutation({
    mutationFn: (req: { password: string; data: E2EMigrateData }) =>
      apiFetch('/e2e/disable', { method: 'POST', body: JSON.stringify(req) }),
    onSuccess: () => {
      clearDEK();
      queryClient.invalidateQueries();
    },
  });
}
