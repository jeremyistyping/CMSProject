import api from './api';
import { API_ENDPOINTS } from '../config/api';

export interface FinancialReportRequest {
  report_type: string;
  start_date: Date;
  end_date: Date;
  period?: string;
  comparative?: boolean;
  show_zero?: boolean;
  format?: string;
}

export interface AccountLineItem {
  accountId: number;
  accountCode: string;
  accountName: string;
  accountType: string;
  category: string;
  balance: number;
}

export interface ReportHeader {
  reportType: string;
  companyName: string;
  reportTitle: string;
  startDate: Date;
  endDate: Date;
  generatedAt: Date;
  currency: string;
  isComparative: boolean;
}

export interface ProfitLossStatement {
  reportHeader: ReportHeader;
  revenue: AccountLineItem[];
  totalRevenue: number;
  cogs: AccountLineItem[];
  totalCOGS: number;
  grossProfit: number;
  expenses: AccountLineItem[];
  totalExpenses: number;
  netIncome: number;
  comparative?: {
    previousPeriod: ProfitLossStatement;
    variance: {
      revenueVariance: number;
      cogsVariance: number;
      grossProfitVariance: number;
      expenseVariance: number;
      netIncomeVariance: number;
    };
  };
}

export interface BalanceSheetCategory {
  name: string;
  accounts: AccountLineItem[];
  total: number;
}

export interface BalanceSheetSection {
  categories: BalanceSheetCategory[];
  total: number;
}

export interface BalanceSheet {
  reportHeader: ReportHeader;
  assets: BalanceSheetSection;
  liabilities: BalanceSheetSection;
  equity: BalanceSheetSection;
  totalAssets: number;
  totalLiabilities: number;
  totalEquity: number;
  isBalanced: boolean;
  comparative?: {
    previousPeriod: BalanceSheet;
    variance: {
      assetsVariance: number;
      liabilitiesVariance: number;
      equityVariance: number;
    };
  };
}

export interface CashFlowItem {
  description: string;
  amount: number;
}

export interface CashFlowSection {
  items: CashFlowItem[];
  total: number;
}

export interface CashFlowStatement {
  reportHeader: ReportHeader;
  operatingActivities: CashFlowSection;
  investingActivities: CashFlowSection;
  financingActivities: CashFlowSection;
  netCashFlow: number;
  beginningCash: number;
  endingCash: number;
}

export interface TrialBalanceItem {
  accountId: number;
  accountCode: string;
  accountName: string;
  accountType: string;
  debitBalance: number;
  creditBalance: number;
}

export interface TrialBalance {
  reportHeader: ReportHeader;
  accounts: TrialBalanceItem[];
  totalDebits: number;
  totalCredits: number;
  isBalanced: boolean;
}

export interface GeneralLedgerEntry {
  date: Date;
  journalCode: string;
  description: string;
  reference: string;
  debitAmount: number;
  creditAmount: number;
  balance: number;
}

export interface GeneralLedger {
  reportHeader: ReportHeader;
  account: AccountLineItem;
  transactions: GeneralLedgerEntry[];
  beginningBalance: number;
  endingBalance: number;
  totalDebits: number;
  totalCredits: number;
}

export interface FinancialRatios {
  // Liquidity Ratios
  currentRatio: number;
  quickRatio: number;
  cashRatio: number;
  
  // Profitability Ratios
  grossProfitMargin: number;
  netProfitMargin: number;
  roa: number; // Return on Assets
  roe: number; // Return on Equity
  
  // Leverage Ratios
  debtRatio: number;
  debtToEquityRatio: number;
  
  // Efficiency Ratios
  assetTurnover: number;
  inventoryTurnover: number;
  
  calculatedAt: Date;
  periodStart: Date;
  periodEnd: Date;
}

export interface RealTimeFinancialMetrics {
  asOfDate: Date;
  cashPosition: number;
  dailyRevenue: number;
  dailyExpenses: number;
  dailyNetIncome: number;
  monthlyRevenue: number;
  monthlyExpenses: number;
  monthlyNetIncome: number;
  yearlyRevenue: number;
  yearlyExpenses: number;
  yearlyNetIncome: number;
  pendingReceivables: number;
  pendingPayables: number;
  inventoryValue: number;
  lastUpdated: Date;
}

export interface FinancialAlert {
  type: string;
  severity: string;
  title: string;
  description: string;
  value: number;
  threshold: number;
  createdAt: Date;
}

export interface FinancialHealthComponents {
  liquidityScore: number;
  profitabilityScore: number;
  leverageScore: number;
  efficiencyScore: number;
  growthScore: number;
}

export interface HealthRecommendation {
  category: string;
  priority: string;
  title: string;
  description: string;
  action: string;
}

export interface FinancialHealthScore {
  overallScore: number;
  scoreGrade: string;
  components: FinancialHealthComponents;
  recommendations: HealthRecommendation[];
  calculatedAt: Date;
}

export interface FinancialDashboard {
  reportDate: Date;
  keyMetrics: {
    totalRevenue: number;
    totalExpenses: number;
    netIncome: number;
    totalAssets: number;
    totalLiabilities: number;
    totalEquity: number;
    cashBalance: number;
    accountsReceivable: number;
    accountsPayable: number;
    inventory: number;
  };
  ratios: FinancialRatios;
  cashPosition: {
    totalCash: number;
    cashAccounts: any[];
    bankAccounts: any[];
    cashFlow30Day: number;
  };
  accountBalances: any[];
  recentActivity: any[];
  alerts: FinancialAlert[];
}

export interface ReportMetadata {
  reportType: string;
  name: string;
  description: string;
  supportsComparative: boolean;
  requiredParams: string[];
  optionalParams: string[];
}

