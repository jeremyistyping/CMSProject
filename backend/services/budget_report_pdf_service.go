package services

import (
	"app-sistem-akuntansi/models"
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

type BudgetReportPDFService interface {
	GenerateBudgetReportPDF(report *models.BudgetReportResponse) ([]byte, error)
}

type budgetReportPDFService struct{}

func NewBudgetReportPDFService() BudgetReportPDFService {
	return &budgetReportPDFService{}
}

func (s *budgetReportPDFService) GenerateBudgetReportPDF(report *models.BudgetReportResponse) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "BUDGET VS ACTUAL REPORT")
	pdf.Ln(8)

	// Project Info
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Project: %s", report.ProjectName))
	pdf.Ln(5)
	pdf.Cell(0, 6, fmt.Sprintf("Period: %s to %s", 
		report.StartDate.Format("02 Jan 2006"), 
		report.EndDate.Format("02 Jan 2006")))
	pdf.Ln(5)
	pdf.Cell(0, 6, fmt.Sprintf("Report Date: %s", report.ReportDate.Format("02 Jan 2006 15:04")))
	pdf.Ln(10)

	// Summary Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "SUMMARY")
	pdf.Ln(8)

	// Summary Table
	s.drawSummaryTable(pdf, report)
	pdf.Ln(10)

	// Labour Budget Detail
	if report.LabourBudget != nil && len(report.LabourBudget.Transactions) > 0 {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "LABOUR BUDGET DETAIL")
		pdf.Ln(8)
		s.drawCategoryDetail(pdf, report.LabourBudget, false)
	}

	// Operational Budget Detail
	if report.OperationalBudget != nil && len(report.OperationalBudget.Transactions) > 0 {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "OPERATIONAL BUDGET DETAIL")
		pdf.Ln(8)
		s.drawCategoryDetail(pdf, report.OperationalBudget, true)
	}

	// Other Budget Detail
	if report.OtherBudget != nil && len(report.OtherBudget.Transactions) > 0 {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "OTHER BUDGET DETAIL")
		pdf.Ln(8)
		s.drawCategoryDetail(pdf, report.OtherBudget, false)
	}

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *budgetReportPDFService) drawSummaryTable(pdf *gofpdf.Fpdf, report *models.BudgetReportResponse) {
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 200, 200)

	// Header
	pdf.CellFormat(60, 7, "Category", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 7, "Budget", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 7, "Actual", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 7, "Variance", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "%", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)

	// Labour
	if report.LabourBudget != nil {
		s.drawSummaryRow(pdf, "Labour Budget", 
			report.LabourBudget.BudgetEstimation,
			report.LabourBudget.Actual,
			report.LabourBudget.Variance)
	}

	// Operational
	if report.OperationalBudget != nil {
		s.drawSummaryRow(pdf, "Operational Budget", 
			report.OperationalBudget.BudgetEstimation,
			report.OperationalBudget.Actual,
			report.OperationalBudget.Variance)
	}

	// Other
	if report.OtherBudget != nil {
		s.drawSummaryRow(pdf, "Other Budget", 
			report.OtherBudget.BudgetEstimation,
			report.OtherBudget.Actual,
			report.OtherBudget.Variance)
	}

	// Total
	totalBudget := 0.0
	totalActual := 0.0
	if report.LabourBudget != nil {
		totalBudget += report.LabourBudget.BudgetEstimation
		totalActual += report.LabourBudget.Actual
	}
	if report.OperationalBudget != nil {
		totalBudget += report.OperationalBudget.BudgetEstimation
		totalActual += report.OperationalBudget.Actual
	}
	if report.OtherBudget != nil {
		totalBudget += report.OtherBudget.BudgetEstimation
		totalActual += report.OtherBudget.Actual
	}

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220)
	s.drawSummaryRow(pdf, "TOTAL", totalBudget, totalActual, totalBudget-totalActual)
}

