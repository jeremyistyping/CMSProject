import axios from 'axios';
import apiClient from './apiClient';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export interface WeeklyReportDTO {
  id: number;
  project_id: number;
  project_name: string;
  project_code?: string;
  week: number;
  year: number;
  week_label: string;
  project_manager: string;
  total_work_days: number;
  weather_delays: number;
  team_size: number;
  accomplishments: string;
  challenges: string;
  next_week_priorities: string;
  generated_date: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateWeeklyReportRequest {
  project_id: number;
  week: number;
  year: number;
  project_manager: string;
  total_work_days: number;
  weather_delays: number;
  team_size: number;
  accomplishments: string;
  challenges: string;
  next_week_priorities: string;
  created_by?: string;
}

export interface UpdateWeeklyReportRequest {
  project_manager?: string;
  total_work_days?: number;
  weather_delays?: number;
  team_size?: number;
  accomplishments?: string;
  challenges?: string;
  next_week_priorities?: string;
}

class WeeklyReportService {
  async getWeeklyReports(projectId: number): Promise<WeeklyReportDTO[]> {
    try {
      const response = await apiClient.get(`/projects/${projectId}/weekly-reports`);
      return response.data.data || [];
    } catch (error) {
      console.error('Error fetching weekly reports:', error);
      throw error;
    }
  }

  async getWeeklyReportsByYear(projectId: number, year: number): Promise<WeeklyReportDTO[]> {
    try {
      const response = await apiClient.get(`/projects/${projectId}/weekly-reports?year=${year}`);
      return response.data.data || [];
    } catch (error) {
      console.error('Error fetching weekly reports by year:', error);
      throw error;
    }
  }

  async getWeeklyReport(projectId: number, reportId: number): Promise<WeeklyReportDTO> {
    try {
      const response = await apiClient.get(`/projects/${projectId}/weekly-reports/${reportId}`);
      return response.data.data;
    } catch (error) {
      console.error('Error fetching weekly report:', error);
      throw error;
    }
  }

  async createWeeklyReport(
    projectId: number,
    data: CreateWeeklyReportRequest
  ): Promise<WeeklyReportDTO> {
    try {
      const response = await apiClient.post(`/projects/${projectId}/weekly-reports`, data);
      return response.data.data;
    } catch (error) {
      console.error('Error creating weekly report:', error);
      throw error;
    }
  }

  async updateWeeklyReport(
    projectId: number,
    reportId: number,
    data: UpdateWeeklyReportRequest
  ): Promise<WeeklyReportDTO> {
    try {
      const response = await apiClient.put(
        `/projects/${projectId}/weekly-reports/${reportId}`,
        data
      );
      return response.data.data;
    } catch (error) {
      console.error('Error updating weekly report:', error);
      throw error;
    }
  }

  async deleteWeeklyReport(projectId: number, reportId: number): Promise<void> {
    try {
      await apiClient.delete(`/projects/${projectId}/weekly-reports/${reportId}`);
    } catch (error) {
      console.error('Error deleting weekly report:', error);
      throw error;
    }
  }

  getPDFUrl(projectId: number, reportId: number): string {
    return `${API_BASE}/projects/${projectId}/weekly-reports/${reportId}/pdf`;
  }

  // Helper to get current week number
  getCurrentWeek(): number {
    const now = new Date();
    const start = new Date(now.getFullYear(), 0, 1);
    const diff = now.getTime() - start.getTime();
    const oneWeek = 1000 * 60 * 60 * 24 * 7;
    return Math.ceil(diff / oneWeek);
  }
}

export default new WeeklyReportService();

