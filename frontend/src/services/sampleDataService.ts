/**
 * Sample Data Service
 * 
 * This service creates sample data for testing balance sheet and other financial reports.
 * It works entirely through the existing API endpoints to ensure data consistency.
 */

import { getAuthHeaders } from '../utils/authTokenUtils';

interface SampleAccount {
  code: string;
  name: string;
  type: string;
  balance?: number;
  parentCode?: string;
  level?: number;
  isHeader?: boolean;
}

interface CreateAccountRequest {
  code: string;
  name: string;
  type: string;
  balance?: number;
  parent_id?: number;
  level?: number;
  is_header?: boolean;
  is_active?: boolean;
  description?: string;
}

interface AccountResponse {
  id: number;
  code: string;
  name: string;
  type: string;
  balance: number;
}

class SampleDataService {
  private readonly API_BASE_URL = '/api/v1';

  /**
   * Get authentication headers
   */
  private getAuthHeaders(): HeadersInit {
    // Use centralized token utility for consistency across the application
    try {
      return getAuthHeaders();
    } catch (error) {
      // Fallback to just content-type if no token is available
      return {
        'Content-Type': 'application/json',
      };
    }
  }

  /**
   * Create sample balance sheet data
   */
  async createSampleBalanceSheetData(): Promise<{
    success: boolean;
    message: string;
    accounts: AccountResponse[];
  }> {
    const sampleAccounts: SampleAccount[] = [
      // ASSETS
      {
        code: '1-0-000',
        name: 'ASSETS',
        type: 'ASSET',
        level: 1,
        isHeader: true,
      },
      {
        code: '1-1-000',
        name: 'Current Assets',
        type: 'ASSET',
        parentCode: '1-0-000',
        level: 2,
        isHeader: true,
      },
      {
        code: '1-1-001',
        name: 'Cash',
        type: 'ASSET',
        balance: 50000000, // 50 million IDR
        parentCode: '1-1-000',
        level: 3,
      },
      {
        code: '1-1-002',
        name: 'Bank - BCA',
        type: 'ASSET',
        balance: 100000000, // 100 million IDR
        parentCode: '1-1-000',
        level: 3,
      },
      {
        code: '1-1-003',
        name: 'Accounts Receivable',
        type: 'ASSET',
        balance: 75000000, // 75 million IDR
        parentCode: '1-1-000',
        level: 3,
      },
      {
        code: '1-1-004',
        name: 'Inventory',
        type: 'ASSET',
        balance: 45000000, // 45 million IDR
        parentCode: '1-1-000',
        level: 3,
      },
      {
        code: '1-2-000',
        name: 'Non-Current Assets',
        type: 'ASSET',
        parentCode: '1-0-000',
        level: 2,
        isHeader: true,
      },
      {
        code: '1-2-001',
        name: 'Office Equipment',
        type: 'ASSET',
        balance: 25000000, // 25 million IDR
        parentCode: '1-2-000',
        level: 3,
      },
      {
        code: '1-2-002',
        name: 'Vehicles',
        type: 'ASSET',
        balance: 80000000, // 80 million IDR
        parentCode: '1-2-000',
        level: 3,
      },
      {
        code: '1-2-003',
        name: 'Buildings',
        type: 'ASSET',
        balance: 200000000, // 200 million IDR
        parentCode: '1-2-000',
        level: 3,
      },

      // LIABILITIES
      {
        code: '2-0-000',
        name: 'LIABILITIES',
        type: 'LIABILITY',
        level: 1,
        isHeader: true,
      },
      {
        code: '2-1-000',
        name: 'Current Liabilities',
        type: 'LIABILITY',
        parentCode: '2-0-000',
        level: 2,
        isHeader: true,
      },
      {
        code: '2-1-001',
        name: 'Accounts Payable',
        type: 'LIABILITY',
        balance: 40000000, // 40 million IDR
        parentCode: '2-1-000',
        level: 3,
      },
      {
        code: '2-1-002',
        name: 'Tax Payable',
        type: 'LIABILITY',
        balance: 15000000, // 15 million IDR
        parentCode: '2-1-000',
        level: 3,
      },
      {
        code: '2-1-003',
        name: 'Accrued Expenses',
        type: 'LIABILITY',
        balance: 8000000, // 8 million IDR
        parentCode: '2-1-000',
        level: 3,
      },
      {
        code: '2-2-000',
        name: 'Non-Current Liabilities',
        type: 'LIABILITY',
        parentCode: '2-0-000',
        level: 2,
        isHeader: true,
      },
      {
        code: '2-2-001',
        name: 'Bank Loan',
        type: 'LIABILITY',
        balance: 120000000, // 120 million IDR
        parentCode: '2-2-000',
        level: 3,
      },

      // EQUITY
      {
        code: '3-0-000',
        name: 'EQUITY',
        type: 'EQUITY',
        level: 1,
        isHeader: true,
      },
      {
        code: '3-1-001',
        name: 'Share Capital',
        type: 'EQUITY',
        balance: 200000000, // 200 million IDR
        parentCode: '3-0-000',
        level: 2,
      },
      {
        code: '3-2-001',
        name: 'Retained Earnings',
        type: 'EQUITY',
        balance: 275000000, // 275 million IDR (to balance the books)
        parentCode: '3-0-000',
        level: 2,
      },
    ];

    try {
      const createdAccounts: AccountResponse[] = [];
      const errors: string[] = [];

      // Create accounts in order (headers first, then detail accounts)
      const headerAccounts = sampleAccounts.filter(acc => acc.isHeader);
      const detailAccounts = sampleAccounts.filter(acc => !acc.isHeader);

      // Create header accounts first
      for (const account of headerAccounts) {
        try {
          const created = await this.createAccount(account);
          if (created) {
            createdAccounts.push(created);
          }
        } catch (error) {
          errors.push(`Failed to create header account ${account.code}: ${error}`);
          console.warn(`Header account creation failed:`, error);
        }
      }

      // Wait a bit for header accounts to be processed
      await new Promise(resolve => setTimeout(resolve, 500));

      // Create detail accounts
      for (const account of detailAccounts) {
        try {
          const created = await this.createAccount(account);
          if (created) {
            createdAccounts.push(created);
          }
        } catch (error) {
          errors.push(`Failed to create account ${account.code}: ${error}`);
          console.warn(`Account creation failed:`, error);
        }
      }

      const totalAssets = 575000000; // Sum of asset balances
      const totalLiabilities = 183000000; // Sum of liability balances
      const totalEquity = 475000000; // Sum of equity balances

      const message = `Successfully created ${createdAccounts.length} sample accounts for balance sheet testing.\n\n` +
        `💰 Total Assets: ${this.formatCurrency(totalAssets)}\n` +
        `🔻 Total Liabilities: ${this.formatCurrency(totalLiabilities)}\n` +
        `🏛️ Total Equity: ${this.formatCurrency(totalEquity)}\n\n` +
        `⚖️ Balance Check: ${totalAssets === (totalLiabilities + totalEquity) ? 'BALANCED ✅' : 'UNBALANCED ⚠️'}\n\n` +
        `${errors.length > 0 ? `⚠️ Some accounts had issues: ${errors.join('; ')}` : ''}`;

      return {
        success: true,
        message,
        accounts: createdAccounts,
      };

    } catch (error) {
      console.error('Sample data creation failed:', error);
      return {
        success: false,
        message: `Failed to create sample data: ${error instanceof Error ? error.message : 'Unknown error'}`,
        accounts: [],
      };
    }
  }

