export interface CBSNode {
    id: number;
    project_id: number;
    parent_id?: number;
    code: string;
    name: string;
    description?: string;
    coa_account_id?: number;
    budget_amount: number;
    actual_cost: number;
    children?: CBSNode[];
    level?: number;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface PRCBSMapping {
    id: number;
    purchase_request_id: number;
    cbs_node_id: number;
    cbs_node?: CBSNode;
    pr_item_id?: number;
    allocated_amount: number;
    notes?: string;
    created_by?: number;
    created_at: string;
}

export interface CBSNodeSummary {
    node_id: number;
    budget_amount: number;
    actual_cost: number;
    variance: number;
    children_cost: number;
    total_cost: number;
}
