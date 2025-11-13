# Weekly Reports Frontend Implementation Guide

## Overview
This guide explains how to implement the frontend for the Weekly Reports feature. The backend API is ready and available at `/api/v1/projects/:id/weekly-reports`.

## Backend API Endpoints

```
GET    /api/v1/projects/:id/weekly-reports          - List all reports for a project
GET    /api/v1/projects/:id/weekly-reports/:reportId - Get specific report
POST   /api/v1/projects/:id/weekly-reports          - Create new report
PUT    /api/v1/projects/:id/weekly-reports/:reportId - Update report
DELETE /api/v1/projects/:id/weekly-reports/:reportId - Delete report
GET    /api/v1/projects/:id/weekly-reports/:reportId/pdf - Download PDF
```

## Frontend Component Structure

```
frontend/src/
├── components/
│   └── projects/
│       ├── WeeklyReports.tsx (or .jsx)     - Main component
│       ├── WeeklyReportForm.tsx            - Form for create/edit
│       └── WeeklyReportList.tsx            - List of reports
└── services/
    └── weeklyReportService.ts (or .js)     - API service calls
```

## 1. API Service (`weeklyReportService.ts`)

```typescript
import axios from 'axios';

const API_BASE = '/api/v1';

export interface WeeklyReportDTO {
  id: number;
  project_id: number;
  project_name: string;
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
}

class WeeklyReportService {
  async getWeeklyReports(projectId: number): Promise<WeeklyReportDTO[]> {
    const response = await axios.get(`${API_BASE}/projects/${projectId}/weekly-reports`);
    return response.data.data;
  }

  async getWeeklyReport(projectId: number, reportId: number): Promise<WeeklyReportDTO> {
    const response = await axios.get(`${API_BASE}/projects/${projectId}/weekly-reports/${reportId}`);
    return response.data.data;
  }

  async createWeeklyReport(projectId: number, data: CreateWeeklyReportRequest): Promise<WeeklyReportDTO> {
    const response = await axios.post(`${API_BASE}/projects/${projectId}/weekly-reports`, data);
    return response.data.data;
  }

  async updateWeeklyReport(projectId: number, reportId: number, data: Partial<CreateWeeklyReportRequest>): Promise<WeeklyReportDTO> {
    const response = await axios.put(`${API_BASE}/projects/${projectId}/weekly-reports/${reportId}`, data);
    return response.data.data;
  }

  async deleteWeeklyReport(projectId: number, reportId: number): Promise<void> {
    await axios.delete(`${API_BASE}/projects/${projectId}/weekly-reports/${reportId}`);
  }

  getPDFUrl(projectId: number, reportId: number): string {
    return `${API_BASE}/projects/${projectId}/weekly-reports/${reportId}/pdf`;
  }
}

export default new WeeklyReportService();
```

## 2. Main Component (`WeeklyReports.tsx`)

```typescript
import React, { useState, useEffect } from 'react';
import weeklyReportService, { WeeklyReportDTO } from '../../services/weeklyReportService';
import WeeklyReportForm from './WeeklyReportForm';
import WeeklyReportList from './WeeklyReportList';

interface WeeklyReportsProps {
  projectId: number;
  projectName: string;
}

const WeeklyReports: React.FC<WeeklyReportsProps> = ({ projectId, projectName }) => {
  const [reports, setReports] = useState<WeeklyReportDTO[]>([]);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(true); // Show form by default
  const [selectedReport, setSelectedReport] = useState<WeeklyReportDTO | null>(null);

  useEffect(() => {
    loadReports();
  }, [projectId]);

  const loadReports = async () => {
    setLoading(true);
    try {
      const data = await weeklyReportService.getWeeklyReports(projectId);
      setReports(data);
    } catch (error) {
      console.error('Failed to load reports:', error);
      alert('Failed to load weekly reports');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateReport = async (formData: any) => {
    try {
      await weeklyReportService.createWeeklyReport(projectId, formData);
      alert('Weekly report created successfully');
      loadReports();
      setShowForm(true); // Keep form visible for next report
    } catch (error: any) {
      console.error('Failed to create report:', error);
      alert(error.response?.data?.details || 'Failed to create report');
    }
  };

  const handleDownloadPDF = (reportId: number) => {
    const url = weeklyReportService.getPDFUrl(projectId, reportId);
    window.open(url, '_blank');
  };

  const handleDelete = async (reportId: number) => {
    if (!confirm('Are you sure you want to delete this report?')) return;
    
    try {
      await weeklyReportService.deleteWeeklyReport(projectId, reportId);
      alert('Report deleted successfully');
      loadReports();
    } catch (error) {
      console.error('Failed to delete report:', error);
      alert('Failed to delete report');
    }
  };

  return (
    <div className="weekly-reports-container">
      <h2>Weekly Reports - {projectName}</h2>
      
      {/* Form Section */}
      <div className="report-form-section">
        <h3>Generate Weekly Report</h3>
        <WeeklyReportForm 
          projectId={projectId}
          onSubmit={handleCreateReport}
        />
      </div>

      {/* Previous Reports Section */}
      <div className="previous-reports-section">
        <div className="section-header">
          <h3>Previous Reports</h3>
          <button 
            className="btn-export-all" 
            onClick={() => alert('Export all feature - coming soon')}
          >
            Export All PDF
          </button>
        </div>

        {loading ? (
          <div className="loading">Loading reports...</div>
        ) : reports.length === 0 ? (
          <div className="no-reports">
            <div className="empty-state-icon">📊</div>
            <p>No weekly reports yet</p>
            <p className="text-muted">Generate your first weekly report to track progress</p>
          </div>
        ) : (
          <WeeklyReportList 
            reports={reports}
            onDownloadPDF={handleDownloadPDF}
            onDelete={handleDelete}
          />
        )}
      </div>
    </div>
  );
};

export default WeeklyReports;
```

