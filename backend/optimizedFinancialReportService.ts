/**
 * Optimized Financial Report Service
 * 
 * Leverages backend materialized view for ultra-fast financial report generation
 * Replaces fragmented service approach with unified, optimized solution
 */

import { API_V1_BASE } from '../config/api';
import { getAuthHeaders } from '../utils/authTokenUtils';

export interface OptimizedReportRequest {
  start_date?: string;
  end_date?: string;
  as_of_date?: string;
  format?: 'json' | 'pdf' | 'excel';
  auto_refresh?: boolean;
}

export interface OptimizedReportResponse {
  status: string;
  message: string;
  data: any;
  metadata: {
    source: string;
    generation_time_ms: number;
    data_freshness: string;
    record_count: number;
    uses_materialized_view: boolean;
  };
  generated_at: string;
}

export interface AccountBalanceData {
  account_id: number;
  account_code: string;
  account_name: string;
  account_type: string;
  account_category: string;
  normal_balance: string;
  total_debits: number;
  total_credits: number;
  transaction_count: number;
  last_transaction_date: string | null;
  current_balance: number;
  last_updated: string;
  is_active: boolean;
  is_header: boolean;
}

class OptimizedFinancialReportService {
  private readonly API_BASE = `${API_V1_BASE}/reports/optimized`;

  /**
   * ULTRA FAST Balance Sheet - Uses materialized view
   */
  async generateBalanceSheet(params: OptimizedReportRequest = {}): Promise<OptimizedReportResponse> {
    console.log('🚀 Generating optimized balance sheet using materialized view...');
    
    const queryParams = new URLSearchParams();
    if (params.as_of_date) queryParams.append('as_of_date', params.as_of_date);
    if (params.format) queryParams.append('format', params.format);
    
    const response = await fetch(`${this.API_BASE}/balance-sheet?${queryParams}`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error(`Balance sheet generation failed: ${response.statusText}`);
    }

    const result = await response.json();
    console.log(`✅ Balance sheet generated in ${result.metadata.generation_time_ms}ms`);
    return result;
  }

  /**
   * LIGHTNING FAST Trial Balance - Direct from materialized view
   */
  async generateTrialBalance(params: OptimizedReportRequest = {}): Promise<OptimizedReportResponse> {
    console.log('⚡ Generating lightning-fast trial balance...');
    
    const queryParams = new URLSearchParams();
    if (params.as_of_date) queryParams.append('as_of_date', params.as_of_date);
    if (params.format) queryParams.append('format', params.format);
    
    const response = await fetch(`${this.API_BASE}/trial-balance?${queryParams}`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error(`Trial balance generation failed: ${response.statusText}`);
    }

    const result = await response.json();
    console.log(`✅ Trial balance generated in ${result.metadata.generation_time_ms}ms`);
    return result;
  }

  /**
   * OPTIMIZED Profit & Loss - Fast P&L from materialized view
   */
  async generateProfitLoss(params: OptimizedReportRequest): Promise<OptimizedReportResponse> {
    if (!params.start_date || !params.end_date) {
      throw new Error('Start date and end date are required for P&L report');
    }

    console.log('💰 Generating optimized profit & loss statement...');
    
    const queryParams = new URLSearchParams();
    queryParams.append('start_date', params.start_date);
    queryParams.append('end_date', params.end_date);
    if (params.format) queryParams.append('format', params.format);
    
    const response = await fetch(`${this.API_BASE}/profit-loss?${queryParams}`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error(`P&L generation failed: ${response.statusText}`);
    }

    const result = await response.json();
    console.log(`✅ P&L generated in ${result.metadata.generation_time_ms}ms`);
    return result;
  }