func (s *budgetReportPDFService) drawSummaryRow(pdf *gofpdf.Fpdf, category string, budget, actual, variance float64) {
	percentage := 0.0
	if budget > 0 {
		percentage = (actual / budget) * 100
	}

	pdf.CellFormat(60, 7, category, "1", 0, "L", false, 0, "")
	pdf.CellFormat(45, 7, s.formatCurrency(budget), "1", 0, "R", false, 0, "")
	pdf.CellFormat(45, 7, s.formatCurrency(actual), "1", 0, "R", false, 0, "")
	
	// Variance with color
	if variance < 0 {
		pdf.SetTextColor(255, 0, 0) // Red for over budget
	} else {
		pdf.SetTextColor(0, 128, 0) // Green for under budget
	}
	pdf.CellFormat(45, 7, s.formatCurrency(variance), "1", 0, "R", false, 0, "")
	pdf.SetTextColor(0, 0, 0) // Reset to black
	
	pdf.CellFormat(30, 7, fmt.Sprintf("%.1f%%", percentage), "1", 0, "R", false, 0, "")
	pdf.Ln(-1)
}

func (s *budgetReportPDFService) drawCategoryDetail(pdf *gofpdf.Fpdf, category *models.BudgetCategoryReport, showWorkPackage bool) {
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 200, 200)

	// Header
	colWidths := []float64{25, 60, 20, 20, 35, 30}
	headers := []string{"Date", "Description", "Unit", "Qty", "Amount", "COA"}
	
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 7, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 8)

	// Group by work package if operational
	if showWorkPackage && len(category.ByWorkPackage) > 0 {
		for _, wp := range category.ByWorkPackage {
			// Work package header
			pdf.SetFont("Arial", "B", 9)
			pdf.SetFillColor(230, 230, 230)
			pdf.CellFormat(190, 6, fmt.Sprintf("Work Package: %s", wp.WorkPackage), "1", 0, "L", true, 0, "")
			pdf.Ln(-1)

			pdf.SetFont("Arial", "", 8)
			for _, tx := range wp.Transactions {
				s.drawTransactionRow(pdf, tx, colWidths)
			}

			// Work package summary
			pdf.SetFont("Arial", "B", 8)
			pdf.CellFormat(125, 6, "Subtotal", "1", 0, "R", false, 0, "")
			pdf.CellFormat(35, 6, s.formatCurrency(wp.Actual), "1", 0, "R", false, 0, "")
			pdf.CellFormat(30, 6, "", "1", 0, "L", false, 0, "")
			pdf.Ln(-1)
			pdf.SetFont("Arial", "", 8)
		}
	} else {
		// No work package grouping
		for _, tx := range category.Transactions {
			s.drawTransactionRow(pdf, tx, colWidths)
		}
	}

	// Category total
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(125, 7, "TOTAL", "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 7, s.formatCurrency(category.Actual), "1", 0, "R", true, 0, "")
	pdf.CellFormat(30, 7, "", "1", 0, "L", true, 0, "")
	pdf.Ln(-1)
}

func (s *budgetReportPDFService) drawTransactionRow(pdf *gofpdf.Fpdf, tx models.ExpenseTransactionDetail, colWidths []float64) {
	pdf.CellFormat(colWidths[0], 6, tx.Date.Format("02/01/06"), "1", 0, "C", false, 0, "")
	
	// Truncate description if too long
	desc := tx.Description
	if len(desc) > 40 {
		desc = desc[:37] + "..."
	}
	pdf.CellFormat(colWidths[1], 6, desc, "1", 0, "L", false, 0, "")
	pdf.CellFormat(colWidths[2], 6, tx.Unit, "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidths[3], 6, fmt.Sprintf("%.0f", tx.Quantity), "1", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[4], 6, s.formatCurrency(tx.TotalPrice), "1", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[5], 6, tx.COACode, "1", 0, "L", false, 0, "")
	pdf.Ln(-1)
}

func (s *budgetReportPDFService) formatCurrency(amount float64) string {
	if amount < 0 {
		return fmt.Sprintf("(Rp %.0f)", -amount)
	}
	return fmt.Sprintf("Rp %.0f", amount)
}