## 3. Form Component (`WeeklyReportForm.tsx`)

```typescript
import React, { useState } from 'react';

interface WeeklyReportFormProps {
  projectId: number;
  onSubmit: (data: any) => Promise<void>;
}

const WeeklyReportForm: React.FC<WeeklyReportFormProps> = ({ projectId, onSubmit }) => {
  const [formData, setFormData] = useState({
    week: getCurrentWeek(),
    year: new Date().getFullYear(),
    project_manager: '',
    total_work_days: 5,
    weather_delays: 0,
    team_size: 1,
    accomplishments: '',
    challenges: '',
    next_week_priorities: '',
  });

  const [submitting, setSubmitting] = useState(false);

  function getCurrentWeek(): number {
    const now = new Date();
    const start = new Date(now.getFullYear(), 0, 1);
    const diff = now.getTime() - start.getTime();
    const oneWeek = 1000 * 60 * 60 * 24 * 7;
    return Math.ceil(diff / oneWeek);
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: name.includes('days') || name.includes('delays') || name.includes('size') 
        ? parseInt(value) || 0 
        : value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({ ...formData, project_id: projectId });
      // Reset form after successful submission
      setFormData({
        ...formData,
        accomplishments: '',
        challenges: '',
        next_week_priorities: '',
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="weekly-report-form">
      <div className="form-row">
        <div className="form-group">
          <label>Report Week</label>
          <input
            type="number"
            name="week"
            value={formData.week}
            onChange={handleChange}
            min="1"
            max="53"
            required
          />
        </div>
        <div className="form-group">
          <label>Year</label>
          <input
            type="number"
            name="year"
            value={formData.year}
            onChange={handleChange}
            min="2000"
            required
          />
        </div>
      </div>

      <div className="form-group">
        <label>Project Manager</label>
        <input
          type="text"
          name="project_manager"
          value={formData.project_manager}
          onChange={handleChange}
          placeholder="Manager name"
        />
      </div>

      <div className="form-row">
        <div className="form-group">
          <label>Total Work Days</label>
          <input
            type="number"
            name="total_work_days"
            value={formData.total_work_days}
            onChange={handleChange}
            min="0"
            required
          />
        </div>
        <div className="form-group">
          <label>Weather Delays</label>
          <input
            type="number"
            name="weather_delays"
            value={formData.weather_delays}
            onChange={handleChange}
            min="0"
          />
        </div>
        <div className="form-group">
          <label>Team Size</label>
          <input
            type="number"
            name="team_size"
            value={formData.team_size}
            onChange={handleChange}
            min="1"
            required
          />
        </div>
      </div>

      <div className="form-group">
        <label>Major Accomplishments</label>
        <textarea
          name="accomplishments"
          value={formData.accomplishments}
          onChange={handleChange}
          rows={4}
          placeholder="List major accomplishments this week..."
        />
      </div>

      <div className="form-group">
        <label>Challenges & Issues</label>
        <textarea
          name="challenges"
          value={formData.challenges}
          onChange={handleChange}
          rows={4}
          placeholder="Describe any challenges encountered..."
        />
      </div>

      <div className="form-group">
        <label>Next Week's Priorities</label>
        <textarea
          name="next_week_priorities"
          value={formData.next_week_priorities}
          onChange={handleChange}
          rows={4}
          placeholder="List next week's priorities..."
        />
      </div>

      <button 
        type="submit" 
        className="btn-submit"
        disabled={submitting}
      >
        {submitting ? 'Generating...' : 'Generate Report'}
      </button>
    </form>
  );
};

export default WeeklyReportForm;
```

## 4. List Component (`WeeklyReportList.tsx`)

