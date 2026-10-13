import { describe, it, expect } from 'vitest';
import { formatCurrency, getCurrencySymbol } from './currency';

describe('currency utils', () => {
  describe('formatCurrency', () => {
    it('formats USD by default', () => {
      expect(formatCurrency(1234.56)).toBe('$1,234.56');
      expect(formatCurrency(0)).toBe('$0.00');
    });

    it('formats EUR correctly', () => {
      expect(formatCurrency(1234.56, 'EUR')).toBe('€1,234.56');
      expect(formatCurrency(18, 'EUR')).toBe('€18.00');
    });

    it('formats GBP correctly', () => {
      expect(formatCurrency(1234.56, 'GBP')).toBe('£1,234.56');
    });

    it('formats JPY correctly', () => {
      expect(formatCurrency(5000, 'JPY')).toBe('¥5,000');
    });

    it('formats CAD correctly', () => {
      expect(formatCurrency(1234.56, 'CAD')).toBe('CA$1,234.56');
    });

    it('handles custom fraction digits options', () => {
      expect(formatCurrency(1234.567, 'USD', { minimumFractionDigits: 0, maximumFractionDigits: 0 })).toBe('$1,235');
      expect(formatCurrency(18, 'EUR', { minimumFractionDigits: 0, maximumFractionDigits: 0 })).toBe('€18');
    });

    it('handles fallback for empty or undefined currency', () => {
      expect(formatCurrency(100, '')).toBe('$100.00');
      expect(formatCurrency(100, undefined)).toBe('$100.00');
    });

    it('handles negative amounts', () => {
      expect(formatCurrency(-50, 'USD')).toBe('-$50.00');
      expect(formatCurrency(-50, 'EUR')).toBe('-€50.00');
    });
  });

  describe('getCurrencySymbol', () => {
    it('returns $ for USD or empty/undefined', () => {
      expect(getCurrencySymbol('USD')).toBe('$');
      expect(getCurrencySymbol('')).toBe('$');
      expect(getCurrencySymbol(undefined)).toBe('$');
    });

    it('returns € for EUR', () => {
      expect(getCurrencySymbol('EUR')).toBe('€');
    });

    it('returns £ for GBP', () => {
      expect(getCurrencySymbol('GBP')).toBe('£');
    });

    it('returns ¥ for JPY', () => {
      expect(getCurrencySymbol('JPY')).toBe('¥');
    });

    it('returns CA$ for CAD', () => {
      expect(getCurrencySymbol('CAD')).toBe('CA$');
    });

    it('returns A$ for AUD', () => {
      expect(getCurrencySymbol('AUD')).toBe('A$');
    });

    it('handles symbol inputs seamlessly', () => {
      expect(getCurrencySymbol('$')).toBe('$');
      expect(getCurrencySymbol('€')).toBe('€');
      expect(formatCurrency(25, '$')).toBe('$25.00');
      expect(formatCurrency(25, '€')).toBe('€25.00');
    });
  });
});

