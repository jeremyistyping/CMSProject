export interface Milestone {
    id: number;
    title: string;
    description?: string;
    work_area?: string;
    priority: string;
    assigned_team?: string;
    target_date: string;
    status: string;
    completion_date?: string;
    created_at?: string;
    updated_at?: string;
}
