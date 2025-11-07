import { API_V1_BASE } from '../config/api';
import { getAuthHeaders } from '../utils/authTokenUtils';
import { ssotCashFlowReportService } from './ssotCashFlowReportService';

export interface CashFlowExportParams {
  start_date: string;
  end_date: string;
}

class CashFlowExportService {
  private getAuthHeaders() {
    return getAuthHeaders();
  }

  /**
   * Export Cash Flow Statement as CSV
   * @param params Export parameters
   * @returns Promise<void> (triggers download)
   */
  async exportToCSV(params: CashFlowExportParams): Promise<void> {
    try {
      if (!params.start_date || !params.end_date) {
        throw new Error('Start date and end date are required for export');
      }

      const queryParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
        format: 'csv'
      });

      const url = `${API_V1_BASE}/reports/ssot/cash-flow?${queryParams.toString()}`;
      
      const response = await fetch(url, {
        method: 'GET',
        headers: this.getAuthHeaders(),
      });

      if (!response.ok) {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
        } catch {
          // Use default error message if JSON parsing fails
        }
        throw new Error(errorMessage);
      }

      // Get the blob data
      const blob = await response.blob();
      if (blob.size === 0) {
        throw new Error('Empty file received from server');
      }

      // Generate filename
      const filename = this.generateCSVFilename(params.start_date, params.end_date);
      
      // Trigger download
      await this.downloadFile(blob, filename);

    } catch (error) {
      console.error('CSV export error:', error);
      throw new Error(`Failed to export Cash Flow as CSV: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Export Cash Flow Statement as PDF
   * @param params Export parameters
   * @returns Promise<void> (triggers download)
   */
  async exportToPDF(params: CashFlowExportParams): Promise<void> {
    try {
      if (!params.start_date || !params.end_date) {
        throw new Error('Start date and end date are required for export');
      }

      const queryParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
        format: 'pdf'
      });

      const url = `${API_V1_BASE}/reports/ssot/cash-flow?${queryParams.toString()}`;
      
      const response = await fetch(url, {
        method: 'GET',
        headers: this.getAuthHeaders(),
      });

      if (!response.ok) {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
        } catch {
          // Use default error message if JSON parsing fails
        }
        throw new Error(errorMessage);
      }

      // Get the blob data
      const blob = await response.blob();
      if (blob.size === 0) {
        throw new Error('Empty file received from server');
      }

      // Generate filename
      const filename = this.generatePDFFilename(params.start_date, params.end_date);
      
      // Trigger download
      await this.downloadFile(blob, filename);

    } catch (error) {
      console.error('PDF export error:', error);
      throw new Error(`Failed to export Cash Flow as PDF: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Download file to user's device
   * @param blob File blob data
   * @param filename Name for the downloaded file
   */
  private async downloadFile(blob: Blob, filename: string): Promise<void> {
    try {
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      
      // Append to body, click, then remove
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      // Clean up the object URL
      window.URL.revokeObjectURL(url);
    } catch (error) {
      throw new Error(`Failed to download file: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Generate CSV filename
   * @param startDate Start date
   * @param endDate End date
   * @returns CSV filename
   */
  private generateCSVFilename(startDate: string, endDate: string): string {
    const start = startDate.replace(/-/g, '');
    const end = endDate.replace(/-/g, '');
    return `cash_flow_${start}_to_${end}.csv`;
  }

  /**
   * Generate PDF filename
   * @param startDate Start date
   * @param endDate End date
   * @returns PDF filename
   */
  private generatePDFFilename(startDate: string, endDate: string): string {
    const start = startDate.replace(/-/g, '');
    const end = endDate.replace(/-/g, '');
    return `cash_flow_${start}_to_${end}.pdf`;
  }

  /**
   * Format date for display
   * @param dateString Date string
   * @returns Formatted date
   */
  formatDate(dateString: string): string {
    try {
      const date = new Date(dateString);
      return date.toLocaleDateString('id-ID', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      });
    } catch {
      return dateString;
    }
  }

  /**
   * Format date range for display
   * @param startDate Start date
   * @param endDate End date
   * @returns Formatted date range
   */
  formatDateRange(startDate: string, endDate: string): string {
    return `${this.formatDate(startDate)} - ${this.formatDate(endDate)}`;
  }

  /**
   * Validate export parameters
   * @param params Export parameters
   * @returns Validation result
   */
  validateExportParams(params: CashFlowExportParams): { valid: boolean; errors: string[] } {
    const errors: string[] = [];

    if (!params.start_date) {
      errors.push('Start date is required');
    }

    if (!params.end_date) {
      errors.push('End date is required');
    }

    if (params.start_date && params.end_date) {
      const startDate = new Date(params.start_date);
      const endDate = new Date(params.end_date);

      if (isNaN(startDate.getTime())) {
        errors.push('Invalid start date format');
      }

      if (isNaN(endDate.getTime())) {
        errors.push('Invalid end date format');
      }

      if (startDate > endDate) {
        errors.push('Start date cannot be later than end date');
      }
    }

    return {
      valid: errors.length === 0,
      errors
    };
  }

  /**
   * Get available export formats
   * @returns Array of available formats
   */
  getAvailableFormats(): Array<{ value: string; label: string; description: string }> {
    return [
      {
        value: 'csv',
        label: 'CSV',
        description: 'Comma-separated values file for spreadsheet applications'
      },
      {
        value: 'pdf',
        label: 'PDF',
        description: 'Portable Document Format for professional presentation'
      }
    ];
  }
}

// Export singleton instance
export const cashFlowExportService = new CashFlowExportService();
export default cashFlowExportService;