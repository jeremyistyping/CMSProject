"use client";

import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable'; // Import the autotable plugin to register with jsPDF
import { SSOTCashFlowData } from '../services/ssotCashFlowReportService';
import { exportCashFlowToPDF } from './cashFlowExportUtils';

// Extend jsPDF type to include autoTable functionality
declare module 'jspdf' {
  interface jsPDF {
    autoTable: any;
    lastAutoTable: {
      finalY: number;
    };
  }
}

// Download PDF in the browser
export function downloadCashFlowPDF(doc: jsPDF, filename?: string): void {
  const pdfName = filename || `cash_flow_${new Date().toISOString().split('T')[0]}.pdf`;
  doc.save(pdfName);
}

// Export Cash Flow as PDF and trigger download (client-only)
export function exportAndDownloadCashFlowPDF(
  data: SSOTCashFlowData,
  options?: {
    companyName?: string;
    logoUrl?: string;
    includeAccountDetails?: boolean;
    filename?: string;
  }
): void {
  const doc = exportCashFlowToPDF(data, options);
  downloadCashFlowPDF(doc, options?.filename);
}