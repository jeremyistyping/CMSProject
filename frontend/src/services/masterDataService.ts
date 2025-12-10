import api from './api';

// =====================================================
// Types
// =====================================================

export interface COAAccount {
    id: number;
    code: string;
    name: string;
    description?: string;
    type: 'ASSET' | 'LIABILITY' | 'EQUITY' | 'REVENUE' | 'EXPENSE';
    category?: string;
    budget_category?: string;
    work_package?: string;
    parent_id?: number;
    level: number;
    is_active: boolean;
    is_header: boolean;
    balance: number;
    children?: COAAccount[];
}

export interface COATreeNode {
    id: number;
    code: string;
    name: string;
    type: string;
    category?: string;
    budget_category?: string;
    work_package?: string;
    level: number;
    is_active: boolean;
    is_header: boolean;
    balance: number;
    children?: COATreeNode[];
}

export interface Material {
    id: number;
    code: string;
    name: string;
    description?: string;
    category_id?: number;
    category?: MaterialCategory;
    unit: string;
    unit_price: number;
    min_stock: number;
    max_stock: number;
    current_stock: number;
    coa_account_id?: number;
    coa_account?: COAAccount;
    is_active: boolean;
}


export interface MaterialCategory {
    id: number;
    code: string;
    name: string;
    description?: string;
    parent_id?: number;
    is_active: boolean;
    children?: MaterialCategory[];
}

export interface UnitOfMeasure {
    id: number;
    code: string;
    name: string;
    symbol?: string;
    category: string;
    is_active: boolean;
}

export interface Vendor {
    id: number;
    code: string;
    name: string;
    contact_person?: string;
    email?: string;
    phone?: string;
    address?: string;
    city?: string;
    province?: string;
    postal_code?: string;
    npwp?: string;
    bank_name?: string;
    bank_account?: string;
    bank_branch?: string;
    payment_terms: number;
    category_id?: number;
    category?: VendorCategory;
    rating: number;
    notes?: string;
    is_active: boolean;
}

export interface VendorCategory {
    id: number;
    code: string;
    name: string;
    description?: string;
    is_active: boolean;
}

export interface VendorMaterial {
    id: number;
    vendor_id: number;
    material_id: number;
    vendor?: Vendor;
    material?: Material;
    unit_price: number;
    lead_time_days: number;
    min_order_qty: number;
    is_preferred: boolean;
    notes?: string;
}

export interface MaterialSummary {
    total_materials: number;
    active_materials: number;
    low_stock_count: number;
    total_stock_value: number;
    total_categories: number;
}

export interface VendorSummary {
    total_vendors: number;
    active_vendors: number;
    total_categories: number;
    average_rating: number;
}

// =====================================================
// COA Service
// =====================================================

export const coaService = {
    getAll: async (filter?: Record<string, any>): Promise<{ data: COAAccount[]; total: number }> => {
        const params = new URLSearchParams();
        if (filter) {
            Object.entries(filter).forEach(([key, value]) => {
                if (value !== undefined && value !== '') params.append(key, String(value));
            });
        }
        const response = await api.get(`/api/v1/master-data/coa?${params.toString()}`);
        return response.data;
    },

    getById: async (id: number): Promise<COAAccount> => {
        const response = await api.get(`/api/v1/master-data/coa/${id}`);
        return response.data;
    },

    getTree: async (): Promise<{ data: COATreeNode[] }> => {
        const response = await api.get('/api/v1/master-data/coa/tree');
        return response.data;
    },

    getByType: async (type: string): Promise<{ data: COAAccount[] }> => {
        const response = await api.get(`/api/v1/master-data/coa/type/${type}`);
        return response.data;
    },

    getByCategory: async (category: string): Promise<{ data: COAAccount[] }> => {
        const response = await api.get(`/api/v1/master-data/coa/category/${category}`);
        return response.data;
    },

    getByBudgetCategory: async (budgetCategory: string): Promise<{ data: COAAccount[] }> => {
        const response = await api.get(`/api/v1/master-data/coa/budget-category/${budgetCategory}`);
        return response.data;
    },

    create: async (data: Partial<COAAccount>): Promise<COAAccount> => {
        const response = await api.post('/api/v1/master-data/coa', data);
        return response.data;
    },

    update: async (id: number, data: Partial<COAAccount>): Promise<COAAccount> => {
        const response = await api.put(`/api/v1/master-data/coa/${id}`, data);
        return response.data;
    },

    delete: async (id: number): Promise<void> => {
        await api.delete(`/api/v1/master-data/coa/${id}`);
    },
};


// =====================================================
// Material Service
// =====================================================

