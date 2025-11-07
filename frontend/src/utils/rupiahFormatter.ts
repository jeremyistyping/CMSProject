/**
 * Utility functions for formatting Indonesian Rupiah (IDR) currency
 */

/**
 * Format a number as Rupiah currency
 * @param amount - The number to format
 * @param options - Formatting options
 * @returns Formatted Rupiah string (e.g., "Rp 1.000.000")
 */
export const formatRupiah = (
  amount: number, 
  options: {
    includePrefix?: boolean;
    minimumFractionDigits?: number;
    maximumFractionDigits?: number;
  } = {}
): string => {
  const {
    includePrefix = true,
    minimumFractionDigits = 0,
    maximumFractionDigits = 0
  } = options;

  if (isNaN(amount)) return includePrefix ? 'Rp 0' : '0';

  const formatted = amount.toLocaleString('id-ID', {
    minimumFractionDigits,
    maximumFractionDigits
  });

  return includePrefix ? `Rp ${formatted}` : formatted;
};

/**
 * Format a number as Rupiah with automatic decimal handling
 * @param amount - The number to format
 * @param showDecimals - Whether to show decimal places
 * @returns Formatted Rupiah string
 */
export const formatRupiahAuto = (amount: number, showDecimals: boolean = false): string => {
  if (isNaN(amount)) return 'Rp 0';

  // If amount has decimal places and showDecimals is true, show 2 decimal places
  const hasDecimals = amount % 1 !== 0;
  const decimals = showDecimals && hasDecimals ? 2 : 0;

  return formatRupiah(amount, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  });
};

/**
 * Parse a Rupiah string to number
 * @param rupiahString - The Rupiah string to parse (e.g., "Rp 1.000.000" or "1.000.000")
 * @returns Parsed number
 */
export const parseRupiah = (rupiahString: string): number => {
  if (!rupiahString || typeof rupiahString !== 'string') return 0;

  // Remove Rp prefix, spaces, and dots (thousand separators)
  const cleanString = rupiahString
    .replace(/^Rp\s*/, '')
    .replace(/\./g, '')
    .replace(/,/g, '.') // Handle comma as decimal separator
    .trim();

  const parsed = parseFloat(cleanString);
  return isNaN(parsed) ? 0 : parsed;
};

/**
 * Format Rupiah for display in tables or lists
 * @param amount - The number to format
 * @param compact - Whether to use compact notation (K, M, B)
 * @returns Formatted string
 */
export const formatRupiahCompact = (amount: number, compact: boolean = false): string => {
  if (isNaN(amount)) return 'Rp 0';

  if (!compact) {
    return formatRupiah(amount);
  }

  // Compact notation
  if (amount >= 1000000000) {
    return `Rp ${(amount / 1000000000).toFixed(1)}M`;
  } else if (amount >= 1000000) {
    return `Rp ${(amount / 1000000).toFixed(1)}Jt`;
  } else if (amount >= 1000) {
    return `Rp ${(amount / 1000).toFixed(1)}K`;
  } else {
    return formatRupiah(amount);
  }
};

/**
 * Format Rupiah for accounting display (with parentheses for negative)
 * @param amount - The number to format
 * @param showParentheses - Whether to show parentheses for negative amounts
 * @returns Formatted string
 */
export const formatRupiahAccounting = (amount: number, showParentheses: boolean = true): string => {
  if (isNaN(amount)) return 'Rp 0';

  const absAmount = Math.abs(amount);
  const formattedAbs = formatRupiah(absAmount);

  if (amount < 0 && showParentheses) {
    return `(${formattedAbs})`;
  } else if (amount < 0) {
    return `-${formattedAbs}`;
  } else {
    return formattedAbs;
  }
};

/**
 * Get currency symbol
 * @returns Rupiah symbol
 */
export const getCurrencySymbol = (): string => 'Rp';

/**
 * Get currency code
 * @returns Currency code
 */
export const getCurrencyCode = (): string => 'IDR';

/**
 * Validate if a string is a valid Rupiah format
 * @param rupiahString - The string to validate
 * @returns Boolean indicating if valid
 */
export const isValidRupiahFormat = (rupiahString: string): boolean => {
  if (!rupiahString || typeof rupiahString !== 'string') return false;

  // Regex for valid Rupiah format: optional "Rp" followed by numbers with dots as thousand separators
  const rupiahRegex = /^(Rp\s?)?[\d,.]+(,\d{1,2})?$/;
  return rupiahRegex.test(rupiahString.trim());
};

/**
 * Convert amount to words (Indonesian)
 * @param amount - The number to convert
 * @returns Amount in words (Indonesian)
 */
export const rupiahToWords = (amount: number): string => {
  if (isNaN(amount) || amount < 0) return 'nol rupiah';

  const ones = [
    '', 'satu', 'dua', 'tiga', 'empat', 'lima', 'enam', 'tujuh', 'delapan', 'sembilan',
    'sepuluh', 'sebelas', 'dua belas', 'tiga belas', 'empat belas', 'lima belas', 
    'enam belas', 'tujuh belas', 'delapan belas', 'sembilan belas'
  ];

  const tens = ['', '', 'dua puluh', 'tiga puluh', 'empat puluh', 'lima puluh', 
                'enam puluh', 'tujuh puluh', 'delapan puluh', 'sembilan puluh'];

  const convertToWords = (num: number): string => {
    if (num === 0) return '';
    if (num < 20) return ones[num];
    if (num < 100) return tens[Math.floor(num / 10)] + (num % 10 !== 0 ? ' ' + ones[num % 10] : '');
    if (num < 1000) {
      const hundreds = Math.floor(num / 100);
      const remainder = num % 100;
      const hundredWord = hundreds === 1 ? 'seratus' : ones[hundreds] + ' ratus';
      return hundredWord + (remainder !== 0 ? ' ' + convertToWords(remainder) : '');
    }
    // Add more units as needed (ribuan, jutaan, miliaran, etc.)
    return num.toString(); // Fallback for very large numbers
  };

  if (amount === 0) return 'nol rupiah';
  
  const integerPart = Math.floor(amount);
  let result = convertToWords(integerPart) + ' rupiah';
  
  return result.charAt(0).toUpperCase() + result.slice(1);
};

// Export default object with all functions
export default {
  formatRupiah,
  formatRupiahAuto,
  formatRupiahCompact,
  formatRupiahAccounting,
  parseRupiah,
  getCurrencySymbol,
  getCurrencyCode,
  isValidRupiahFormat,
  rupiahToWords
};