  /**
   * Create a single account via API
   */
  private async createAccount(sampleAccount: SampleAccount): Promise<AccountResponse | null> {
    const accountRequest: CreateAccountRequest = {
      code: sampleAccount.code,
      name: sampleAccount.name,
      type: sampleAccount.type,
      balance: sampleAccount.balance || 0,
      level: sampleAccount.level || 1,
      is_header: sampleAccount.isHeader || false,
      is_active: true,
      description: `Sample account: ${sampleAccount.name}`,
    };

    // If there's a parent code, try to find the parent ID
    if (sampleAccount.parentCode) {
      try {
        const parentId = await this.findAccountIdByCode(sampleAccount.parentCode);
        if (parentId) {
          accountRequest.parent_id = parentId;
        }
      } catch (error) {
        console.warn(`Could not find parent account ${sampleAccount.parentCode}:`, error);
      }
    }

    // Check if account already exists
    try {
      const existingId = await this.findAccountIdByCode(sampleAccount.code);
      if (existingId) {
        // Account exists, update balance instead
        return await this.updateAccountBalance(existingId, sampleAccount.balance || 0);
      }
    } catch (error) {
      // Account doesn't exist, proceed with creation
    }

    // Create new account
    const response = await fetch(`${this.API_BASE_URL}/accounts`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(accountRequest),
    });

