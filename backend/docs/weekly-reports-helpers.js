/**
 * Weekly Reports PDF Download Helpers
 * 
 * Helper functions untuk download weekly reports dalam format PDF
 * Dapat digunakan di React, Vue, atau vanilla JavaScript
 */

// Configuration - sesuaikan dengan environment Anda
const API_BASE_URL = 'http://localhost:8080/api/v1';

/**
 * Get authentication token from storage
 * Sesuaikan dengan state management Anda (Redux, Context, Vuex, dll)
 */
function getAuthToken() {
  // Option 1: dari localStorage
  return localStorage.getItem('token');
  
  // Option 2: dari sessionStorage
  // return sessionStorage.getItem('token');
  
  // Option 3: dari cookie
  // return document.cookie.split('; ').find(row => row.startsWith('token='))?.split('=')[1];
}

/**
 * Download individual weekly report PDF
 * @param {number} projectId - ID of the project
 * @param {number} reportId - ID of the weekly report
 * @returns {Promise<boolean>} Success status
 */
async function downloadWeeklyReportPDF(projectId, reportId) {
  const token = getAuthToken();
  
  if (!token) {
    throw new Error('Authentication token not found. Please login first.');
  }

  try {
    console.log(`Downloading PDF for report ${reportId}...`);
    
    const response = await fetch(
      `${API_BASE_URL}/projects/${projectId}/weekly-reports/${reportId}/pdf`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      }
    );

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

    // Get filename from Content-Disposition header if available
    const contentDisposition = response.headers.get('Content-Disposition');
    let filename = `weekly_report_${reportId}.pdf`;
    
    if (contentDisposition) {
      const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(contentDisposition);
      if (matches && matches[1]) {
        filename = matches[1].replace(/['"]/g, '');
      }
    }

    // Convert response to blob
    const blob = await response.blob();
    
    // Create and trigger download
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    
    // Cleanup
    setTimeout(() => {
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    }, 100);
    
    console.log(`PDF downloaded successfully: ${filename}`);
    return true;
  } catch (error) {
    console.error('Error downloading PDF:', error);
    throw error;
  }
}

/**
 * Export all weekly reports as ZIP file
 * @param {number} projectId - ID of the project
 * @param {number|null} year - Optional: filter by year (e.g., 2025)
 * @returns {Promise<boolean>} Success status
 */
async function exportAllWeeklyReportsPDF(projectId, year = null) {
  const token = getAuthToken();
  
  if (!token) {
    throw new Error('Authentication token not found. Please login first.');
  }

  try {
    // Build URL with optional year parameter
    let url = `${API_BASE_URL}/projects/${projectId}/weekly-reports/export-all`;
    if (year) {
      url += `?year=${year}`;
    }
    
    console.log(`Exporting all PDFs${year ? ` for year ${year}` : ''}...`);
    
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

    // Get filename from Content-Disposition header if available
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

    // Convert response to blob
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
    
    console.log(`ZIP file downloaded successfully: ${filename}`);
    return true;
  } catch (error) {
    console.error('Error exporting PDFs:', error);
    throw error;
  }
}

/**
 * Download weekly report with loading and error handling
 * Wrapper function dengan UI feedback
 * @param {number} projectId 
 * @param {number} reportId 
 * @param {Function} onStart - Callback saat mulai download
 * @param {Function} onSuccess - Callback saat sukses
 * @param {Function} onError - Callback saat error
 */
async function downloadWeeklyReportWithFeedback(projectId, reportId, onStart, onSuccess, onError) {
  try {
    if (onStart) onStart();
    await downloadWeeklyReportPDF(projectId, reportId);
    if (onSuccess) onSuccess();
  } catch (error) {
    if (onError) onError(error);
  }
}

/**
 * Export all weekly reports with loading and error handling
 * Wrapper function dengan UI feedback
 * @param {number} projectId 
 * @param {number|null} year 
 * @param {Function} onStart - Callback saat mulai export
 * @param {Function} onSuccess - Callback saat sukses
 * @param {Function} onError - Callback saat error
 */
async function exportAllWeeklyReportsWithFeedback(projectId, year, onStart, onSuccess, onError) {
  try {
    if (onStart) onStart();
    await exportAllWeeklyReportsPDF(projectId, year);
    if (onSuccess) onSuccess();
  } catch (error) {
    if (onError) onError(error);
  }
}

// ===== USAGE EXAMPLES =====

/**
 * Example 1: Simple usage
 */
function example1_simple() {
  const projectId = 1;
  const reportId = 5;
  
  // Download single PDF
  downloadWeeklyReportPDF(projectId, reportId)
    .then(() => alert('PDF downloaded!'))
    .catch(error => alert('Error: ' + error.message));
  
  // Export all PDFs
  exportAllWeeklyReportsPDF(projectId)
    .then(() => alert('All PDFs exported!'))
    .catch(error => alert('Error: ' + error.message));
}

/**
 * Example 2: With UI feedback (vanilla JavaScript)
 */
function example2_withFeedback() {
  const projectId = 1;
  const reportId = 5;
  
  const downloadBtn = document.getElementById('download-btn');
  const exportAllBtn = document.getElementById('export-all-btn');
  const statusDiv = document.getElementById('status');
  
  // Download single PDF
  downloadBtn.addEventListener('click', async () => {
    downloadBtn.disabled = true;
    statusDiv.textContent = 'Downloading PDF...';
    
    try {
      await downloadWeeklyReportPDF(projectId, reportId);
      statusDiv.textContent = 'PDF downloaded successfully!';
      statusDiv.style.color = 'green';
    } catch (error) {
      statusDiv.textContent = 'Error: ' + error.message;
      statusDiv.style.color = 'red';
    } finally {
      downloadBtn.disabled = false;
    }
  });
  
  // Export all PDFs
  exportAllBtn.addEventListener('click', async () => {
    exportAllBtn.disabled = true;
    statusDiv.textContent = 'Exporting all PDFs...';
    
    try {
      await exportAllWeeklyReportsPDF(projectId);
      statusDiv.textContent = 'All PDFs exported successfully!';
      statusDiv.style.color = 'green';
    } catch (error) {
      statusDiv.textContent = 'Error: ' + error.message;
      statusDiv.style.color = 'red';
    } finally {
      exportAllBtn.disabled = false;
    }
  });
}

/**
 * Example 3: React component usage (example code)
 */
const ReactExample = `
import React, { useState } from 'react';
import { downloadWeeklyReportPDF, exportAllWeeklyReportsPDF } from './weekly-reports-helpers';

function WeeklyReportActions({ projectId, reportId }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  
  const handleDownloadPDF = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      await downloadWeeklyReportPDF(projectId, reportId);
      setSuccess('PDF downloaded successfully!');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };
  
  const handleExportAll = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      await exportAllWeeklyReportsPDF(projectId);
      setSuccess('All PDFs exported successfully!');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <div className="weekly-report-actions">
      {reportId && (
        <button 
          onClick={handleDownloadPDF} 
          disabled={loading}
          className="btn btn-primary"
        >
          {loading ? 'Downloading...' : 'Download PDF'}
        </button>
      )}
      
      <button 
        onClick={handleExportAll} 
        disabled={loading}
        className="btn btn-secondary"
      >
        {loading ? 'Exporting...' : 'Export All PDF'}
      </button>
      
      {error && <div className="alert alert-danger">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}
    </div>
  );
}

export default WeeklyReportActions;
`;

/**
 * Example 4: Vue component usage (example code)
 */
const VueExample = `
<template>
  <div class="weekly-report-actions">
    <button 
      v-if="reportId"
      @click="handleDownloadPDF" 
      :disabled="loading"
      class="btn btn-primary"
    >
      {{ loading ? 'Downloading...' : 'Download PDF' }}
    </button>
    
    <button 
      @click="handleExportAll" 
      :disabled="loading"
      class="btn btn-secondary"
    >
      {{ loading ? 'Exporting...' : 'Export All PDF' }}
    </button>
    
    <div v-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-if="success" class="alert alert-success">{{ success }}</div>
  </div>
</template>

<script>
import { downloadWeeklyReportPDF, exportAllWeeklyReportsPDF } from './weekly-reports-helpers';

export default {
  name: 'WeeklyReportActions',
  props: {
    projectId: {
      type: Number,
      required: true
    },
    reportId: {
      type: Number,
      required: false
    }
  },
  data() {
    return {
      loading: false,
      error: null,
      success: null
    };
  },
  methods: {
    async handleDownloadPDF() {
      this.loading = true;
      this.error = null;
      this.success = null;
      
      try {
        await downloadWeeklyReportPDF(this.projectId, this.reportId);
        this.success = 'PDF downloaded successfully!';
      } catch (err) {
        this.error = err.message;
      } finally {
        this.loading = false;
      }
    },
    
    async handleExportAll() {
      this.loading = true;
      this.error = null;
      this.success = null;
      
      try {
        await exportAllWeeklyReportsPDF(this.projectId);
        this.success = 'All PDFs exported successfully!';
      } catch (err) {
        this.error = err.message;
      } finally {
        this.loading = false;
      }
    }
  }
};
</script>
`;

// Export functions
if (typeof module !== 'undefined' && module.exports) {
  // CommonJS (Node.js)
  module.exports = {
    downloadWeeklyReportPDF,
    exportAllWeeklyReportsPDF,
    downloadWeeklyReportWithFeedback,
    exportAllWeeklyReportsWithFeedback
  };
}

// Export for ES6 modules (comment out if not using)
// export { downloadWeeklyReportPDF, exportAllWeeklyReportsPDF };

