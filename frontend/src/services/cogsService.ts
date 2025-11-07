import { API_V1_BASE } from '@/config/api';
import { getAuthHeaders } from '@/utils/authTokenUtils';

/**
 * COGS Service - Cost of Goods Sold Management
 * 
 * This service handles COGS (Cost of Goods Sold) operations including:
 * - Getting COGS summary for periods
 * - Finding sales without COGS entries
 * - Backfilling COGS for existing sales
 * - Recording COGS for individual sales
 */

export interface COGSSummary {
  total_cogs: number;
  total_sales: number;
  total_revenue: number;
  avg_cogs: number;
  cogs_percentage: number;
}

export interface SaleWithoutCOGS {
  id: number;
  invoice_number: string;
  sale_date: string;
  total_amount: number;
  status: string;
  estimated_cogs: number;
  customer_name: string;
}

export interface BackfillPreview {
  sale_id: number;
  invoice_number: string;
  sale_date: string;
  total_amount: number;
  estimated_cogs: number;
  cogs_percentage: number;
}

export interface BackfillResponse {
  status: string;
  dry_run?: boolean;
  sales_to_process?: number;
  sales_processed?: number;
  total_estimated_cogs: number;
  preview?: BackfillPreview[];
  message: string;
}

class COGSService {
  private getAuthHeaders() {
    return getAuthHeaders();
  }

  /**
   * Get COGS summary for a period
   */
  async getCOGSSummary(startDate: string, endDate: string): Promise<COGSSummary> {
    const params = new URLSearchParams({
      start_date: startDate,
      end_date: endDate,
    });

    const response = await fetch(
      `${API_V1_BASE}/cogs/summary?${params.toString()}`,
      {
        headers: this.getAuthHeaders(),
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to get COGS summary');
    }

    const result = await response.json();
    return result.data;
  }

  /**
   * Get list of sales without COGS entries
   */
  async getSalesWithoutCOGS(
    startDate: string,
    endDate: string,
    page: number = 1,
    limit: number = 20
  ): Promise<{
    sales: SaleWithoutCOGS[];
    total_count: number;
    page: number;
    limit: number;
    total_pages: number;
  }> {
    const params = new URLSearchParams({
      start_date: startDate,
      end_date: endDate,
      page: page.toString(),
      limit: limit.toString(),
    });

    const response = await fetch(
      `${API_V1_BASE}/cogs/missing?${params.toString()}`,
      {
        headers: this.getAuthHeaders(),
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to get sales without COGS');
    }

    const result = await response.json();
    return result.data;
  }

  /**
   * Backfill COGS entries for existing sales
   * @param startDate Start date (YYYY-MM-DD)
   * @param endDate End date (YYYY-MM-DD)
   * @param dryRun If true, only preview without making changes
   */
  async backfillCOGS(
    startDate: string,
    endDate: string,
    dryRun: boolean = false
  ): Promise<BackfillResponse> {
    const params = new URLSearchParams({
      start_date: startDate,
      end_date: endDate,
      dry_run: dryRun.toString(),
    });

    const response = await fetch(
      `${API_V1_BASE}/cogs/backfill?${params.toString()}`,
      {
        method: 'POST',
        headers: this.getAuthHeaders(),
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to backfill COGS');
    }

    const result = await response.json();
    return {
      status: result.status,
      dry_run: result.dry_run,
      sales_to_process: result.sales_to_process,
      sales_processed: result.sales_processed,
      total_estimated_cogs: result.total_estimated_cogs,
      preview: result.preview,
      message: result.message,
    };
  }

  /**
   * Record COGS for a specific sale
   */
  async recordCOGSForSale(saleId: number): Promise<{
    status: string;
    message: string;
  }> {
    const response = await fetch(
      `${API_V1_BASE}/cogs/record/${saleId}`,
      {
        method: 'POST',
        headers: this.getAuthHeaders(),
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to record COGS for sale');
    }

    return await response.json();
  }

  /**
   * Check if there are sales without COGS in a period
   */
  async hasMissingCOGS(startDate: string, endDate: string): Promise<boolean> {
    try {
      const result = await this.getSalesWithoutCOGS(startDate, endDate, 1, 1);
      return result.total_count > 0;
    } catch (error) {
      console.error('Error checking for missing COGS:', error);
      return false;
    }
  }

  /**
   * Get COGS health status for a period
   * Returns information about COGS completeness and accuracy
   */
  async getCOGSHealthStatus(startDate: string, endDate: string): Promise<{
    healthy: boolean;
    total_sales: number;
    sales_with_cogs: number;
    sales_without_cogs: number;
    completeness_percentage: number;
    avg_cogs_percentage: number;
    message: string;
  }> {
    try {
      const [summary, missing] = await Promise.all([
        this.getCOGSSummary(startDate, endDate),
        this.getSalesWithoutCOGS(startDate, endDate, 1, 1),
      ]);

      const salesWithCOGS = summary.total_sales - missing.total_count;
      const completenessPercentage = summary.total_sales > 0
        ? (salesWithCOGS / summary.total_sales) * 100
        : 100;

      const healthy = completenessPercentage >= 95 && summary.cogs_percentage > 0;

      let message = '';
      if (completenessPercentage < 95) {
        message = `⚠️ ${missing.total_count} sales are missing COGS entries. Consider running backfill.`;
      } else if (summary.cogs_percentage === 0) {
        message = '⚠️ No COGS data found. This may indicate a data issue.';
      } else {
        message = '✅ COGS data is complete and healthy.';
      }

      return {
        healthy,
        total_sales: summary.total_sales,
        sales_with_cogs: salesWithCOGS,
        sales_without_cogs: missing.total_count,
        completeness_percentage: completenessPercentage,
        avg_cogs_percentage: summary.cogs_percentage,
        message,
      };
    } catch (error) {
      console.error('Error getting COGS health status:', error);
      throw error;
    }
  }
}

export const cogsService = new COGSService();

