import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type { CryptoSummary, CoinGeckoToken, Wallet, WalletTransaction } from '@/types';

export function useCryptoSummary() {
  return useQuery({
    queryKey: ['crypto-summary'],
    queryFn: () => apiFetch<CryptoSummary>('/crypto/summary'),
  });
}

export function useSearchTokens(query: string) {
  return useQuery({
    queryKey: ['token-search', query],
    queryFn: () => apiFetch<CoinGeckoToken[]>(`/crypto/search?q=${encodeURIComponent(query)}`),
    enabled: query.length >= 2,
  });
}

export function useWallets() {
  return useQuery({
    queryKey: ['wallets'],
    queryFn: () => apiFetch<Wallet[]>('/crypto/wallets'),
  });
}

export function useCreateWallet() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: { address: string; network?: string; label?: string }) =>
      apiFetch<Wallet>('/crypto/wallets', { method: 'POST', body: JSON.stringify(req) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wallets'] });
      queryClient.invalidateQueries({ queryKey: ['crypto-summary'] });
    },
  });
}

export function useDeleteWallet() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/crypto/wallets/${id}`, { method: 'DELETE' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wallets'] });
      queryClient.invalidateQueries({ queryKey: ['crypto-summary'] });
    },
  });
}

export function useSyncWallet() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch<{ synced: number }>(`/crypto/wallets/${id}/sync`, { method: 'POST' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wallets'] });
      queryClient.invalidateQueries({ queryKey: ['wallet-transactions'] });
    },
  });
}

export function useWalletTransactions(walletId: string) {
  return useQuery({
    queryKey: ['wallet-transactions', walletId],
    queryFn: () => apiFetch<WalletTransaction[]>(`/crypto/wallets/${walletId}/transactions`),
    enabled: !!walletId,
  });
}

export function useAllWalletTransactions() {
  return useQuery({
    queryKey: ['wallet-transactions', 'all'],
    queryFn: () => apiFetch<WalletTransaction[]>('/crypto/wallet-transactions'),
  });
}

export function useRefreshCryptoPrices() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => apiFetch('/crypto/refresh-prices', { method: 'POST' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['holdings'] });
      queryClient.invalidateQueries({ queryKey: ['portfolio-summary'] });
      queryClient.invalidateQueries({ queryKey: ['crypto-summary'] });
    },
  });
}