  /**
   * REFRESH Materialized View - Manual refresh for latest data
   */
  async refreshAccountBalances(): Promise<OptimizedReportResponse> {
    console.log('🔄 Refreshing account balances materialized view...');
    
    const response = await fetch(`${this.API_BASE}/refresh-balances`, {
      method: 'POST',
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error(`Materialized view refresh failed: ${response.statusText}`);
    }

    const result = await response.json();
    console.log(`✅ Materialized view refreshed in ${result.metadata.generation_time_ms}ms`);
    return result;
  }

  /**
   * BATCH Report Generation - Generate multiple reports efficiently
   */
  async generateBatchReports(
    reportTypes: ('balance-sheet' | 'trial-balance' | 'profit-loss')[],
    params: OptimizedReportRequest
  ): Promise<{ [key: string]: OptimizedReportResponse }> {
    console.log('📊 Generating batch reports using materialized view...');
    
    // Refresh materialized view once before batch generation
    if (params.auto_refresh !== false) {
      await this.refreshAccountBalances();
    }
    
    const results: { [key: string]: OptimizedReportResponse } = {};
    const promises: Promise<void>[] = [];

    // Generate all reports concurrently for maximum performance
    reportTypes.forEach(reportType => {
      promises.push(
        (async () => {
          try {
            let result: OptimizedReportResponse;
            
            switch (reportType) {
              case 'balance-sheet':
                result = await this.generateBalanceSheet({ ...params, auto_refresh: false });
                break;
              case 'trial-balance':
                result = await this.generateTrialBalance({ ...params, auto_refresh: false });
                break;
              case 'profit-loss':
                result = await this.generateProfitLoss({ ...params, auto_refresh: false });
                break;
              default:
                throw new Error(`Unsupported report type: ${reportType}`);
            }
            
            results[reportType] = result;
          } catch (error) {
            console.error(`Failed to generate ${reportType}:`, error);
            results[reportType] = {
              status: 'error',
              message: `Failed to generate ${reportType}: ${error}`,
              data: null,
              metadata: {
                source: 'error',
                generation_time_ms: 0,
                data_freshness: 'error',
                record_count: 0,
                uses_materialized_view: false,
              },
              generated_at: new Date().toISOString(),
            };
          }
        })()
      );
    });

    await Promise.all(promises);
    
    console.log(`✅ Batch reports generated: ${Object.keys(results).length} reports`);
    return results;
  }

  /**
   * PERFORMANCE Metrics - Get generation time statistics
   */
  async getPerformanceMetrics(): Promise<{
    average_generation_time: number;
    fastest_report: string;
    slowest_report: string;
    materialized_view_usage: boolean;
  }> {
    // This would typically be tracked and stored, for demo we'll return mock data
    return {
      average_generation_time: 250, // ms
      fastest_report: 'trial-balance',
      slowest_report: 'profit-loss',
      materialized_view_usage: true,
    };
  }

  /**
   * VALIDATION - Validate report data integrity
   */
  validateReportData(reportData: any, reportType: string): {
    is_valid: boolean;
    errors: string[];
    warnings: string[];
  } {
    const errors: string[] = [];
    const warnings: string[] = [];

    if (!reportData || !reportData.data) {
      errors.push('Report data is missing or empty');
      return { is_valid: false, errors, warnings };
    }

    const data = reportData.data;

    // Common validations
    if (!data.generated_at) {
      warnings.push('Generated timestamp is missing');
    }

    // Report-specific validations
    switch (reportType) {
      case 'balance-sheet':
        if (data.total_assets !== undefined && data.total_liabilities_equity !== undefined) {
          if (Math.abs(data.total_assets - data.total_liabilities_equity) > 0.01) {
            errors.push('Balance sheet is not balanced: Assets ≠ Liabilities + Equity');
          }
        }
        break;

      case 'trial-balance':
        if (data.total_debits !== undefined && data.total_credits !== undefined) {
          if (Math.abs(data.total_debits - data.total_credits) > 0.01) {
            errors.push('Trial balance is not balanced: Total Debits ≠ Total Credits');
          }
        }
        break;

      case 'profit-loss':
        if (!data.period || !data.period.start_date || !data.period.end_date) {
          errors.push('P&L report is missing period information');
        }
        break;
    }

    return {
      is_valid: errors.length === 0,
      errors,
      warnings,
    };
  }

  /**
   * EXPORT Reports - Download reports in various formats
   */
  async downloadReport(
    reportType: 'balance-sheet' | 'trial-balance' | 'profit-loss',
    params: OptimizedReportRequest,
    format: 'pdf' | 'excel' = 'pdf'
  ): Promise<Blob> {
    const reportParams = { ...params, format };
    
    let result: OptimizedReportResponse;
    switch (reportType) {
      case 'balance-sheet':
        result = await this.generateBalanceSheet(reportParams);
        break;
      case 'trial-balance':
        result = await this.generateTrialBalance(reportParams);
        break;
      case 'profit-loss':
        result = await this.generateProfitLoss(reportParams);
        break;
    }

    // If result.data is already a Blob (from PDF/Excel generation)
    if (result.data instanceof Blob) {
      return result.data;
    }

    // Otherwise, convert JSON to downloadable format
    const jsonStr = JSON.stringify(result.data, null, 2);
    return new Blob([jsonStr], { type: 'application/json' });
  }

  /**
   * UTILITY - Format currency for display
   */
  formatCurrency(amount: number, currency: 'IDR' | 'USD' = 'IDR'): string {
    return new Intl.NumberFormat(currency === 'IDR' ? 'id-ID' : 'en-US', {
      style: 'currency',
      currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  }

  /**
   * UTILITY - Format percentage
   */
  formatPercentage(value: number, decimals: number = 2): string {
    return `${value.toFixed(decimals)}%`;
  }

  /**
   * UTILITY - Format date
   */
  formatDate(date: string | Date): string {
    const d = typeof date === 'string' ? new Date(date) : date;
    return d.toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }
}

// Export singleton instance
export const optimizedFinancialReportService = new OptimizedFinancialReportService();
export default optimizedFinancialReportService;