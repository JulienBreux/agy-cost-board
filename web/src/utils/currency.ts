/**
 * Normalizes ISO-4217 currency code. Defaults to 'USD' if absent or empty.
 */
export function normalizeCurrency(currency?: string | null): string {
  if (!currency || !currency.trim()) {
    return 'USD';
  }
  const trimmed = currency.trim().toUpperCase();
  if (trimmed === '$') return 'USD';
  if (trimmed === '€') return 'EUR';
  if (trimmed === '£') return 'GBP';
  if (trimmed === '¥') return 'JPY';
  return trimmed;
}


/**
 * Returns the display symbol for a given ISO-4217 currency code.
 */
export function getCurrencySymbol(currency?: string | null): string {
  const code = normalizeCurrency(currency);
  try {
    const formatter = new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: code,
    });
    const parts = formatter.formatToParts(0);
    const symbolPart = parts.find((p) => p.type === 'currency');
    if (symbolPart?.value) {
      return symbolPart.value;
    }
  } catch {
    // Fallback if Intl throws on unusual codes
  }

  // Fallback map
  switch (code) {
    case 'USD':
      return '$';
    case 'EUR':
      return '€';
    case 'GBP':
      return '£';
    case 'JPY':
      return '¥';
    case 'CAD':
      return 'CA$';
    case 'AUD':
      return 'A$';
    case 'CHF':
      return 'CHF ';
    case 'INR':
      return '₹';
    default:
      return `${code} `;
  }
}

/**
 * Formats a monetary number into a localized currency string using Intl.NumberFormat.
 */
export function formatCurrency(
  amount: number,
  currency?: string | null,
  options?: Intl.NumberFormatOptions
): string {
  const safeAmount = Number(amount ?? 0);
  const code = normalizeCurrency(currency);
  try {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: code,
      ...options,
    }).format(safeAmount);
  } catch {
    // If Intl.NumberFormat fails on an invalid currency code, fall back to USD or manual formatting
    const symbol = getCurrencySymbol(code);
    return `${symbol}${safeAmount.toLocaleString('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
      ...options,
    })}`;
  }
}
