import api from './api';
import { CBSNode, CBSNodeSummary, PRCBSMapping } from '../types/cbs';

const cbsService = {
    // Get CBS Tree for a project
    getCBSTree: async (projectId: number): Promise<CBSNode[]> => {
        try {
            const response = await api.get(`/api/v1/projects/${projectId}/cbs`);
            return response.data;
        } catch (error: any) {
            if (error.response?.status === 404) {
                console.warn('CBS tree endpoint not available');
                return [];
            }
            throw error;
        }
    },

    // Create a new CBS Node
    createCBSNode: async (node: Partial<CBSNode>): Promise<CBSNode> => {
        const response = await api.post('/api/v1/cbs-nodes', node);
        return response.data;
    },

    // Update an existing CBS Node
    updateCBSNode: async (id: number, node: Partial<CBSNode>): Promise<void> => {
        await api.put(`/api/v1/cbs-nodes/${id}`, node);
    },

    // Delete a CBS Node
    deleteCBSNode: async (id: number): Promise<void> => {
        await api.delete(`/api/v1/cbs-nodes/${id}`);
    },

    // Get Cost Summary for a Node
    getNodeSummary: async (id: number): Promise<CBSNodeSummary> => {
        try {
            const response = await api.get(`/api/v1/cbs-nodes/${id}/summary`);
            return response.data;
        } catch (error: any) {
            if (error.response?.status === 404) {
                console.warn('CBS node summary endpoint not available');
                return {
                    node_id: id,
                    budget_amount: 0,
                    actual_cost: 0,
                    variance: 0,
                    children_cost: 0,
                    total_cost: 0,
                };
            }
            throw error;
        }
    },

    // Verify PR and map to CBS
    verifyPR: async (prId: number, mappings: Partial<PRCBSMapping>[], notes: string): Promise<void> => {
        await api.post(`/api/v1/purchase-requests/${prId}/verify`, {
            mappings,
            notes
        });
    },

    // Get CBS Mappings for a PR
    getPRCBSMappings: async (prId: number): Promise<PRCBSMapping[]> => {
        try {
            const response = await api.get(`/api/v1/purchase-requests/${prId}/cbs-mappings`);
            return response.data;
        } catch (error: any) {
            if (error.response?.status === 404) {
                console.warn('PR CBS mappings endpoint not available');
                return [];
            }
            throw error;
        }
    },

    // Get Project Budget Summary from CBS
    getProjectBudgetSummary: async (projectId: number): Promise<{
        project_id: number;
        total_budget: number;
        total_actual: number;
        total_variance: number;
        node_count: number;
    }> => {
        const response = await api.get(`/api/v1/projects/${projectId}/cbs/summary`);
        return response.data;
    }
};

export default cbsService;
