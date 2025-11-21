import api from './api';
import { PurchaseRequest, CreatePRData, UpdatePRData, PRFilter } from '../types/purchaseRequest';

const purchaseRequestService = {
    getAll: async (filter?: PRFilter): Promise<PurchaseRequest[]> => {
        const params = new URLSearchParams();
        if (filter?.project_id) params.append('project_id', filter.project_id.toString());
        if (filter?.status) params.append('status', filter.status);

        const response = await api.get(`/api/v1/purchase-requests?${params.toString()}`);
        return response.data;
    },

    getById: async (id: number): Promise<PurchaseRequest> => {
        const response = await api.get(`/api/v1/purchase-requests/${id}`);
        return response.data;
    },

    create: async (data: CreatePRData): Promise<PurchaseRequest> => {
        const response = await api.post('/api/v1/purchase-requests', data);
        return response.data;
    },

    update: async (id: number, data: UpdatePRData): Promise<void> => {
        await api.put(`/api/v1/purchase-requests/${id}`, data);
    },

    delete: async (id: number): Promise<void> => {
        await api.delete(`/api/v1/purchase-requests/${id}`);
    },

    updateStatus: async (id: number, status: string, reason?: string): Promise<void> => {
        await api.patch(`/api/v1/purchase-requests/${id}/status`, { status, reason });
    },
};

export default purchaseRequestService;
