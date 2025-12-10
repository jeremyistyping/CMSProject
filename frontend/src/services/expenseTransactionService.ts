import api from './api';

// =====================================================
// Types
// =====================================================

export interface ExpenseTransaction {
    id: number;
    project_id: number;
    transaction_date: string;
    coa_account_id: number;
    description: string;
    amount: number;
    unit: string;
    quantity: number;
    transaction_type: string;
    reference_type?: string;
    reference_id?: number;
    reference_no?: string;
    notes?: string;
    created_by: number;
    created_at: string;
    updated_at: string;
    project?: any;
    coa_account?: any;
    creator?: any;
}

export interface CreateExpenseTransactionDTO {
    project_id: number;
    transaction_date: string;
    coa_account_id: number;
    description: string;
    amount: number;
    unit?: string;
    quantity?: number;
    transaction_type?: string;
    reference_type?: string;
    reference_id?: number;
    reference_no?: string;
    notes?: string;
}

export interface BudgetReportResponse {
    project_id: number;
    project_name: string;
    report_date: string;
    start_date: string;
    end_date: string;
    labour_budget?: BudgetCategoryReport;
    operasional_budget?: BudgetCategoryReport;
    other_budget?: BudgetCategoryReport;
}

export interface BudgetCategoryReport {
    budget_estimation: number;
    actual: number;
    variance: number;
    transactions: ExpenseTransactionDetail[];
    by_work_package?: WorkPackageSummary[];
}

export interface ExpenseTransactionDetail {
    date: string;
    description: string;
    unit: string;
    quantity: number;
    total_price: number;
    coa_code: string;
    coa_name: string;
    work_package?: string;
    reference_no?: string;
}

export interface WorkPackageSummary {
    work_package: string;
    budget_estimation: number;
    actual: number;
    variance: number;
    transactions: ExpenseTransactionDetail[];
}

// =====================================================
// Expense Transaction Service
// =====================================================

export const expenseTransactionService = {
    // Get all expense transactions with filters
    getAll: async (filter?: Record<string, any>): Promise<{ data: ExpenseTransaction[]; total: number }> => {
        const params = new URLSearchParams();
        if (filter) {
            Object.entries(filter).forEach(([key, value]) => {
                if (value !== undefined && value !== '') params.append(key, String(value));
            });
        }
        const response = await api.get(`/api/v1/expense-transactions?${params.toString()}`);
        return response.data;
    },

    // Get expense transaction by ID
    getById: async (id: number): Promise<ExpenseTransaction> => {
        const response = await api.get(`/api/v1/expense-transactions/${id}`);
        return response.data;
    },

    // Get expense transactions by project
    getByProject: async (projectId: number, filter?: Record<string, any>): Promise<{ data: ExpenseTransaction[]; total: number }> => {
        const params = new URLSearchParams();
        if (filter) {
            Object.entries(filter).forEach(([key, value]) => {
                if (value !== undefined && value !== '') params.append(key, String(value));
            });
        }
        const response = await api.get(`/api/v1/projects/${projectId}/expenses?${params.toString()}`);
        return response.data;
    },

    // Get budget vs actual report
    getBudgetReport: async (projectId: number, startDate: string, endDate: string): Promise<BudgetReportResponse> => {
        const response = await api.get(`/api/v1/projects/${projectId}/reports/budget-vs-actual`, {
            params: { start_date: startDate, end_date: endDate }
        });
        return response.data;
    },

    // Create expense transaction
    create: async (projectId: number, data: CreateExpenseTransactionDTO): Promise<ExpenseTransaction> => {
        const response = await api.post(`/api/v1/projects/${projectId}/expenses`, data);
        return response.data;
    },

    // Batch create expense transactions
    batchCreate: async (projectId: number, data: CreateExpenseTransactionDTO[]): Promise<{ data: ExpenseTransaction[]; total: number }> => {
        const response = await api.post(`/api/v1/projects/${projectId}/expenses/batch`, data);
        return response.data;
    },

    // Update expense transaction
    update: async (id: number, data: Partial<CreateExpenseTransactionDTO>): Promise<ExpenseTransaction> => {
        const response = await api.put(`/api/v1/expense-transactions/${id}`, data);
        return response.data;
    },

    // Delete expense transaction
    delete: async (id: number): Promise<void> => {
        await api.delete(`/api/v1/expense-transactions/${id}`);
    },

    // Export budget report as PDF
    exportBudgetReportPDF: async (projectId: number, startDate: string, endDate: string): Promise<string> => {
        const params = new URLSearchParams({
            start_date: startDate,
            end_date: endDate
        });
        
        // Return the URL for opening in new tab
        const baseURL = api.defaults.baseURL || 'http://localhost:8080';
        const token = localStorage.getItem('token');
        return `${baseURL}/api/v1/projects/${projectId}/reports/budget-vs-actual/pdf?${params.toString()}&token=${token}`;
    },
};

export default expenseTransactionService;
