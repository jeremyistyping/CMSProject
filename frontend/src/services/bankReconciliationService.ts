import api from './api';

// ========== TYPES ==========

export interface BankReconciliationSnapshot {
  id: number;
  cash_bank_id: number;
  period: string;
  snapshot_date: string;
  generated_by: number;
  opening_balance: number;
  closing_balance: number;
  total_debit: number;
  total_credit: number;
  transaction_count: number;
  data_hash: string;
  is_locked: boolean;
  locked_at?: string;
  locked_by?: number;
  notes?: string;
  status: 'ACTIVE' | 'ARCHIVED' | 'SUPERSEDED';
  created_at: string;
  updated_at: string;
  cash_bank?: {
    id: number;
    name: string;
    code: string;
    type: string;
  };
  generated_by_user?: {
    id: number;
    username: string;
  };
  transactions?: ReconciliationTransactionSnapshot[];
}

export interface ReconciliationTransactionSnapshot {
  id: number;
  snapshot_id: number;
  transaction_id: number;
  transaction_date: string;
  reference_type?: string;
  reference_id?: number;
  reference_number?: string;
  amount: number;
  debit_amount: number;
  credit_amount: number;
  balance_after: number;
  description?: string;
  notes?: string;
  created_by: number;
  created_at: string;
}

export interface BankReconciliation {
  id: number;
  reconciliation_number: string;
  cash_bank_id: number;
  period: string;
  base_snapshot_id: number;
  comparison_snapshot_id?: number;
  reconciliation_date: string;
  reconciliation_by: number;
  base_balance: number;
  current_balance: number;
  variance: number;
  base_transaction_count: number;
  current_transaction_count: number;
  transaction_variance: number;
  missing_transactions: number;
  added_transactions: number;
  modified_transactions: number;
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'NEEDS_REVIEW';
  reviewed_by?: number;
  reviewed_at?: string;
  review_notes?: string;
  is_balanced: boolean;
  balance_confirmed: boolean;
  notes?: string;
  created_at: string;
  updated_at: string;
  cash_bank?: {
    id: number;
    name: string;
    code: string;
  };
  base_snapshot?: BankReconciliationSnapshot;
  comparison_snapshot?: BankReconciliationSnapshot;
  reconciliation_by_user?: {
    id: number;
    username: string;
  };
  reviewed_by_user?: {
    id: number;
    username: string;
  };
  differences?: ReconciliationDifference[];
}

export interface ReconciliationDifference {
  id: number;
  reconciliation_id: number;
  difference_type: 'MISSING' | 'ADDED' | 'MODIFIED' | 'AMOUNT_CHANGE' | 'DATE_CHANGE';
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  base_transaction_id?: number;
  current_transaction_id?: number;
  field?: string;
  old_value?: string;
  new_value?: string;
  amount_difference?: number;
  status: 'PENDING' | 'RESOLVED' | 'IGNORED' | 'ESCALATED';
  resolution_notes?: string;
  resolved_by?: number;
  resolved_at?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CashBankAuditTrail {
  id: number;
  cash_bank_id: number;
  transaction_id?: number;
  action: 'CREATE' | 'UPDATE' | 'DELETE' | 'VOID' | 'RESTORE';
  entity_type: 'CASH_BANK' | 'TRANSACTION' | 'TRANSFER' | 'DEPOSIT' | 'WITHDRAWAL';
  entity_id: number;
  field_changed?: string;
  old_value?: string;
  new_value?: string;
  reason?: string;
  ip_address?: string;
  user_agent?: string;
  requires_approval: boolean;
  approved_by?: number;
  approved_at?: string;
  approval_status?: 'PENDING' | 'APPROVED' | 'REJECTED';
  user_id: number;
  created_at: string;
  user?: {
    id: number;
    username: string;
  };
  approved_by_user?: {
    id: number;
    username: string;
  };
}

// ========== REQUEST TYPES ==========

export interface CreateSnapshotRequest {
  cash_bank_id: number;
  period: string; // Format: YYYY-MM
  notes?: string;
}

export interface CreateReconciliationRequest {
  cash_bank_id: number;
  period: string;
  base_snapshot_id: number;
  comparison_snapshot_id?: number;
  notes?: string;
}

export interface ApproveRejectRequest {
  notes?: string;
}

// ========== SERVICE CLASS ==========

class BankReconciliationService {
  private baseUrl = '/api/bank-reconciliation';

  // ========== SNAPSHOT OPERATIONS ==========

  async generateSnapshot(request: CreateSnapshotRequest): Promise<BankReconciliationSnapshot> {
    try {
      const response = await api.post(`${this.baseUrl}/snapshots`, request);
      return response.data.data;
    } catch (error) {
      console.error('Error generating snapshot:', error);
      throw error;
    }
  }

