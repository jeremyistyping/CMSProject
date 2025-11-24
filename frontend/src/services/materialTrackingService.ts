import api from './api';
import { API_V1_BASE } from '@/config/api';

export interface MaterialSummaryStats {
    total_purchased_value: number;
    total_used_value: number;
    total_remaining_value: number;
    total_items: number;
    low_stock_items: number;
}

export interface MaterialItemSummary {
    product_id: number;
    product_code: string;
    product_name: string;
    unit: string;
    category: string;
    budget_qty: number;
    purchased_qty: number;
    used_qty: number;
    remaining_qty: number;
    avg_unit_cost: number;
    total_value: number;
    status: 'OK' | 'LOW' | 'CRITICAL';
}

export interface MaterialMovement {
    id: number;
    product_id: number;
    reference_type: string;
    reference_id: number;
    type: 'IN' | 'OUT';
    quantity: number;
    unit_cost: number;
    total_cost: number;
    transaction_date: string;
    notes: string;
    product: {
        name: string;
        code: string;
        unit: string;
    };
}

class MaterialTrackingService {
    async getSummary(projectId: number): Promise<MaterialSummaryStats> {
        const response = await api.get(`${API_V1_BASE}/material-tracking/${projectId}/summary`);
        return response.data.data;
    }

    async getItems(projectId: number): Promise<MaterialItemSummary[]> {
        const response = await api.get(`${API_V1_BASE}/material-tracking/${projectId}/items`);
        return response.data.data;
    }

    async getMovements(projectId: number): Promise<MaterialMovement[]> {
        const response = await api.get(`${API_V1_BASE}/material-tracking/${projectId}/movements`);
        return response.data.data;
    }
}

export const materialTrackingService = new MaterialTrackingService();