```typescript
import React from 'react';
import { WeeklyReportDTO } from '../../services/weeklyReportService';

interface WeeklyReportListProps {
  reports: WeeklyReportDTO[];
  onDownloadPDF: (reportId: number) => void;
  onDelete: (reportId: number) => void;
}

const WeeklyReportList: React.FC<WeeklyReportListProps> = ({ reports, onDownloadPDF, onDelete }) => {
  return (
    <div className="report-list">
      {reports.map((report) => (
        <div key={report.id} className="report-card">
          <div className="report-header">
            <div className="report-title">
              <h4>{report.week_label}</h4>
              <span className="report-date">
                Generated: {new Date(report.generated_date).toLocaleDateString()}
              </span>
            </div>
            <div className="report-actions">
              <button 
                className="btn-icon btn-download"
                onClick={() => onDownloadPDF(report.id)}
                title="Download PDF"
              >
                📥 PDF
              </button>
              <button 
                className="btn-icon btn-delete"
                onClick={() => onDelete(report.id)}
                title="Delete"
              >
                🗑️
              </button>
            </div>
          </div>

          <div className="report-stats">
            <div className="stat">
              <span className="stat-label">Work Days:</span>
              <span className="stat-value">{report.total_work_days}</span>
            </div>
            <div className="stat">
              <span className="stat-label">Weather Delays:</span>
              <span className="stat-value">{report.weather_delays}</span>
            </div>
            <div className="stat">
              <span className="stat-label">Team Size:</span>
              <span className="stat-value">{report.team_size}</span>
            </div>
            <div className="stat">
              <span className="stat-label">Manager:</span>
              <span className="stat-value">{report.project_manager || 'N/A'}</span>
            </div>
          </div>

          {report.accomplishments && (
            <div className="report-section">
              <strong>Accomplishments:</strong>
              <p className="report-text">{report.accomplishments}</p>
            </div>
          )}
        </div>
      ))}
    </div>
  );
};

export default WeeklyReportList;
```

## 5. CSS Styling (`WeeklyReports.css`)

```css
.weekly-reports-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.report-form-section {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 32px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.weekly-report-form .form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-weight: 600;
  margin-bottom: 8px;
  color: #333;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-group textarea {
  resize: vertical;
  font-family: inherit;
}

.btn-submit {
  background: #4F46E5;
  color: white;
  padding: 12px 32px;
  border: none;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-submit:hover:not(:disabled) {
  background: #4338CA;
}

.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.previous-reports-section {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.btn-export-all {
  background: #10B981;
  color: white;
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.no-reports {
  text-align: center;
  padding: 60px 20px;
  color: #6B7280;
}

.empty-state-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.report-list {
  display: grid;
  gap: 16px;
}

.report-card {
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  padding: 20px;
  background: #F9FAFB;
  transition: box-shadow 0.2s;
}

.report-card:hover {
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.report-header {
  display: flex;
  justify-content: space-between;
  align-items: start;
  margin-bottom: 16px;
}

.report-title h4 {
  margin: 0 0 4px 0;
  color: #111827;
}

.report-date {
  font-size: 12px;
  color: #6B7280;
}

.report-actions {
  display: flex;
  gap: 8px;
}

.btn-icon {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: opacity 0.2s;
}

.btn-download {
  background: #3B82F6;
  color: white;
}

.btn-delete {
  background: #EF4444;
  color: white;
}

.btn-icon:hover {
  opacity: 0.8;
}

.report-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: white;
  border-radius: 6px;
}

.stat {
  display: flex;
  flex-direction: column;
}

.stat-label {
  font-size: 12px;
  color: #6B7280;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

.report-section {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #E5E7EB;
}

.report-section strong {
  display: block;
  margin-bottom: 8px;
  color: #374151;
}

.report-text {
  margin: 0;
  line-height: 1.6;
  color: #6B7280;
  white-space: pre-wrap;
}
```

## 6. Integration into Project Detail Page

Add the WeeklyReports component as a tab in your project detail page:

```typescript
import WeeklyReports from './WeeklyReports';

// In your Project Detail component:
<Tabs>
  <Tab label="Dashboard">...</Tab>
  <Tab label="Daily Updates">...</Tab>
  <Tab label="Milestones">...</Tab>
  <Tab label="Weekly Reports">
    <WeeklyReports 
      projectId={projectId} 
      projectName={projectName} 
    />
  </Tab>
  <Tab label="Timeline">...</Tab>
  <Tab label="Technical Data">...</Tab>
</Tabs>
```

## Testing Steps

1. **Run database migration:**
   ```bash
   # The migration will be automatically run when you start the backend
   ```

2. **Start backend server:**
   ```bash
   cd backend
   go run cmd/main.go
   ```

3. **Test the endpoints:**
   - Navigate to your project detail page
   - Click on "Weekly Reports" tab
   - Fill out the form and submit
   - Verify the report appears in the list
   - Click "PDF" button to download the PDF report

## Next Steps

1. Implement the frontend components as described above
2. Style the components to match your application's design
3. Test all CRUD operations
4. Test PDF generation
5. Add error handling and loading states
6. Add validation for form inputs
7. Consider adding pagination for the reports list if needed

## Notes

- The backend uses `gofpdf` for PDF generation, which creates a professional-looking report
- All API endpoints are protected (add authentication if needed)
- The form automatically calculates the current week number
- Reports are soft-deleted (can be restored if needed)
- Each report is unique per project/week/year combination

