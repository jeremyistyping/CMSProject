import api from './api';
import { CBSNode, CBSNodeSummary, PRCBSMapping } from '../types/cbs';

const cbsService = {
    // Get CBS Tree for a project
    getCBSTree: async (projectId: number): Promise<CBSNode[]> => {
        const response = await api.get(`/api/v1/projects/${projectId}/cbs`);
        return response.data;
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
        const response = await api.get(`/api/v1/cbs-nodes/${id}/summary`);
        return response.data;
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
        const response = await api.get(`/api/v1/purchase-requests/${prId}/cbs-mappings`);
        return response.data;
    }
};

export default cbsService;
