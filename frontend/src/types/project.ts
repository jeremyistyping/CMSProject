export interface Project {
  id: string | number;
  project_name: string;
  project_description: string;
  customer: string;
  city: string;
  address: string;
  project_type: 'New Build' | 'Renovation' | 'Expansion' | 'Maintenance';
  budget: number;
  deadline: string;
  overall_progress: number;
  foundation_progress: number;
  utilities_progress: number;
  interior_progress: number;
  equipment_progress: number;
  status: 'active' | 'archived' | 'completed' | 'on-hold';
  created_at: string;
  updated_at: string;

  // Optional fields from backend cost tracking
  actual_cost?: number;
  material_cost?: number;
  labor_cost?: number;
  equipment_cost?: number;
  overhead_cost?: number;
  variance?: number;
  variance_percent?: number;
}

export interface ProjectFormData {
  project_name: string;
  project_description: string;
  customer: string;
  city: string;
  address: string;
  project_type: string;
  budget: number;
  deadline: string;
  overall_progress: number;
  foundation_progress: number;
  utilities_progress: number;
  interior_progress: number;
  equipment_progress: number;
}

export interface ProjectProgress {
  overall: number;
  foundation: number;
  utilities: number;
  interior: number;
  equipment: number;
}

// Time-series snapshot of physical progress per project (project_progress table)
export interface ProjectProgressEntry {
  id: number;
  project_id: number;
  date: string; // ISO date string
  physical_progress_percent: number;
  volume_achieved?: number | null;
  remarks?: string;
}

// Derived actual cost rows per project (project_actual_costs view)
export interface ProjectActualCost {
  id: number;
  project_id: number;
  project_budget_id?: number | null;
  source_type: string;
  source_id: number;
  date: string;
  amount: number;
  category: string;
  status: string;
}

export interface Milestone {
  id: string;
  project_id: string;
  title: string;
  description: string;
  target_date: string;
  completion_date?: string;
  status: 'pending' | 'in-progress' | 'completed' | 'delayed';
  created_at: string;
}

export interface DailyUpdate {
  id: string;
  project_id: string;
  date: string;
  weather: string;
  workers_present: number;
  work_description: string;
  materials_used: string;
  issues: string;
  tomorrows_plan: string;
  photos: string[];
  created_by: string;
  created_at: string;
}

export interface WeeklyReport {
  id: string;
  project_id: string;
  week_number: number;
  start_date: string;
  end_date: string;
  progress_summary: string;
  accomplishments: string;
  challenges: string;
  next_week_plan: string;
  budget_status: string;
  created_at: string;
}

export interface TimelineSchedule {
  id: string | number;
  project_id: string | number;
  work_area: string;
  assigned_team: string;
  start_date: string;
  end_date: string;
  start_time: string;
  end_time: string;
  notes: string;
  status: 'not-started' | 'in-progress' | 'completed';
  duration?: number;
  days_remaining?: number;
  is_active?: boolean;
  is_overdue?: boolean;
  status_color?: string;
  created_at?: string;
  updated_at?: string;
}

export interface TechnicalData {
  id: string;
  project_id: string;
  category: string;
  title: string;
  description: string;
  specifications: string;
  documents: string[];
  created_at: string;
}

// ProjectBudget represents a single COA budget line for a project
export interface ProjectBudget {
  id: number;
  project_id: number;
  account_id: number;
  estimated_amount: number;
  created_at?: string;
  updated_at?: string;
}

export interface ProjectBudgetInput {
  account_id: number;
  estimated_amount: number;
}

// High-level cost summary for a project (Budget vs Actual + Progress)
export interface ProjectCostSummary {
  project_id: number;
  project_name: string;
  budget: number;
  actual_cost: number;
  material_cost: number;
  labor_cost: number;
  equipment_cost: number;
  overhead_cost: number;
  variance: number;
  variance_percent: number;
  budget_utilization: number;
  remaining_budget: number;
  is_over_budget: boolean;
  total_purchases: number;
  overall_progress: number;
  status: string;
}

// Budget vs Actual report (per project, grouped by COA)
export interface BudgetVsActualCOAGroup {
  coa_code: string;
  coa_name: string;
  coa_type: string;
  budget: number;
  actual: number;
  variance: number;
  variance_rate: number;
  status: string; // OVER_BUDGET, UNDER_BUDGET, ON_TARGET
}

export interface BudgetVsActualReport {
  report_date: string;
  project_id?: number | null;
  project_name?: string;
  start_date: string;
  end_date: string;
  coa_groups: BudgetVsActualCOAGroup[];
  total_budget: number;
  total_actual: number;
  total_variance: number;
  variance_rate: number;
}

// Portfolio Budget vs Actual (all projects)
export interface ProjectBudgetVsActualSummary {
  project_id: number;
  project_name: string;
  budget: number;
  actual: number;
  variance: number;
  variance_percent: number;
  cost_progress: number;
  physical_progress: number;
  progress_gap: number;
  status: string; // OVER_BUDGET, UNDER_UTILIZED, ON_TRACK, NO_BUDGET
}

export interface PortfolioBudgetVsActualReport {
  report_date: string;
  start_date: string;
  end_date: string;
  projects: ProjectBudgetVsActualSummary[];
}

// Progress vs Cost correlation per project (time-series)
export interface ProgressVsCostPoint {
  date: string;
  physical_progress: number;
  cumulative_actual: number;
  budget: number;
  cost_progress: number;
  progress_gap: number;
}

export interface ProgressVsCostReport {
  project_id: number;
  project_name: string;
  start_date: string;
  end_date: string;
  budget: number;
  points: ProgressVsCostPoint[];
}

