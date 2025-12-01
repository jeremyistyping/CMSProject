export interface PurchaseRequestItem {
    id?: number;
    purchase_request_id?: number;
    item_name: string;
    product_id?: number;
    quantity: number;
    unit: string;
    estimated_price: number;
    total_price: number;
    notes: string;
}

export interface PurchaseRequest {
    id: number;
    code: string;
    project_id: number;
    request_date: string;
    required_date: string;
    vendor_id?: number;
    notes: string;
    status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'REVISION' | 'PO_CREATED';
    total_amount: number;

    // Approval fields
    approved_by?: number;
    approved_at?: string;
    rejection_reason?: string;

    created_by: number;
    created_at: string;
    updated_at: string;

    // Relations
    project?: {
        id: number;
        project_name: string;
    };
    vendor?: {
        id: number;
        name: string;
    };
    requester?: {
        id: number;
        username: string;
        first_name: string;
        last_name: string;
        role: string;
    };
    approver?: {
        id: number;
        name: string;
    };
    items?: PurchaseRequestItem[];
}

export interface CreatePRData {
    project_id: number;
    request_date: string;
    required_date: string;
    vendor_id?: number;
    notes: string;
    items: PurchaseRequestItem[];
}

export interface UpdatePRData extends CreatePRData { }

export interface PRFilter {
    project_id?: number;
    status?: string;
}

export interface MaterialImpact {
    product_id?: number;
    product_name: string;
    product_code: string;
    unit: string;
    requested_qty: number;
    current_stock: number;
    projected_stock: number;
    status: 'OK' | 'LOW' | 'CRITICAL';
}
