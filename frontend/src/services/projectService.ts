import api from './api';
import {
  Project,
  ProjectFormData,
  Milestone,
  DailyUpdate,
  WeeklyReport,
  TimelineSchedule,
  TechnicalData,
  ProjectBudget,
  ProjectBudgetInput,
  ProjectCostSummary,
  ProjectProgressEntry,
  ProjectActualCost,
} from '@/types/project';

const PROJECT_ENDPOINT = '/api/v1/projects';

export const projectService = {
  // Projects CRUD
  async getAllProjects(): Promise<Project[]> {
    const response = await api.get(PROJECT_ENDPOINT);
    return response.data;
  },

  async getActiveProjects(): Promise<Project[]> {
    const response = await api.get(PROJECT_ENDPOINT, {
      params: { status: 'active' }
    });
    return response.data;
  },

  async getProjectById(id: string): Promise<Project> {
    const response = await api.get(`${PROJECT_ENDPOINT}/${id}`);
    return response.data;
  },

  async createProject(data: ProjectFormData): Promise<Project> {
    // Convert deadline to ISO 8601 format with timezone
    const formattedData = {
      ...data,
      deadline: data.deadline ? new Date(data.deadline).toISOString() : new Date().toISOString(),
    };
    const response = await api.post(PROJECT_ENDPOINT, formattedData);
    return response.data;
  },

  async updateProject(id: string, data: Partial<ProjectFormData>): Promise<Project> {
    const response = await api.put(`${PROJECT_ENDPOINT}/${id}`, data);
    return response.data;
  },

  async deleteProject(id: string): Promise<void> {
    await api.delete(`${PROJECT_ENDPOINT}/${id}`);
  },

  async archiveProject(id: string): Promise<Project> {
    const response = await api.post(`${PROJECT_ENDPOINT}/${id}/archive`);
    return response.data;
  },

  async updateProgress(id: string, progressData: Partial<Project>): Promise<Project> {
    const response = await api.patch(`${PROJECT_ENDPOINT}/${id}/progress`, progressData);
    return response.data;
  },

  // Milestones
  async getMilestones(projectId: string): Promise<Milestone[]> {
    const response = await api.get(`${PROJECT_ENDPOINT}/${projectId}/milestones`);
    return response.data;
  },

  async createMilestone(projectId: string, data: Partial<Milestone>): Promise<Milestone> {
    const response = await api.post(`${PROJECT_ENDPOINT}/${projectId}/milestones`, data);
    return response.data;
  },

  async updateMilestone(projectId: string, milestoneId: string, data: Partial<Milestone>): Promise<Milestone> {
    const response = await api.put(`${PROJECT_ENDPOINT}/${projectId}/milestones/${milestoneId}`, data);
    return response.data;
  },

  async deleteMilestone(projectId: string, milestoneId: string): Promise<void> {
    await api.delete(`${PROJECT_ENDPOINT}/${projectId}/milestones/${milestoneId}`);
  },

  // Daily Updates
  async getDailyUpdates(projectId: string, startDate?: string, endDate?: string): Promise<DailyUpdate[]> {
    const params: any = {};
    if (startDate) params.start_date = startDate;
    if (endDate) params.end_date = endDate;
    
    const response = await api.get(`${PROJECT_ENDPOINT}/${projectId}/daily-updates`, { params });
    return response.data;
  },

  async createDailyUpdate(projectId: string, data: Partial<DailyUpdate>, photos?: File[]): Promise<DailyUpdate> {
    // If photos are provided, use FormData for multipart upload
    if (photos && photos.length > 0) {
      const formData = new FormData();
      
      // Append all form fields as JSON string or individual fields
      formData.append('date', data.date || new Date().toISOString());
      formData.append('weather', data.weather || 'Sunny');
      formData.append('workers_present', String(data.workers_present || 0));
      formData.append('work_description', data.work_description || '');
      formData.append('materials_used', data.materials_used || '');
      formData.append('issues', data.issues || '');
      formData.append('tomorrows_plan', data.tomorrows_plan || '');
      formData.append('created_by', data.created_by || 'Current User');
      
      // Append photo files
      photos.forEach((photo) => {
        formData.append('photos', photo);
      });
      
      const response = await api.post(`${PROJECT_ENDPOINT}/${projectId}/daily-updates`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      return response.data;
    }
    
    // No photos, use regular JSON
    const response = await api.post(`${PROJECT_ENDPOINT}/${projectId}/daily-updates`, data);
    return response.data;
  },

  async updateDailyUpdate(projectId: string, updateId: string, data: Partial<DailyUpdate>): Promise<DailyUpdate> {
    const response = await api.put(`${PROJECT_ENDPOINT}/${projectId}/daily-updates/${updateId}`, data);
    return response.data;
  },

  async deleteDailyUpdate(projectId: string, updateId: string): Promise<void> {
    await api.delete(`${PROJECT_ENDPOINT}/${projectId}/daily-updates/${updateId}`);
  },

  // Weekly Reports
  async getWeeklyReports(projectId: string): Promise<WeeklyReport[]> {
    const response = await api.get(`${PROJECT_ENDPOINT}/${projectId}/weekly-reports`);
    return response.data;
  },

  async createWeeklyReport(projectId: string, data: Partial<WeeklyReport>): Promise<WeeklyReport> {
    const response = await api.post(`${PROJECT_ENDPOINT}/${projectId}/weekly-reports`, data);
    return response.data;
  },

  async updateWeeklyReport(projectId: string, reportId: string, data: Partial<WeeklyReport>): Promise<WeeklyReport> {
    const response = await api.put(`${PROJECT_ENDPOINT}/${projectId}/weekly-reports/${reportId}`, data);
    return response.data;
  },

  async deleteWeeklyReport(projectId: string, reportId: string): Promise<void> {
    await api.delete(`${PROJECT_ENDPOINT}/${projectId}/weekly-reports/${reportId}`);
  },

  // Timeline Schedule
  async getTimelineSchedules(projectId: string): Promise<TimelineSchedule[]> {
    const response = await api.get(`${PROJECT_ENDPOINT}/${projectId}/timeline-schedules`);
    return response.data;
  },

  async getTimelineSchedule(projectId: string, scheduleId: string): Promise<TimelineSchedule> {
    const response = await api.get(`${PROJECT_ENDPOINT}/${projectId}/timeline-schedules/${scheduleId}`);
    return response.data;
  },

  async createTimelineSchedule(projectId: string, data: Partial<TimelineSchedule>): Promise<TimelineSchedule> {
    const response = await api.post(`${PROJECT_ENDPOINT}/${projectId}/timeline-schedules`, data);
    return response.data;
  },

  async updateTimelineSchedule(projectId: string, scheduleId: string, data: Partial<TimelineSchedule>): Promise<TimelineSchedule> {
    const response = await api.put(`${PROJECT_ENDPOINT}/${projectId}/timeline-schedules/${scheduleId}`, data);
    return response.data;
  },

  async deleteTimelineSchedule(projectId: string, scheduleId: string): Promise<void> {
    await api.delete(`${PROJECT_ENDPOINT}/${projectId}/timeline-schedules/${scheduleId}`);
  },

  async updateTimelineScheduleStatus(projectId: string, scheduleId: string, status: string): Promise<TimelineSchedule> {
    const response = await api.patch(`${PROJECT_ENDPOINT}/${projectId}/timeline-schedules/${scheduleId}/status`, { status });
    return response.data;
  },

  // Project Budgets (project_budgets)
  async getProjectBudgets(projectId: string | number): Promise<ProjectBudget[]> {
    const response = await api.get(`${PROJECT_ENDPOINT}/${projectId}/budgets`);
    return response.data;
  },

  async upsertProjectBudgets(projectId: string | number, items: ProjectBudgetInput[]): Promise<void> {
    await api.post(`${PROJECT_ENDPOINT}/${projectId}/budgets`, items);
  },

  async deleteProjectBudget(projectId: string | number, budgetId: number): Promise<void> {
    await api.delete(`${PROJECT_ENDPOINT}/${projectId}/budgets/${budgetId}`);
  },

  async getProjectCostSummary(projectId: string | number): Promise<ProjectCostSummary> {
    const response = await api.get(`${PROJECT_ENDPOINT}/${projectId}/cost-summary`);
    return response.data;
  },

  // Project progress history (project_progress)
  async getProjectProgressHistory(
    projectId: string | number,
    params?: { start_date?: string; end_date?: string },
  ): Promise<ProjectProgressEntry[]> {
    const response = await api.get(`${PROJECT_ENDPOINT}/${projectId}/progress-history`, { params });
    return response.data;
  },

  async createProjectProgressSnapshot(
    projectId: string | number,
    payload: {
      date: string;
      physical_progress_percent: number;
      volume_achieved?: number;
      remarks?: string;
    },
  ): Promise<ProjectProgressEntry> {
    const response = await api.post(`${PROJECT_ENDPOINT}/${projectId}/progress-history`, payload);
    return response.data;
  },

  // Project actual costs (project_actual_costs)
  async getProjectActualCosts(
    projectId: string | number,
    params?: { start_date?: string; end_date?: string },
  ): Promise<ProjectActualCost[]> {
    const response = await api.get(`${PROJECT_ENDPOINT}/${projectId}/actual-costs`, { params });
    return response.data;
  },

  // Technical Data
  async getTechnicalData(projectId: string): Promise<TechnicalData[]> {
    const response = await api.get(`${PROJECT_ENDPOINT}/${projectId}/technical-data`);
    return response.data;
  },

  async createTechnicalData(projectId: string, data: Partial<TechnicalData>): Promise<TechnicalData> {
    const response = await api.post(`${PROJECT_ENDPOINT}/${projectId}/technical-data`, data);
    return response.data;
  },

  async updateTechnicalData(projectId: string, dataId: string, data: Partial<TechnicalData>): Promise<TechnicalData> {
    const response = await api.put(`${PROJECT_ENDPOINT}/${projectId}/technical-data/${dataId}`, data);
    return response.data;
  },

  async deleteTechnicalData(projectId: string, dataId: string): Promise<void> {
    await api.delete(`${PROJECT_ENDPOINT}/${projectId}/technical-data/${dataId}`);
  },
};

export default projectService;