  async getSnapshots(cashBankId: number): Promise<BankReconciliationSnapshot[]> {
    try {
      const response = await api.get(`${this.baseUrl}/snapshots`, {
        params: { cash_bank_id: cashBankId, _t: Date.now() },
      });
      return response.data.data;
    } catch (error) {
      console.error('Error fetching snapshots:', error);
      throw error;
    }
  }

  async getSnapshotById(id: number): Promise<BankReconciliationSnapshot> {
    try {
      const response = await api.get(`${this.baseUrl}/snapshots/${id}`, {
        params: { _t: Date.now() },
      });
      return response.data.data;
    } catch (error) {
      console.error('Error fetching snapshot:', error);
      throw error;
    }
  }

  async lockSnapshot(id: number): Promise<void> {
    try {
      await api.post(`${this.baseUrl}/snapshots/${id}/lock`);
    } catch (error) {
      console.error('Error locking snapshot:', error);
      throw error;
    }
  }

  // ========== RECONCILIATION OPERATIONS ==========

  async performReconciliation(request: CreateReconciliationRequest): Promise<BankReconciliation> {
    try {
      const response = await api.post(`${this.baseUrl}/reconcile`, request);
      return response.data.data;
    } catch (error) {
      console.error('Error performing reconciliation:', error);
      throw error;
    }
  }

  async getReconciliations(cashBankId: number): Promise<BankReconciliation[]> {
    try {
      const response = await api.get(`${this.baseUrl}/reconciliations`, {
        params: { cash_bank_id: cashBankId, _t: Date.now() },
      });
      return response.data.data;
    } catch (error) {
      console.error('Error fetching reconciliations:', error);
      throw error;
    }
  }

  async getReconciliationById(id: number): Promise<BankReconciliation> {
    try {
      const response = await api.get(`${this.baseUrl}/reconciliations/${id}`, {
        params: { _t: Date.now() },
      });
      return response.data.data;
    } catch (error) {
      console.error('Error fetching reconciliation:', error);
      throw error;
    }
  }

  async approveReconciliation(id: number, notes?: string): Promise<void> {
    try {
      await api.post(`${this.baseUrl}/reconciliations/${id}/approve`, { notes });
    } catch (error) {
      console.error('Error approving reconciliation:', error);
      throw error;
    }
  }

  async rejectReconciliation(id: number, notes: string): Promise<void> {
    try {
      await api.post(`${this.baseUrl}/reconciliations/${id}/reject`, { notes });
    } catch (error) {
      console.error('Error rejecting reconciliation:', error);
      throw error;
    }
  }

  // ========== AUDIT TRAIL OPERATIONS ==========

  async getAuditTrail(cashBankId: number, limit: number = 100): Promise<CashBankAuditTrail[]> {
    try {
      const response = await api.get(`${this.baseUrl}/audit-trail`, {
        params: { cash_bank_id: cashBankId, limit, _t: Date.now() },
      });
      return response.data.data;
    } catch (error) {
      console.error('Error fetching audit trail:', error);
      throw error;
    }
  }

  async logAuditTrail(log: Partial<CashBankAuditTrail>): Promise<void> {
    try {
      await api.post(`${this.baseUrl}/audit-trail`, log);
    } catch (error) {
      console.error('Error logging audit trail:', error);
      throw error;
    }
  }

  // ========== HELPER FUNCTIONS ==========

  getSeverityColor(severity: string): string {
    switch (severity) {
      case 'CRITICAL':
        return 'red';
      case 'HIGH':
        return 'orange';
      case 'MEDIUM':
        return 'yellow';
      case 'LOW':
        return 'green';
      default:
        return 'gray';
    }
  }

  getStatusColor(status: string): string {
    switch (status) {
      case 'APPROVED':
        return 'green';
      case 'REJECTED':
        return 'red';
      case 'PENDING':
        return 'yellow';
      case 'NEEDS_REVIEW':
        return 'orange';
      default:
        return 'gray';
    }
  }

  getDifferenceTypeLabel(type: string): string {
    switch (type) {
      case 'MISSING':
        return 'Transaction Missing';
      case 'ADDED':
        return 'Transaction Added';
      case 'MODIFIED':
        return 'Transaction Modified';
      case 'AMOUNT_CHANGE':
        return 'Amount Changed';
      case 'DATE_CHANGE':
        return 'Date Changed';
      default:
        return type;
    }
  }

  formatCurrency(amount: number, currency: string = 'IDR'): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    }).format(amount);
  }

  formatDate(date: string): string {
    return new Date(date).toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  formatPeriod(period: string): string {
    const [year, month] = period.split('-');
    const monthNames = [
      'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
      'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
    ];
    return `${monthNames[parseInt(month) - 1]} ${year}`;
  }
}

export default new BankReconciliationService();
