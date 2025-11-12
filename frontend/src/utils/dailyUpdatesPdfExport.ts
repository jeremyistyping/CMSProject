import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import { DailyUpdate } from '@/types/project';

// Extend jsPDF to include autoTable
declare module 'jspdf' {
  interface jsPDF {
    autoTable: (options: any) => jsPDF;
    lastAutoTable?: {
      finalY: number;
    };
  }
}

interface ProjectInfo {
  name: string;
  location: string;
  client: string;
  status: string;
  progress: number;
}

interface ExportConfig {
  projectInfo: ProjectInfo;
  dailyUpdates: DailyUpdate[];
  dateRange?: {
    startDate?: string;
    endDate?: string;
  };
}

/**
 * Generate PDF report for Daily Updates
 */
export const generateDailyUpdatesPDF = async (config: ExportConfig): Promise<void> => {
  const { projectInfo, dailyUpdates, dateRange } = config;
  
  // Create PDF document
  const doc = new jsPDF('portrait', 'mm', 'a4');
  const pageWidth = doc.internal.pageSize.width;
  const pageHeight = doc.internal.pageSize.height;
  const margin = 15;
  let yPosition = margin;

  // Title
  doc.setFontSize(18);
  doc.setFont('helvetica', 'bold');
  doc.text('Daily Updates Report', margin, yPosition);
  yPosition += 10;

  // Date Range (if provided)
  if (dateRange?.startDate || dateRange?.endDate) {
    doc.setFontSize(10);
    doc.setFont('helvetica', 'normal');
    let dateText = 'Period: ';
    if (dateRange.startDate) {
      dateText += formatDate(dateRange.startDate);
    }
    if (dateRange.endDate) {
      dateText += ` to ${formatDate(dateRange.endDate)}`;
    }
    doc.text(dateText, margin, yPosition);
    yPosition += 8;
  }

  // Horizontal line
  doc.setDrawColor(200, 200, 200);
  doc.line(margin, yPosition, pageWidth - margin, yPosition);
  yPosition += 10;

  // Project Information Section
  doc.setFontSize(14);
  doc.setFont('helvetica', 'bold');
  doc.text('Project Information', margin, yPosition);
  yPosition += 8;

  doc.setFontSize(10);
  doc.setFont('helvetica', 'normal');
  
  const projectDetails = [
    ['Project Name:', projectInfo.name],
    ['Location:', projectInfo.location],
    ['Client:', projectInfo.client],
    ['Status:', projectInfo.status],
    ['Progress:', `${projectInfo.progress}%`],
  ];

  projectDetails.forEach(([label, value]) => {
    doc.setFont('helvetica', 'bold');
    doc.text(label, margin, yPosition);
    doc.setFont('helvetica', 'normal');
    doc.text(value, margin + 35, yPosition);
    yPosition += 6;
  });

  yPosition += 5;

  // Daily Updates Summary
  doc.setFontSize(14);
  doc.setFont('helvetica', 'bold');
  doc.text('Daily Updates Summary', margin, yPosition);
  yPosition += 8;

  // Calculate statistics
  const totalUpdates = dailyUpdates.length;
  const totalWorkers = dailyUpdates.reduce((sum, update) => sum + update.workers_present, 0);
  const averageWorkers = totalUpdates > 0 ? Math.round(totalWorkers / totalUpdates) : 0;
  const updatesWithIssues = dailyUpdates.filter(update => update.issues && update.issues.trim() !== '').length;
  const totalPhotos = dailyUpdates.reduce((sum, update) => sum + (update.photos?.length || 0), 0);

  doc.setFontSize(10);
  doc.setFont('helvetica', 'normal');
  doc.text(`Total Updates: ${totalUpdates}`, margin, yPosition);
  doc.text(`Total Workers (sum): ${totalWorkers}`, margin + 60, yPosition);
  yPosition += 6;
  doc.text(`Average Workers: ${averageWorkers}/day`, margin, yPosition);
  doc.text(`Updates with Issues: ${updatesWithIssues}`, margin + 60, yPosition);
  yPosition += 6;
  doc.text(`Total Photos: ${totalPhotos}`, margin, yPosition);
  yPosition += 10;

  // Daily Updates Table
  doc.setFontSize(14);
  doc.setFont('helvetica', 'bold');
  doc.text('Daily Updates Details', margin, yPosition);
  yPosition += 8;

  // Prepare table data
  const tableData = dailyUpdates.map((update) => [
    formatDate(update.date),
    update.weather,
    update.workers_present.toString(),
    truncateText(update.work_description, 60),
    update.issues ? 'Yes' : 'No',
    (update.photos?.length || 0).toString(),
  ]);

  // Add table using autoTable (using functional approach for better compatibility)
  autoTable(doc, {
    startY: yPosition,
    head: [['Date', 'Weather', 'Workers', 'Work Description', 'Issues', 'Photos']],
    body: tableData,
    theme: 'striped',
    headStyles: {
      fillColor: [72, 187, 120], // Green color
      textColor: 255,
      fontStyle: 'bold',
      fontSize: 9,
    },
    bodyStyles: {
      fontSize: 8,
      cellPadding: 3,
    },
    columnStyles: {
      0: { cellWidth: 25 }, // Date
      1: { cellWidth: 20 }, // Weather
      2: { cellWidth: 18 }, // Workers
      3: { cellWidth: 70 }, // Work Description
      4: { cellWidth: 15 }, // Issues
      5: { cellWidth: 15 }, // Photos
    },
    margin: { left: margin, right: margin },
    didDrawPage: (data: any) => {
      // Footer with page numbers
      const pageCount = doc.getNumberOfPages();
      doc.setFontSize(8);
      doc.setFont('helvetica', 'normal');
      doc.text(
        `Page ${data.pageNumber} of ${pageCount}`,
        pageWidth / 2,
        pageHeight - 10,
        { align: 'center' }
      );
      
      // Add generation date
      doc.text(
        `Generated on: ${new Date().toLocaleDateString('en-GB')}`,
        pageWidth - margin,
        pageHeight - 10,
        { align: 'right' }
      );
    },
  });

  // Get final Y position after table
  yPosition = (doc as any).lastAutoTable?.finalY || yPosition;
  yPosition += 10;

  // Check if we need a new page for detailed updates
  if (yPosition > pageHeight - 40) {
    doc.addPage();
    yPosition = margin;
  }

  // Detailed Daily Updates (with full descriptions)
  doc.setFontSize(14);
  doc.setFont('helvetica', 'bold');
  doc.text('Detailed Daily Updates', margin, yPosition);
  yPosition += 8;

  dailyUpdates.forEach((update, index) => {
    // Check if we need a new page
    if (yPosition > pageHeight - 60) {
      doc.addPage();
      yPosition = margin;
    }

    // Date header
    doc.setFontSize(11);
    doc.setFont('helvetica', 'bold');
    doc.setTextColor(72, 187, 120); // Green
    doc.text(`${index + 1}. ${formatDate(update.date)} - ${update.weather}`, margin, yPosition);
    doc.setTextColor(0, 0, 0); // Reset to black
    yPosition += 7;

    doc.setFontSize(9);
    doc.setFont('helvetica', 'normal');

    // Workers
    doc.setFont('helvetica', 'bold');
    doc.text('Workers Present:', margin + 5, yPosition);
    doc.setFont('helvetica', 'normal');
    doc.text(update.workers_present.toString(), margin + 40, yPosition);
    yPosition += 5;

    // Work Description
    doc.setFont('helvetica', 'bold');
    doc.text('Work Description:', margin + 5, yPosition);
    yPosition += 5;
    doc.setFont('helvetica', 'normal');
    const descLines = doc.splitTextToSize(update.work_description, pageWidth - margin * 2 - 10);
    doc.text(descLines, margin + 5, yPosition);
    yPosition += descLines.length * 4 + 2;

    // Materials Used (if available)
    if (update.materials_used && update.materials_used.trim() !== '') {
      doc.setFont('helvetica', 'bold');
      doc.text('Materials Used:', margin + 5, yPosition);
      yPosition += 5;
      doc.setFont('helvetica', 'normal');
      const materialLines = doc.splitTextToSize(update.materials_used, pageWidth - margin * 2 - 10);
      doc.text(materialLines, margin + 5, yPosition);
      yPosition += materialLines.length * 4 + 2;
    }

    // Issues (if available)
    if (update.issues && update.issues.trim() !== '') {
      doc.setFont('helvetica', 'bold');
      doc.setTextColor(220, 38, 38); // Red
      doc.text('Issues:', margin + 5, yPosition);
      yPosition += 5;
      doc.setFont('helvetica', 'normal');
      doc.setTextColor(0, 0, 0); // Reset to black
      const issueLines = doc.splitTextToSize(update.issues, pageWidth - margin * 2 - 10);
      doc.text(issueLines, margin + 5, yPosition);
      yPosition += issueLines.length * 4 + 2;
    }

    // Tomorrow's Plan (if available)
    if (update.tomorrows_plan && update.tomorrows_plan.trim() !== '') {
      doc.setFont('helvetica', 'bold');
      doc.setTextColor(59, 130, 246); // Blue
      doc.text("Tomorrow's Plan:", margin + 5, yPosition);
      yPosition += 5;
      doc.setFont('helvetica', 'normal');
      doc.setTextColor(0, 0, 0); // Reset to black
      const planLines = doc.splitTextToSize(update.tomorrows_plan, pageWidth - margin * 2 - 10);
      doc.text(planLines, margin + 5, yPosition);
      yPosition += planLines.length * 4 + 2;
    }

    // Photos count
    if (update.photos && update.photos.length > 0) {
      doc.setFont('helvetica', 'bold');
      doc.text('Photos:', margin + 5, yPosition);
      doc.setFont('helvetica', 'normal');
      doc.text(`${update.photos.length} photo(s) attached`, margin + 20, yPosition);
      yPosition += 5;
    }

    // Created by
    doc.setFontSize(8);
    doc.setTextColor(100, 100, 100); // Gray
    doc.text(`Created by: ${update.created_by || 'Unknown'}`, margin + 5, yPosition);
    doc.setTextColor(0, 0, 0); // Reset to black
    yPosition += 8;

    // Separator line
    if (index < dailyUpdates.length - 1) {
      doc.setDrawColor(220, 220, 220);
      doc.line(margin, yPosition, pageWidth - margin, yPosition);
      yPosition += 6;
    }
  });

  // Save the PDF
  const fileName = `Daily_Updates_${projectInfo.name.replace(/\s+/g, '_')}_${new Date().toISOString().split('T')[0]}.pdf`;
  doc.save(fileName);
};

/**
 * Format date to readable string
 */
const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  });
};

/**
 * Truncate text to a specific length
 */
const truncateText = (text: string, maxLength: number): string => {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength - 3) + '...';
};

export default generateDailyUpdatesPDF;

