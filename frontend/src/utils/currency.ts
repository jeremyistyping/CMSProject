/**
 * Utility functions for Indonesian Rupiah currency formatting
 */

/**
 * Format number to Indonesian Rupiah currency
 * @param value - The numeric value to format
 * @param options - Additional formatting options
 * @returns Formatted currency string (e.g., "Rp 123.456")
 */
export const formatIDR = (
  value: number,
  options: {
    minimumFractionDigits?: number;
    maximumFractionDigits?: number;
    showSymbol?: boolean;
  } = {}
): string => {
  const {
    minimumFractionDigits = 0,
    maximumFractionDigits = 0,
    showSymbol = true,
  } = options;

  if (value === null || value === undefined || isNaN(value)) {
    return showSymbol ? 'Rp 0' : '0';
  }

  // Round value to remove any floating point precision errors
  const roundedValue = Math.round(value);

  const formatted = new Intl.NumberFormat('id-ID', {
    style: showSymbol ? 'currency' : 'decimal',
    currency: 'IDR',
    minimumFractionDigits,
    maximumFractionDigits,
  }).format(roundedValue);

  return formatted;
};

/**
 * Format number to Indonesian Rupiah currency with compact notation for large numbers
 * @param value - The numeric value to format
 * @returns Formatted currency string (e.g., "Rp 1,2 jt", "Rp 1,5 M")
 */
export const formatIDRCompact = (value: number): string => {
  if (value === null || value === undefined || isNaN(value)) {
    return 'Rp 0';
  }

  if (value >= 1000000000) {
    return `Rp ${(value / 1000000000).toFixed(1)} M`;
  } else if (value >= 1000000) {
    return `Rp ${(value / 1000000).toFixed(1)} jt`;
  } else if (value >= 1000) {
    return `Rp ${(value / 1000).toFixed(0)} rb`;
  }

  return formatIDR(value);
};

/**
 * Parse Indonesian formatted currency string back to number
 * @param formatted - The formatted currency string
 * @returns Numeric value
 */
export const parseIDR = (formatted: string): number => {
  if (!formatted) return 0;
  
  // Remove all non-digit characters except decimal separator
  const cleaned = formatted.replace(/[^\d,]/g, '').replace(/,/g, '.');
  const parsed = parseFloat(cleaned);
  
  return isNaN(parsed) ? 0 : parsed;
};

/**
 * Format number as percentage for Indonesian locale
 * @param value - The numeric value (as decimal, e.g., 0.15 for 15%)
 * @returns Formatted percentage string (e.g., "15%")
 */
export const formatPercentage = (value: number): string => {
  if (value === null || value === undefined || isNaN(value)) {
    return '0%';
  }
  
  return new Intl.NumberFormat('id-ID', {
    style: 'percent',
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(value);
};

/**
 * Format number for display in tables/lists (shorter format)
 * @param value - The numeric value to format
 * @returns Formatted currency string without decimals
 */
export const formatCurrencyTable = (value: number): string => {
  return formatIDR(value, { 
    minimumFractionDigits: 0, 
    maximumFractionDigits: 0 
  });
};

/**
 * Format number for detailed display (with decimals if needed)
 * @param value - The numeric value to format
 * @returns Formatted currency string with appropriate decimals
 */
export const formatCurrencyDetailed = (value: number): string => {
  // Show decimals only if the value has decimal places
  const hasDecimals = value % 1 !== 0;
  return formatIDR(value, { 
    minimumFractionDigits: hasDecimals ? 2 : 0, 
    maximumFractionDigits: 2 
  });
};
