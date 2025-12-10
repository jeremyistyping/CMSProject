import api from './api';

export interface PurchaseOrderItem {
  id?: number;
  purchase_order_id?: number;
  pr_item_id?: number;
  material_id?: number;
  item_name: string;
  description?: string;
  quantity: number;
  unit: string;
  unit_price: number;
  total_price: number;
  received_quantity?: number;
  material?: {
    id: number;
    name: string;
    code: string;
  };
}

export interface PurchaseOrder {
  id: number;
  code: string;
  purchase_request_id: number;
  project_id: number;
  vendor_id?: number;
  order_date: string;
  expected_delivery_date?: string;
  delivery_address?: string;
  payment_terms?: string;
  notes?: string;
  subtotal: number;
  tax_amount: number;
  discount_amount: number;
  total_amount: number;
  status: string;
  created_by: number;
  approved_by?: number;
  approved_at?: string;
  created_at: string;
  updated_at: string;
  purchase_request?: {
    id: number;
    code: string;
  };
  project?: {
    id: number;
    project_name: string;
  };
  vendor?: {
    id: number;
    name: string;
  };
  creator?: {
    id: number;
    first_name: string;
    last_name: string;
  };
  items: PurchaseOrderItem[];
}

export interface GoodsReceiptItem {
  id?: number;
  goods_receipt_id?: number;
  po_item_id: number;
  received_quantity: number;
  accepted_quantity: number;
  rejected_quantity?: number;
  rejection_reason?: string;
}

export interface GoodsReceipt {
  id: number;
  code: string;
  purchase_order_id: number;
  project_id: number;
  receipt_date: string;
  received_by: number;
  notes?: string;
  status: string;
  created_at: string;
  items: GoodsReceiptItem[];
  receiver?: {
    id: number;
    first_name: string;
    last_name: string;
  };
}

export interface CreatePORequest {
  purchase_request_id: number;
  vendor_id?: number;
  expected_delivery_date?: string;
  delivery_address?: string;
  payment_terms?: string;
  notes?: string;
  items?: {
    pr_item_id: number;
    material_id?: number;
    item_name: string;
    quantity: number;
    unit: string;
    unit_price: number;
  }[];
}

export interface CreateGRRequest {
  purchase_order_id: number;
  receipt_date?: string;
  notes?: string;
  items: {
    po_item_id: number;
    received_quantity: number;
    accepted_quantity: number;
    rejected_quantity?: number;
    rejection_reason?: string;
  }[];
}

const purchaseOrderService = {
  // Create PO from approved PR
  createFromPR: async (data: CreatePORequest): Promise<PurchaseOrder> => {
    const response = await api.post('/api/v1/purchase-orders/from-pr', data);
    return response.data;
  },

  // Get all POs
  getAll: async (filter?: { project_id?: number; status?: string; vendor_id?: number }): Promise<PurchaseOrder[]> => {
    const params = new URLSearchParams();
    if (filter?.project_id) params.append('project_id', filter.project_id.toString());
    if (filter?.status) params.append('status', filter.status);
    if (filter?.vendor_id) params.append('vendor_id', filter.vendor_id.toString());
    
    const response = await api.get(`/api/v1/purchase-orders?${params.toString()}`);
    return response.data;
  },

  // Get PO by ID
  getById: async (id: number): Promise<PurchaseOrder> => {
    const response = await api.get(`/api/v1/purchase-orders/${id}`);
    return response.data;
  },

  // Get POs by PR ID
  getByPRId: async (prId: number): Promise<PurchaseOrder[]> => {
    const response = await api.get(`/api/v1/purchase-orders/by-pr/${prId}`);
    return response.data;
  },

  // Send PO to vendor
  sendPO: async (id: number): Promise<void> => {
    await api.post(`/api/v1/purchase-orders/${id}/send`);
  },

  // Delete PO
  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/v1/purchase-orders/${id}`);
  },

  // Create Goods Receipt
  createGoodsReceipt: async (data: CreateGRRequest): Promise<GoodsReceipt> => {
    const response = await api.post('/api/v1/purchase-orders/goods-receipt', data);
    return response.data;
  },

  // Get Goods Receipts by PO ID
  getGoodsReceiptsByPOId: async (poId: number): Promise<GoodsReceipt[]> => {
    const response = await api.get(`/api/v1/purchase-orders/${poId}/goods-receipts`);
    return response.data || [];
  },

  // Download PO as PDF
  downloadPOPDF: async (poId: number): Promise<void> => {
    const response = await api.get(`/api/v1/purchase-orders/${poId}/pdf`, {
      responseType: 'blob',
    });
    
    // Create blob link to download
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `PO_${poId}.pdf`);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  },

  // Download Goods Receipt as PDF
  downloadGRPDF: async (poId: number): Promise<void> => {
    const response = await api.get(`/api/v1/purchase-orders/${poId}/goods-receipt-pdf`, {
      responseType: 'blob',
    });
    
    // Create blob link to download
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `GR_${poId}.pdf`);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  },
};

export default purchaseOrderService;
