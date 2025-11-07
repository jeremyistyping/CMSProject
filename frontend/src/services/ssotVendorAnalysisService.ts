import api from './api';
import { API_ENDPOINTS } from '../config/api';

export interface SSOTVendorAnalysisData {
  company?: CompanyInfo;
  start_date: string;
  end_date: string;
  currency: string;
  total_vendors: number;
  active_vendors: number;
  total_purchases: number;
  total_payments: number;
  outstanding_payables: number;
  vendors_by_performance: VendorPerformanceData[];
  payment_analysis: PaymentAnalysisData;
  top_vendors_by_spend: VendorSpendData[];
  vendor_payment_history: VendorPaymentHistory[];
  generated_at: string;
}

export interface CompanyInfo {
  name: string;
  address: string;
  city: string;
  state: string;
  phone: string;
  email: string;
  tax_number: string;
}

export interface VendorPerformanceData {
  vendor_id: number;
  vendor_name: string;
  total_transactions: number;
  on_time_payments: number;
  late_payments: number;
  performance_score: number;
}

export interface PaymentAnalysisData {
  on_time_payments: number;
  late_payments: number;
  average_pay_days: number;
  total_discounts: number;
  penalty_fees: number;
}

export interface VendorSpendData {
  vendor_id: number;
  vendor_name: string;
  total_spend: number;
  percentage_of_total: number;
  transaction_count: number;
}

export interface VendorPaymentHistory {
  vendor_id: number;
  vendor_name: string;
  payment_date: string;
  amount: number;
  payment_method: string;
  status: string;
}

export interface SSOTVendorAnalysisParams {
  start_date: string;
  end_date: string;
  format?: 'json' | 'pdf';
}

class SSOTVendorAnalysisService {

  async generateSSOTVendorAnalysis(params: SSOTVendorAnalysisParams): Promise<SSOTVendorAnalysisData> {
    try {
      const queryParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
        format: params.format || 'json'
      });

      const response = await api.get(API_ENDPOINTS.SSOT_REPORTS.VENDOR_ANALYSIS + `?${queryParams}`);

      if (response.data.status === 'success') {
        return response.data.data;
      }

      throw new Error(response.data.message || 'Failed to generate vendor analysis');
    } catch (error: any) {
      console.error('Error generating SSOT vendor analysis:', error);
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      }
      throw error;
    }
  }

  formatCurrency(value: number): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR'
    }).format(value);
  }

  calculatePerformanceScore(onTime: number, late: number): number {
    const total = onTime + late;
    if (total === 0) return 100;
    return Math.round((onTime / total) * 100);
  }
}

export const ssotVendorAnalysisService = new SSOTVendorAnalysisService();