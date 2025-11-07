/**
 * Balance Sheet Data Types
 * 
 * Tipe data untuk perhitungan Balance Sheet dari SSOT Journal System
 */

export interface BalanceSheetItem {
  id: number;
  account_code: string;
  account_name: string;
  account_type: 'ASSET' | 'LIABILITY' | 'EQUITY';
  category?: string;
  balance: number;
  debit_balance: number;
  credit_balance: number;
  is_header: boolean;
  level: number;
  parent_id?: number;
}

export interface BalanceSheetSection {
  name: string;
  type: 'ASSET' | 'LIABILITY' | 'EQUITY';
  items: BalanceSheetItem[];
  subtotal: number;
  subcategories?: BalanceSheetSubcategory[];
}

export interface BalanceSheetSubcategory {
  name: string;
  items: BalanceSheetItem[];
  subtotal: number;
}

export interface BalanceSheet {
  company_name?: string;
  report_title: string;
  as_of_date: string;
  generated_at: string;
  assets: BalanceSheetSection;
  liabilities: BalanceSheetSection;
  equity: BalanceSheetSection;
  total_assets: number;
  total_liabilities: number;
  total_equity: number;
  total_liabilities_equity: number;
  is_balanced: boolean;
  balance_difference: number;
  metadata: {
    source: 'SSOT_JOURNAL';
    journal_entries_count: number;
    accounts_included: number;
    calculation_method: string;
    data_freshness: string;
  };
}

export interface BalanceSheetCalculationOptions {
  as_of_date?: string;
  include_zero_balances?: boolean;
  include_inactive_accounts?: boolean;
  group_by_category?: boolean;
  currency_format?: 'IDR' | 'USD';
  detail_level?: 'summary' | 'detail' | 'full';
}

export interface AccountBalance {
  account_id: number;
  account_code: string;
  account_name: string;
  account_type: 'ASSET' | 'LIABILITY' | 'EQUITY' | 'REVENUE' | 'EXPENSE';
  category?: string;
  debit_balance: number;
  credit_balance: number;
  net_balance: number;
  is_header: boolean;
  level: number;
  parent_id?: number;
  last_updated: string;
}

export interface JournalSummaryForBalanceSheet {
  total_journal_entries: number;
  posted_entries: number;
  date_range: {
    earliest_entry: string;
    latest_entry: string;
  };
  accounts_affected: number;
  total_debits: number;
  total_credits: number;
  is_balanced: boolean;
}

export interface BalanceSheetValidationResult {
  is_valid: boolean;
  errors: ValidationError[];
  warnings: ValidationWarning[];
  balance_check: {
    assets_total: number;
    liabilities_equity_total: number;
    difference: number;
    tolerance: number;
    is_balanced: boolean;
  };
}

export interface ValidationError {
  type: 'BALANCE_MISMATCH' | 'MISSING_ACCOUNTS' | 'INVALID_DATA' | 'CALCULATION_ERROR';
  message: string;
  details?: any;
}

export interface ValidationWarning {
  type: 'ZERO_BALANCE' | 'NEGATIVE_ASSET' | 'POSITIVE_LIABILITY' | 'DATA_FRESHNESS';
  message: string;
  details?: any;
}

// Asset subcategory types untuk kategorisasi yang lebih detail
export enum AssetCategory {
  CURRENT_ASSETS = 'Current Assets',
  FIXED_ASSETS = 'Fixed Assets', 
  INTANGIBLE_ASSETS = 'Intangible Assets',
  INVESTMENTS = 'Investments',
  OTHER_ASSETS = 'Other Assets'
}

// Liability subcategory types
export enum LiabilityCategory {
  CURRENT_LIABILITIES = 'Current Liabilities',
  LONG_TERM_LIABILITIES = 'Long-term Liabilities',
  OTHER_LIABILITIES = 'Other Liabilities'
}

// Equity subcategory types
export enum EquityCategory {
  PAID_IN_CAPITAL = 'Paid-in Capital',
  RETAINED_EARNINGS = 'Retained Earnings',
  OTHER_EQUITY = 'Other Equity'
}

export interface DetailedBalanceSheetSection extends BalanceSheetSection {
  categories: {
    [key: string]: {
      items: BalanceSheetItem[];
      subtotal: number;
    };
  };
}

export interface EnhancedBalanceSheet extends BalanceSheet {
  assets: DetailedBalanceSheetSection;
  liabilities: DetailedBalanceSheetSection;
  equity: DetailedBalanceSheetSection;
  ratios?: {
    current_ratio?: number;
    debt_to_equity?: number;
    equity_ratio?: number;
    asset_turnover?: number;
  };
}