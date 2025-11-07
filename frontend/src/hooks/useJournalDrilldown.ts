import { useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';

interface JournalDrilldownRequest {
  account_codes?: string[];
  account_ids?: number[];  // Will be converted to uint in backend
  start_date: string;
  end_date: string;
  report_type?: string;
  line_item_name?: string;
  min_amount?: number;
  max_amount?: number;
  transaction_types?: string[];
  page: number;
  limit: number;
}

interface JournalDrilldownHookState {
  isOpen: boolean;
  loading: boolean;
  error: string | null;
  drilldownRequest: JournalDrilldownRequest | null;
  title: string;
}

export const useJournalDrilldown = () => {
  const { token } = useAuth();
  const [state, setState] = useState<JournalDrilldownHookState>({
    isOpen: false,
    loading: false,
    error: null,
    drilldownRequest: null,
    title: 'Journal Entry Details',
  });

  const openDrilldown = (request: Partial<JournalDrilldownRequest>, customTitle?: string) => {
    const fullRequest: JournalDrilldownRequest = {
      start_date: '',
      end_date: '',
      page: 1,
      limit: 20,
      ...request,
    };

    // Validate required fields
    if (!fullRequest.start_date || !fullRequest.end_date) {
      setState(prev => ({
        ...prev,
        error: 'Start date and end date are required',
      }));
      return;
    }

    setState(prev => ({
      ...prev,
      isOpen: true,
      drilldownRequest: fullRequest,
      title: customTitle || 'Journal Entry Details',
      error: null,
    }));
  };

  const closeDrilldown = () => {
    setState(prev => ({
      ...prev,
      isOpen: false,
      drilldownRequest: null,
      error: null,
    }));
  };

  // Helper function to create drill-down for specific line items
  const drillDownByLineItem = (
    lineItemName: string,
    accountCodes: string[] = [],
    accountIds: number[] = [],
    startDate: string,
    endDate: string,
    reportType: string = 'BALANCE_SHEET',
    minAmount?: number,
    maxAmount?: number,
    transactionTypes?: string[]
  ) => {
    openDrilldown({
      account_codes: accountCodes.length > 0 ? accountCodes : undefined,
      account_ids: accountIds.length > 0 ? accountIds : undefined,
      start_date: startDate,
      end_date: endDate,
      report_type: reportType,
      line_item_name: lineItemName,
      min_amount: minAmount,
      max_amount: maxAmount,
      transaction_types: transactionTypes,
    }, `Journal Entries - ${lineItemName}`);
  };

  // Helper for Balance Sheet drill-down
  const drillDownBalanceSheet = (
    lineItemName: string,
    accountCodes: string[] = [],
    accountIds: number[] = [],
    asOfDate: string
  ) => {
    const startDate = new Date(asOfDate);
    startDate.setFullYear(startDate.getFullYear() - 1); // One year ago
    
    drillDownByLineItem(
      lineItemName,
      accountCodes,
      accountIds,
      startDate.toISOString().split('T')[0],
      asOfDate,
      'BALANCE_SHEET'
    );
  };

  // Helper for Profit & Loss drill-down
  const drillDownProfitLoss = (
    lineItemName: string,
    accountCodes: string[] = [],
    accountIds: number[] = [],
    startDate: string,
    endDate: string
  ) => {
    drillDownByLineItem(
      lineItemName,
      accountCodes,
      accountIds,
      startDate,
      endDate,
      'PROFIT_LOSS'
    );
  };

  // Helper for Cash Flow drill-down
  const drillDownCashFlow = (
    lineItemName: string,
    accountCodes: string[] = [],
    accountIds: number[] = [],
    startDate: string,
    endDate: string,
    transactionTypes?: string[]
  ) => {
    drillDownByLineItem(
      lineItemName,
      accountCodes,
      accountIds,
      startDate,
      endDate,
      'CASH_FLOW',
      undefined,
      undefined,
      transactionTypes
    );
  };

  // Helper for specific account drill-down
  const drillDownAccount = (
    accountCode: string,
    accountName: string,
    startDate: string,
    endDate: string,
    reportType: string = 'GENERAL_LEDGER'
  ) => {
    openDrilldown({
      account_codes: [accountCode],
      start_date: startDate,
      end_date: endDate,
      report_type: reportType,
      line_item_name: `${accountCode} - ${accountName}`,
    }, `Journal Entries - ${accountCode} - ${accountName}`);
  };

  // Helper for amount range drill-down
  const drillDownAmountRange = (
    minAmount: number,
    maxAmount: number,
    startDate: string,
    endDate: string,
    lineItemName: string = 'Amount Range Filter'
  ) => {
    openDrilldown({
      start_date: startDate,
      end_date: endDate,
      min_amount: minAmount,
      max_amount: maxAmount,
      line_item_name: lineItemName,
    }, `Journal Entries - ${lineItemName}`);
  };

  // Helper for transaction type drill-down
  const drillDownTransactionType = (
    transactionTypes: string[],
    startDate: string,
    endDate: string,
    lineItemName: string = 'Transaction Type Filter'
  ) => {
    openDrilldown({
      start_date: startDate,
      end_date: endDate,
      transaction_types: transactionTypes,
      line_item_name: lineItemName,
    }, `Journal Entries - ${lineItemName}`);
  };

  return {
    // State
    isOpen: state.isOpen,
    loading: state.loading,
    error: state.error,
    drilldownRequest: state.drilldownRequest,
    title: state.title,
    
    // Actions
    openDrilldown,
    closeDrilldown,
    
    // Helper methods
    drillDownByLineItem,
    drillDownBalanceSheet,
    drillDownProfitLoss,
    drillDownCashFlow,
    drillDownAccount,
    drillDownAmountRange,
    drillDownTransactionType,
  };
};
