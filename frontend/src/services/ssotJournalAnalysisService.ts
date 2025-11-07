import api from './api';
import { API_ENDPOINTS } from '../config/api';

export interface SSOTJournalAnalysisData {
  company?: CompanyInfo;
  start_date: string;
  end_date: string;
  currency: string;
  total_entries: number;
  posted_entries: number;
  draft_entries: number;
  reversed_entries: number;
  total_amount: number;
  entries_by_type: EntryTypeBreakdown[];
  entries_by_account: AccountBreakdown[];
  entries_by_period: PeriodBreakdown[];
  compliance_check: ComplianceReport;
  data_quality_metrics: DataQualityMetrics;
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

export interface EntryTypeBreakdown {
  source_type: string;
  count: number;
  total_amount: number;
  percentage: number;
}

export interface AccountBreakdown {
  account_id: number;
  account_code: string;
  account_name: string;
  count: number;
  total_debit: number;
  total_credit: number;
}

export interface PeriodBreakdown {
  period: string;
  start_date: string;
  end_date: string;
  count: number;
  total_amount: number;
}

export interface ComplianceReport {
  total_checks: number;
  passed_checks: number;
  failed_checks: number;
  compliance_score: number;
  issues: ComplianceIssue[];
  recommendations: string[];
}

export interface ComplianceIssue {
  type: string;
  description: string;
  severity: string;
  journal_id: number;
}

export interface DataQualityMetrics {
  overall_score: number;
  completeness_score: number;
  accuracy_score: number;
  consistency_score: number;
  issues: DataQualityIssue[];
  detailed_metrics: any;
}

export interface DataQualityIssue {
  type: string;
  description: string;
  count: number;
  severity: string;
}

export interface SSOTJournalAnalysisParams {
  start_date: string;
  end_date: string;
  format?: 'json' | 'pdf';
}

class SSOTJournalAnalysisService {

  async generateSSOTJournalAnalysis(params: SSOTJournalAnalysisParams): Promise<SSOTJournalAnalysisData> {
    try {
      const queryParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
        format: params.format || 'json'
      });

      const response = await api.get(API_ENDPOINTS.SSOT_REPORTS.JOURNAL_ANALYSIS + `?${queryParams}`);

      if (response.data.status === 'success') {
        return response.data.data;
      }

      throw new Error(response.data.message || 'Failed to generate journal analysis');
    } catch (error: any) {
      console.error('Error generating SSOT journal analysis:', error);
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

  getSeverityColor(severity: string): string {
    const colors: { [key: string]: string } = {
      'critical': 'red',
      'high': 'orange',
      'medium': 'yellow',
      'low': 'blue',
      'info': 'gray'
    };
    return colors[severity.toLowerCase()] || 'gray';
  }

  calculateCompliancePercentage(passed: number, total: number): number {
    if (total === 0) return 100;
    return Math.round((passed / total) * 100);
  }

  getQualityGrade(score: number): string {
    if (score >= 90) return 'A';
    if (score >= 80) return 'B';
    if (score >= 70) return 'C';
    if (score >= 60) return 'D';
    return 'F';
  }
}

export const ssotJournalAnalysisService = new SSOTJournalAnalysisService();