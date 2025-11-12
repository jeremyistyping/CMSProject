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
  id: string;
  project_id: string;
  work_area: string;
  start_date: string;
  end_date: string;
  duration_days: number;
  status: 'not-started' | 'in-progress' | 'completed' | 'delayed';
  dependencies: string[];
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

