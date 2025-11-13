import axios from 'axios';
import apiClient from './apiClient';

// Use Next.js proxy to avoid CORS issues
// Next.js will rewrite /api/* to backend automatically
const API_BASE = '/api/v1';

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
      const response = await apiClient.get(`/api/v1/projects/${projectId}/weekly-reports`);
      // Handle null response safely
      if (!response || !response.data) {
        console.warn('Empty response from weekly reports API');
        return [];
      }
      return response.data.data || [];
    } catch (error) {
      console.error('Error fetching weekly reports:', error);
      // Return empty array instead of throwing to prevent UI crash
      return [];
    }
  }

  async getWeeklyReportsByYear(projectId: number, year: number): Promise<WeeklyReportDTO[]> {
    try {
      const response = await apiClient.get(`/api/v1/projects/${projectId}/weekly-reports?year=${year}`);
      // Handle null response safely
      if (!response || !response.data) {
        console.warn('Empty response from weekly reports by year API');
        return [];
      }
      return response.data.data || [];
    } catch (error) {
      console.error('Error fetching weekly reports by year:', error);
      // Return empty array instead of throwing to prevent UI crash
      return [];
    }
  }

  async getWeeklyReport(projectId: number, reportId: number): Promise<WeeklyReportDTO> {
    try {
      const response = await apiClient.get(`/api/v1/projects/${projectId}/weekly-reports/${reportId}`);
      // Handle null response safely
      if (!response || !response.data || !response.data.data) {
        throw new Error('Weekly report not found or invalid response');
      }
      return response.data.data;
    } catch (error: any) {
      console.error('Error fetching weekly report:', error);
      throw new Error(error.message || 'Failed to fetch weekly report');
    }
  }

  async createWeeklyReport(
    projectId: number,
    data: CreateWeeklyReportRequest
  ): Promise<WeeklyReportDTO> {
    try {
      const response = await apiClient.post(`/api/v1/projects/${projectId}/weekly-reports`, data);
      // Handle null response safely
      if (!response || !response.data) {
        throw new Error('Failed to create weekly report - empty response from server');
      }
      if (!response.data.data) {
        throw new Error(response.data.error || response.data.message || 'Failed to create weekly report');
      }
      return response.data.data;
    } catch (error: any) {
      console.error('Error creating weekly report:', error);
      // Extract meaningful error message
      const errorMessage = error.response?.data?.error 
        || error.response?.data?.details 
        || error.response?.data?.message
        || error.message 
        || 'Failed to create weekly report';
      throw new Error(errorMessage);
    }
  }

  async updateWeeklyReport(
    projectId: number,
    reportId: number,
    data: UpdateWeeklyReportRequest
  ): Promise<WeeklyReportDTO> {
    try {
      const response = await apiClient.put(
        `/api/v1/projects/${projectId}/weekly-reports/${reportId}`,
        data
      );
      // Handle null response safely
      if (!response || !response.data || !response.data.data) {
        throw new Error('Failed to update weekly report - invalid response');
      }
      return response.data.data;
    } catch (error: any) {
      console.error('Error updating weekly report:', error);
      const errorMessage = error.response?.data?.error 
        || error.response?.data?.details 
        || error.message 
        || 'Failed to update weekly report';
      throw new Error(errorMessage);
    }
  }

  async deleteWeeklyReport(projectId: number, reportId: number): Promise<void> {
    try {
      await apiClient.delete(`/api/v1/projects/${projectId}/weekly-reports/${reportId}`);
    } catch (error: any) {
      console.error('Error deleting weekly report:', error);
      const errorMessage = error.response?.data?.error 
        || error.response?.data?.details 
        || error.message 
        || 'Failed to delete weekly report';
      throw new Error(errorMessage);
    }
  }

  getPDFUrl(projectId: number, reportId: number): string {
    return `${API_BASE}/projects/${projectId}/weekly-reports/${reportId}/pdf`;
  }

  /**
   * Download individual weekly report PDF
   * Automatically triggers browser download with proper filename
   * Compatible with download managers (IDM, etc.)
   * Uses Next.js proxy to avoid CORS issues
   */
  async downloadWeeklyReportPDF(projectId: number, reportId: number): Promise<void> {
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
      
      if (!token) {
        throw new Error('Authentication required. Please login first.');
      }

      // Use relative path - Next.js will proxy to backend
      const url = `${API_BASE}/projects/${projectId}/weekly-reports/${reportId}/pdf`;
      
      const response = await fetch(url, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (!response.ok) {
        if (response.status === 401) {
          throw new Error('Unauthorized. Please login again.');
        }
        if (response.status === 404) {
          throw new Error('Weekly report not found.');
        }
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      // Get filename from Content-Disposition header
      const contentDisposition = response.headers.get('Content-Disposition');
      let filename = `weekly_report_${reportId}.pdf`;
      
      if (contentDisposition) {
        const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(contentDisposition);
        if (matches && matches[1]) {
          filename = matches[1].replace(/['"]/g, '');
        }
      }

      // Try to convert response to blob
      // If download manager (IDM) intercepts, blob() might fail - that's OK
      try {
        const blob = await response.blob();
        
        // Create and trigger download
        const downloadUrl = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = downloadUrl;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        
        // Cleanup
        setTimeout(() => {
          document.body.removeChild(a);
          window.URL.revokeObjectURL(downloadUrl);
        }, 100);
      } catch (blobError) {
        // Download manager likely intercepted the download
        // File is already downloading, so this is not an actual error
        console.log('Download handled by download manager (IDM/browser):', filename);
        // Don't throw error - download is successful via download manager
      }
    } catch (error: any) {
      // Only throw if it's an actual API error (401, 404, etc.)
      console.error('Error downloading PDF:', error);
      throw error;
    }
  }

  /**
   * Export all weekly reports as ZIP file
   * Compatible with download managers (IDM, etc.)
   * @param projectId - ID of the project
   * @param year - Optional: filter by specific year
   */
  async exportAllWeeklyReportsPDF(projectId: number, year?: number): Promise<void> {
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
      
      if (!token) {
        throw new Error('Authentication required. Please login first.');
      }

      // Build URL with optional year parameter
      let url = `${API_BASE}/projects/${projectId}/weekly-reports/export-all`;
      if (year) {
        url += `?year=${year}`;
      }
      
      const response = await fetch(url, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (!response.ok) {
        if (response.status === 401) {
          throw new Error('Unauthorized. Please login again.');
        }
        if (response.status === 404) {
          const errorData = await response.json().catch(() => ({}));
          throw new Error(errorData.message || 'No weekly reports found for this project.');
        }
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      // Get filename from Content-Disposition header
      const contentDisposition = response.headers.get('Content-Disposition');
      let filename = year 
        ? `weekly_reports_${year}.zip` 
        : `weekly_reports_all.zip`;
      
      if (contentDisposition) {
        const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(contentDisposition);
        if (matches && matches[1]) {
          filename = matches[1].replace(/['"]/g, '');
        }
      }

      // Try to convert response to blob
      // If download manager (IDM) intercepts, blob() might fail - that's OK
      try {
        const blob = await response.blob();
        
        // Create and trigger download
        const downloadUrl = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = downloadUrl;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        
        // Cleanup
        setTimeout(() => {
          document.body.removeChild(a);
          window.URL.revokeObjectURL(downloadUrl);
        }, 100);
      } catch (blobError) {
        // Download manager likely intercepted the download
        // File is already downloading, so this is not an actual error
        console.log('Download handled by download manager (IDM/browser):', filename);
        // Don't throw error - download is successful via download manager
      }
    } catch (error: any) {
      // Only throw if it's an actual API error (401, 404, etc.)
      console.error('Error exporting PDFs:', error);
      throw error;
    }
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

