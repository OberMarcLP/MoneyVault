import { describe, it, expect } from 'vitest';
import { formatCurrency, formatDate, getAccountTypeLabel, getTransactionTypeColor } from './utils';

describe('formatCurrency', () => {
  it('formats USD amounts', () => {
    expect(formatCurrency(100, 'USD')).toBe('$100.00');
    expect(formatCurrency(1234.56, 'USD')).toBe('$1,234.56');
    expect(formatCurrency(0, 'USD')).toBe('$0.00');
  });

  it('formats string amounts', () => {
    expect(formatCurrency('100.50', 'USD')).toBe('$100.50');
  });

  it('formats EUR amounts', () => {
    const result = formatCurrency(100, 'EUR', 'de-DE');
    expect(result).toContain('100');
    expect(result).toContain('€');
  });

  it('uses USD as default currency', () => {
    expect(formatCurrency(50)).toBe('$50.00');
  });
});

describe('formatDate', () => {
  it('formats ISO date strings', () => {
    const result = formatDate('2024-01-15');
    expect(result).toContain('Jan');
    expect(result).toContain('15');
    expect(result).toContain('2024');
  });
});

describe('getAccountTypeLabel', () => {
  it('returns correct labels', () => {
    expect(getAccountTypeLabel('checking')).toBe('Checking');
    expect(getAccountTypeLabel('savings')).toBe('Savings');
    expect(getAccountTypeLabel('credit')).toBe('Credit Card');
    expect(getAccountTypeLabel('investment')).toBe('Investment');
    expect(getAccountTypeLabel('crypto_wallet')).toBe('Crypto Wallet');
  });

  it('returns raw type for unknown types', () => {
    expect(getAccountTypeLabel('unknown')).toBe('unknown');
  });
});

describe('getTransactionTypeColor', () => {
  it('returns correct colors', () => {
    expect(getTransactionTypeColor('income')).toBe('text-success');
    expect(getTransactionTypeColor('expense')).toBe('text-destructive');
    expect(getTransactionTypeColor('transfer')).toBe('text-primary');
  });

  it('returns empty string for unknown types', () => {
    expect(getTransactionTypeColor('other')).toBe('');
  });
});