export const materialService = {
    getAll: async (filter?: Record<string, any>): Promise<{ data: Material[]; total: number }> => {
        const params = new URLSearchParams();
        if (filter) {
            Object.entries(filter).forEach(([key, value]) => {
                if (value !== undefined && value !== '') params.append(key, String(value));
            });
        }
        const response = await api.get(`/api/v1/master-data/materials?${params.toString()}`);
        return response.data;
    },

    getById: async (id: number): Promise<Material> => {
        const response = await api.get(`/api/v1/master-data/materials/${id}`);
        return response.data;
    },

    getSummary: async (): Promise<MaterialSummary> => {
        const response = await api.get('/api/v1/master-data/materials/summary');
        return response.data;
    },

    getVendors: async (materialId: number): Promise<{ data: VendorMaterial[] }> => {
        const response = await api.get(`/api/v1/master-data/materials/${materialId}/vendors`);
        return response.data;
    },

    create: async (data: Partial<Material>): Promise<Material> => {
        const response = await api.post('/api/v1/master-data/materials', data);
        return response.data;
    },

    update: async (id: number, data: Partial<Material>): Promise<Material> => {
        const response = await api.put(`/api/v1/master-data/materials/${id}`, data);
        return response.data;
    },

    delete: async (id: number): Promise<void> => {
        await api.delete(`/api/v1/master-data/materials/${id}`);
    },

    // Categories
    getCategories: async (): Promise<{ data: MaterialCategory[] }> => {
        const response = await api.get('/api/v1/master-data/material-categories');
        return response.data;
    },

    getCategoryTree: async (): Promise<{ data: MaterialCategory[] }> => {
        const response = await api.get('/api/v1/master-data/material-categories/tree');
        return response.data;
    },

    createCategory: async (data: Partial<MaterialCategory>): Promise<MaterialCategory> => {
        const response = await api.post('/api/v1/master-data/material-categories', data);
        return response.data;
    },

    updateCategory: async (id: number, data: Partial<MaterialCategory>): Promise<MaterialCategory> => {
        const response = await api.put(`/api/v1/master-data/material-categories/${id}`, data);
        return response.data;
    },

    deleteCategory: async (id: number): Promise<void> => {
        await api.delete(`/api/v1/master-data/material-categories/${id}`);
    },

    // UoM
    getUoM: async (): Promise<{ data: UnitOfMeasure[] }> => {
        const response = await api.get('/api/v1/master-data/uom');
        return response.data;
    },
};

// =====================================================
// Vendor Service
// =====================================================

export const vendorService = {
    getAll: async (filter?: Record<string, any>): Promise<{ data: Vendor[]; total: number }> => {
        const params = new URLSearchParams();
        if (filter) {
            Object.entries(filter).forEach(([key, value]) => {
                if (value !== undefined && value !== '') params.append(key, String(value));
            });
        }
        const response = await api.get(`/api/v1/master-data/vendors?${params.toString()}`);
        return response.data;
    },

    getById: async (id: number): Promise<Vendor> => {
        const response = await api.get(`/api/v1/master-data/vendors/${id}`);
        return response.data;
    },

    getSummary: async (): Promise<VendorSummary> => {
        const response = await api.get('/api/v1/master-data/vendors/summary');
        return response.data;
    },

    getMaterials: async (vendorId: number): Promise<{ data: VendorMaterial[] }> => {
        const response = await api.get(`/api/v1/master-data/vendors/${vendorId}/materials`);
        return response.data;
    },

    create: async (data: Partial<Vendor>): Promise<Vendor> => {
        const response = await api.post('/api/v1/master-data/vendors', data);
        return response.data;
    },

    update: async (id: number, data: Partial<Vendor>): Promise<Vendor> => {
        const response = await api.put(`/api/v1/master-data/vendors/${id}`, data);
        return response.data;
    },

    delete: async (id: number): Promise<void> => {
        await api.delete(`/api/v1/master-data/vendors/${id}`);
    },

    addMaterial: async (vendorId: number, data: { material_id: number; unit_price?: number; lead_time_days?: number; is_preferred?: boolean }): Promise<void> => {
        await api.post(`/api/v1/master-data/vendors/${vendorId}/materials`, data);
    },

    removeMaterial: async (vendorId: number, materialId: number): Promise<void> => {
        await api.delete(`/api/v1/master-data/vendors/${vendorId}/materials/${materialId}`);
    },

    // Categories
    getCategories: async (): Promise<{ data: VendorCategory[] }> => {
        const response = await api.get('/api/v1/master-data/vendor-categories');
        return response.data;
    },

    createCategory: async (data: Partial<VendorCategory>): Promise<VendorCategory> => {
        const response = await api.post('/api/v1/master-data/vendor-categories', data);
        return response.data;
    },

    updateCategory: async (id: number, data: Partial<VendorCategory>): Promise<VendorCategory> => {
        const response = await api.put(`/api/v1/master-data/vendor-categories/${id}`, data);
        return response.data;
    },

    deleteCategory: async (id: number): Promise<void> => {
        await api.delete(`/api/v1/master-data/vendor-categories/${id}`);
    },
};

// Default export for backward compatibility
const masterDataService = {
    getVendors: async () => {
        const response = await vendorService.getAll({ is_active: true });
        return response.data || [];
    },
    getCOA: async () => {
        const response = await coaService.getAll();
        return response.data || [];
    },
    getMaterials: async () => {
        const response = await materialService.getAll();
        return response.data || [];
    },
};

export default masterDataService;