export interface QuickFinancialStats {
  cashBalance: number;
  todayRevenue: number;
  todayExpenses: number;
  monthToDateRevenue: number;
  monthToDateExpenses: number;
  yearToDateRevenue: number;
  yearToDateExpenses: number;
  pendingReceivables: number;
  pendingPayables: number;
  inventoryValue: number;
  lastUpdated: Date;
}

export interface ReportExportRequest {
  reportType: string;
  reportData: any;
  format: string; // 'PDF', 'EXCEL', 'CSV'
  filename?: string;
}

class FinancialReportService {
  // Core Financial Reports
  async generateProfitLossStatement(request: FinancialReportRequest): Promise<ProfitLossStatement> {
    const response = await api.post(API_ENDPOINTS.REPORTS.ENHANCED.PROFIT_LOSS, request);
    return response.data.data;
  }

  async generateBalanceSheet(request: FinancialReportRequest): Promise<BalanceSheet> {
    const response = await api.post(API_ENDPOINTS.REPORTS.ENHANCED.BALANCE_SHEET, request);
    return response.data.data;
  }

  async generateCashFlowStatement(request: FinancialReportRequest): Promise<CashFlowStatement> {
    const response = await api.post(API_ENDPOINTS.REPORTS.ENHANCED.CASH_FLOW, request);
    return response.data.data;
  }

  async generateTrialBalance(request: FinancialReportRequest): Promise<TrialBalance> {
    const response = await api.post(API_ENDPOINTS.REPORTS.ENHANCED.TRIAL_BALANCE, request);
    return response.data.data;
  }

  async generateGeneralLedger(
    accountId: number, 
    startDate: string, 
    endDate: string
  ): Promise<GeneralLedger> {
    const response = await api.get(API_ENDPOINTS.REPORTS.ENHANCED.GENERAL_LEDGER(accountId), {
      params: { start_date: startDate, end_date: endDate }
    });
    return response.data.data;
  }

  // Advanced Reports and Analytics
  async getFinancialDashboard(): Promise<FinancialDashboard> {
    const response = await api.get(API_ENDPOINTS.REPORTS.ENHANCED.DASHBOARD);
    return response.data.data;
  }

  async getRealTimeMetrics(): Promise<RealTimeFinancialMetrics> {
    const response = await api.get(API_ENDPOINTS.REPORTS.ENHANCED.REAL_TIME_METRICS);
    return response.data.data;
  }

  async calculateFinancialRatios(startDate: string, endDate: string): Promise<FinancialRatios> {
    const response = await api.get(API_ENDPOINTS.REPORTS.ENHANCED.FINANCIAL_RATIOS, {
      params: { start_date: startDate, end_date: endDate }
    });
    return response.data.data;
  }

  async getFinancialHealthScore(): Promise<FinancialHealthScore> {
    const response = await api.get(API_ENDPOINTS.REPORTS.ENHANCED.HEALTH_SCORE);
    return response.data.data;
  }

  // Utility Functions
  async getReportsList(): Promise<ReportMetadata[]> {
    const response = await api.get(API_ENDPOINTS.REPORTS.ENHANCED.LIST);
    return response.data.data;
  }

  async validateReportRequest(request: FinancialReportRequest): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
  }> {
    const response = await api.post(API_ENDPOINTS.REPORTS.ENHANCED.VALIDATE, request);
    return response.data.data;
  }

  async getReportFormats(): Promise<string[]> {
    const response = await api.get(API_ENDPOINTS.REPORTS.ENHANCED.EXPORT_FORMATS);
    return response.data.data;
  }

  async getReportSummary(limit?: number): Promise<any> {
    const response = await api.get(API_ENDPOINTS.REPORTS.ENHANCED.SUMMARY, {
      params: limit ? { limit } : undefined
    });
    return response.data.data;
  }

  async getQuickStats(): Promise<QuickFinancialStats> {
    const response = await api.get(API_ENDPOINTS.REPORTS.ENHANCED.QUICK_STATS);
    return response.data.data;
  }

  async exportReport(exportRequest: ReportExportRequest): Promise<Blob> {
    const response = await api.post(API_ENDPOINTS.REPORTS.ENHANCED.EXPORT, exportRequest, {
      responseType: 'blob'
    });
    return response.data;
  }

  // Utility helper methods
  formatCurrency(amount: number, currency: string = 'IDR'): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  }

  formatPercentage(value: number, decimals: number = 2): string {
    return `${value.toFixed(decimals)}%`;
  }

  formatDate(date: Date | string): string {
    const d = typeof date === 'string' ? new Date(date) : date;
    return d.toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }

  formatDateRange(startDate: Date | string, endDate: Date | string): string {
    return `${this.formatDate(startDate)} - ${this.formatDate(endDate)}`;
  }

  calculateVariancePercentage(current: number, previous: number): number {
    if (previous === 0) return current === 0 ? 0 : 100;
    return ((current - previous) / Math.abs(previous)) * 100;
  }

  getVarianceColor(variance: number): string {
    if (variance > 0) return 'green.500';
    if (variance < 0) return 'red.500';
    return 'gray.500';
  }

  getHealthScoreColor(score: number): string {
    if (score >= 90) return 'green.500';
    if (score >= 80) return 'blue.500';
    if (score >= 70) return 'yellow.500';
    if (score >= 60) return 'orange.500';
    return 'red.500';
  }

  getAlertSeverityColor(severity: string): string {
    switch (severity.toUpperCase()) {
      case 'HIGH': return 'red.500';
      case 'MEDIUM': return 'orange.500';
      case 'LOW': return 'yellow.500';
      default: return 'gray.500';
    }
  }
}

// Export a singleton instance
const financialReportService = new FinancialReportService();
export default financialReportService;