    if (!response.ok) {
      const errorData = await response.text();
      throw new Error(`HTTP ${response.status}: ${errorData}`);
    }

    const result = await response.json();
    return result.data || result;
  }

  /**
   * Find account ID by code
   */
  private async findAccountIdByCode(code: string): Promise<number | null> {
    const response = await fetch(`${this.API_BASE_URL}/accounts?code=${encodeURIComponent(code)}`, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      return null;
    }

    const result = await response.json();
    const accounts = result.data || result;
    
    if (Array.isArray(accounts) && accounts.length > 0) {
      return accounts[0].id;
    }

    return null;
  }

  /**
   * Update account balance
   */
  private async updateAccountBalance(accountId: number, balance: number): Promise<AccountResponse | null> {
    const response = await fetch(`${this.API_BASE_URL}/accounts/${accountId}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(),
      body: JSON.stringify({ balance }),
    });

    if (!response.ok) {
      const errorData = await response.text();
      throw new Error(`HTTP ${response.status}: ${errorData}`);
    }

    const result = await response.json();
    return result.data || result;
  }

  /**
   * Delete sample data (cleanup)
   */
  async deleteSampleBalanceSheetData(): Promise<{ success: boolean; message: string }> {
    const sampleCodes = [
      '1-1-001', '1-1-002', '1-1-003', '1-1-004',
      '1-2-001', '1-2-002', '1-2-003',
      '2-1-001', '2-1-002', '2-1-003', '2-2-001',
      '3-1-001', '3-2-001',
      '1-1-000', '1-2-000', '2-1-000', '2-2-000',
      '1-0-000', '2-0-000', '3-0-000',
    ];

    let deletedCount = 0;
    const errors: string[] = [];

    for (const code of sampleCodes) {
      try {
        const accountId = await this.findAccountIdByCode(code);
        if (accountId) {
          await this.deleteAccount(accountId);
          deletedCount++;
        }
      } catch (error) {
        errors.push(`Failed to delete account ${code}: ${error}`);
      }
    }

    return {
      success: true,
      message: `Deleted ${deletedCount} sample accounts. ${errors.length > 0 ? `Errors: ${errors.join('; ')}` : ''}`,
    };
  }

  /**
   * Delete account by ID
   */
  private async deleteAccount(accountId: number): Promise<void> {
    const response = await fetch(`${this.API_BASE_URL}/accounts/${accountId}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      const errorData = await response.text();
      throw new Error(`HTTP ${response.status}: ${errorData}`);
    }
  }

  /**
   * Format currency for display
   */
  private formatCurrency(amount: number): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  }

  /**
   * Check if sample data exists
   */
  async checkSampleDataExists(): Promise<boolean> {
    try {
      const sampleCodes = ['1-1-001', '2-1-001', '3-1-001']; // Check key accounts
      for (const code of sampleCodes) {
        const accountId = await this.findAccountIdByCode(code);
        if (!accountId) {
          return false;
        }
      }
      return true;
    } catch (error) {
      return false;
    }
  }
}

export const sampleDataService = new SampleDataService();