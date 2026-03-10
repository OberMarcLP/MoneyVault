import { useQuery } from '@tanstack/react-query';
import { apiFetch } from '../client';
import type {
  NetWorthSnapshot, SpendingByCategory, SpendingTrend,
  TopExpense, BudgetHistory, CashFlowResult, AssetAllocation,
} from '@/types';

export function useNetWorthHistory(days = 90) {
  return useQuery({
    queryKey: ['analytics', 'net-worth', days],
    queryFn: () => apiFetch<NetWorthSnapshot[]>(`/analytics/net-worth?days=${days}`),
  });
}

export function useSpendingBreakdown(period = 'month') {
  return useQuery({
    queryKey: ['analytics', 'spending', period],
    queryFn: () => apiFetch<SpendingByCategory[]>(`/analytics/spending?period=${period}`),
  });
}

export function useSpendingTrends(months = 12) {
  return useQuery({
    queryKey: ['analytics', 'trends', months],
    queryFn: () => apiFetch<SpendingTrend[]>(`/analytics/trends?months=${months}`),
  });
}

export function useTopExpenses(period = 'month', limit = 10) {
  return useQuery({
    queryKey: ['analytics', 'top-expenses', period, limit],
    queryFn: () => apiFetch<TopExpense[]>(`/analytics/top-expenses?period=${period}&limit=${limit}`),
  });
}

export function useBudgetVsActual(period = 'month') {
  return useQuery({
    queryKey: ['analytics', 'budget-vs-actual', period],
    queryFn: () => apiFetch<BudgetHistory>(`/analytics/budget-vs-actual?period=${period}`),
  });
}

export function useCashFlowForecast(months = 6) {
  return useQuery({
    queryKey: ['analytics', 'cash-flow', months],
    queryFn: () => apiFetch<CashFlowResult>(`/analytics/cash-flow?months=${months}`),
  });
}

export function useAssetAllocation() {
  return useQuery({
    queryKey: ['analytics', 'asset-allocation'],
    queryFn: () => apiFetch<AssetAllocation[]>('/analytics/asset-allocation'),
  });
}
