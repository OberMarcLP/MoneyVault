import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch, apiDownload } from '../client';
import type { CSVPreview, CSVImportMapping, ImportJob } from '@/types';

// CSV Import
export function useCSVPreview() {
  return useMutation({
    mutationFn: (file: File) => {
      const formData = new FormData();
      formData.append('file', file);
      return apiFetch<{ preview: CSVPreview }>('/import/preview', {
        method: 'POST',
        body: formData,
        headers: {},
      });
    },
  });
}

export function useCSVImport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ file, accountId, mapping, postedOnly }: {
      file: File;
      accountId: string;
      mapping: CSVImportMapping;
      postedOnly?: boolean;
    }) => {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('account_id', accountId);
      formData.append('map_date', mapping.date);
      formData.append('map_amount', mapping.amount);
      formData.append('map_description', mapping.description);
      formData.append('map_merchant', mapping.merchant || '');
      formData.append('map_category', mapping.category || '');
      formData.append('map_sub_category', mapping.sub_category || '');
      formData.append('map_type', mapping.type || '');
      formData.append('map_status', mapping.status || '');
      formData.append('map_currency', mapping.currency || '');
      formData.append('posted_only', postedOnly ? 'true' : 'false');
      return apiFetch<{ import: ImportJob }>('/import/csv', {
        method: 'POST',
        body: formData,
        headers: {},
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
    },
  });
}

export function useImportHistory() {
  return useQuery({
    queryKey: ['imports'],
    queryFn: () => apiFetch<{ imports: ImportJob[] }>('/import/history'),
  });
}

// Export
export function useExportTransactions() {
  return useMutation({
    mutationFn: ({ format, from, to }: { format: 'csv' | 'json'; from?: string; to?: string }) => {
      const params = new URLSearchParams({ format });
      if (from) params.set('from', from);
      if (to) params.set('to', to);
      return apiDownload(`/export/transactions?${params}`);
    },
  });
}

export function useExportAccounts() {
  return useMutation({
    mutationFn: ({ format }: { format: 'csv' | 'json' }) =>
      apiDownload(`/export/accounts?format=${format}`),
  });
}

export function useExportAll() {
  return useMutation({
    mutationFn: () => apiDownload('/export/all'),
  });
}
