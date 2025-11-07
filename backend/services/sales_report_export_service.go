package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/utils"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// SalesReportExportService handles export functionality for Sales Report
type SalesReportExportService struct{ db *gorm.DB }

// NewSalesReportExportService creates a new sales report export service
func NewSalesReportExportService(db *gorm.DB) *SalesReportExportService {
	return &SalesReportExportService{db: db}
}

// getCompanyInfo from settings with defaults
func (s *SalesReportExportService) getCompanyInfo() *models.Settings {
	if s.db == nil {
		return &models.Settings{CompanyName: "PT. Sistem Akuntansi Indonesia"}
	}
	var settings models.Settings
	if err := s.db.First(&settings).Error; err != nil {
		return &models.Settings{CompanyName: "PT. Sistem Akuntansi Indonesia"}
	}
	return &settings
}

// getCurrencyFromSettings gets currency from settings
func (s *SalesReportExportService) getCurrencyFromSettings() string {
	settings := s.getCompanyInfo()
	if settings.Currency != "" {
		return settings.Currency
	}
	return "IDR"
}

// ExportToCSV exports sales report to CSV bytes with localization
func (s *SalesReportExportService) ExportToCSV(data *SalesSummaryData, userID uint) ([]byte, error) {
	// Get user language preference
	language := utils.GetUserLanguageFromSettings(s.db)
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header with localization
	w.Write([]string{utils.T("sales_report", language)})
	w.Write([]string{data.Company.Name})
	w.Write([]string{utils.T("period", language) + ":", data.StartDate.Format("2006-01-02"), utils.T("to", language), data.EndDate.Format("2006-01-02")})
	w.Write([]string{utils.T("generated_on", language) + ":", time.Now().In(time.Local).Format("2006-01-02 15:04")})
	w.Write([]string{})

	// Summary with localization
	w.Write([]string{utils.T("summary", language)})
	w.Write([]string{utils.T("total_revenue", language), fmt.Sprintf("%.2f", data.TotalRevenue)})
	w.Write([]string{utils.T("total_transactions", language), fmt.Sprintf("%d", data.TotalTransactions)})
	w.Write([]string{utils.T("average_order_value", language), fmt.Sprintf("%.2f", data.AverageOrderValue)})
	w.Write([]string{utils.T("total_customers", language), fmt.Sprintf("%d", data.TotalCustomers)})
	w.Write([]string{})

	// Sales by customer (top 20)
	if len(data.SalesByCustomer) > 0 {
		w.Write([]string{utils.T("sales_by_customer", language)})
		w.Write([]string{"Customer ID", "Customer Name", "Total Sales", "Transaction Count", "Average Order", "Last Order Date"})
		limit := len(data.SalesByCustomer)
		if limit > 20 {
			limit = 20
		}
		for i := 0; i < limit; i++ {
			c := data.SalesByCustomer[i]
			w.Write([]string{
				fmt.Sprintf("%d", c.CustomerID),
				c.CustomerName,
				fmt.Sprintf("%.2f", c.TotalAmount),
				fmt.Sprintf("%d", c.TransactionCount),
				fmt.Sprintf("%.2f", c.AverageOrder),
				c.LastOrderDate.Format("2006-01-02"),
			})
		}
		w.Write([]string{})
	}

	// Items Sold section
	hasItems := false
	for _, c := range data.SalesByCustomer {
		if len(c.Items) > 0 {
			hasItems = true
			break
		}
	}

	if hasItems {
		w.Write([]string{"Items Sold"})
		w.Write([]string{"Customer", "Product Code", "Product Name", "Quantity", "Unit", "Unit Price", "Total Price", "Sale Date", "Invoice"})

		for _, customer := range data.SalesByCustomer {
			if len(customer.Items) == 0 {
				continue
			}

			for _, item := range customer.Items {
				w.Write([]string{
					customer.CustomerName,
					item.ProductCode,
					item.ProductName,
					fmt.Sprintf("%.2f", item.Quantity),
					item.Unit,
					fmt.Sprintf("%.2f", item.UnitPrice),
					fmt.Sprintf("%.2f", item.TotalPrice),
					item.SaleDate.Format("2006-01-02"),
					item.InvoiceNumber,
				})
			}

			// Customer subtotal
			w.Write([]string{
				fmt.Sprintf("Subtotal (%s)", customer.CustomerName),
				"", "", "", "", "",
				fmt.Sprintf("%.2f", customer.TotalAmount),
				"", "",
			})
		}
		w.Write([]string{})
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %v", err)
	}
	return buf.Bytes(), nil
}

// ExportToPDF exports sales report to PDF bytes with localization
func (s *SalesReportExportService) ExportToPDF(data *SalesSummaryData, userID uint) ([]byte, error) {
	// Get user language preference
	language := utils.GetUserLanguageFromSettings(s.db)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	lm, tm, rm, _ := pdf.GetMargins()
	pageW, pageH := pdf.GetPageSize()
	contentW := pageW - lm - rm

	// Company settings for consistent letterhead
	settings := s.getCompanyInfo()

	// Try to render real logo at top-left
	logoW := 35.0
	logoPath := strings.TrimSpace(settings.CompanyLogo)
	logoDrawn := false
	if logoPath != "" {
		if strings.HasPrefix(logoPath, "/") {
			logoPath = "." + logoPath
		}
		if _, err := os.Stat(logoPath); err != nil {
			alt := filepath.Clean("./" + strings.TrimPrefix(settings.CompanyLogo, "/"))
			if _, err2 := os.Stat(alt); err2 == nil {
				logoPath = alt
			} else {
				logoPath = ""
			}
		}
		if logoPath != "" {
			if imgType := detectImageType(logoPath); imgType != "" {
				pdf.ImageOptions(logoPath, lm, tm, logoW, 0, false, gofpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, 0, "")
				logoDrawn = true
			}
		}
	}
	if !logoDrawn {
		pdf.SetDrawColor(220, 220, 220)
		pdf.SetFillColor(248, 249, 250)
		pdf.SetLineWidth(0.3)
		pdf.Rect(lm, tm, logoW, logoW, "FD")
		pdf.SetFont("Arial", "B", 16)
		pdf.SetTextColor(120, 120, 120)
		pdf.SetXY(lm+8, tm+19)
		pdf.CellFormat(19, 8, "</>", "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	}

	// Company info (right-aligned)
	companyName := strings.TrimSpace(settings.CompanyName)
	if companyName == "" {
		companyName = data.Company.Name
	}
	pdf.SetFont("Arial", "B", 12)
	w := pdf.GetStringWidth(companyName)
	pdf.SetXY(pageW-rm-w, tm)
	pdf.Cell(w, 6, companyName)
	pdf.SetFont("Arial", "", 9)
	addr := strings.TrimSpace(settings.CompanyAddress)
	if addr == "" {
		addr = strings.TrimSpace(data.Company.Address)
	}
	if addr != "" {
		pdf.SetXY(pageW-rm-pdf.GetStringWidth(addr), tm+8)
		pdf.Cell(pdf.GetStringWidth(addr), 4, addr)
	}

	phoneText := fmt.Sprintf("Phone: %s", strings.TrimSpace(settings.CompanyPhone))
	if strings.TrimSpace(settings.CompanyPhone) != "" {
		pdf.SetXY(pageW-rm-pdf.GetStringWidth(phoneText), tm+14)
		pdf.Cell(pdf.GetStringWidth(phoneText), 4, phoneText)
	}

	// Title
	pdf.SetY(tm + logoW + 8)
	pdf.SetFont("Arial", "B", 14)
	title := utils.T("sales_report", language)
	pdf.Cell(contentW, 8, title)
	pdf.Ln(8)

	// Period
	pdf.SetFont("Arial", "", 10)
	period := fmt.Sprintf("%s: %s %s %s",
		utils.T("period", language),
		data.StartDate.Format("02/01/2006"),
		utils.T("to", language),
		data.EndDate.Format("02/01/2006"))
	pdf.Cell(contentW, 5, period)
	pdf.Ln(5)

	// Generated at
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	generated := fmt.Sprintf("%s: %s", utils.T("generated_on", language), time.Now().Format("02/01/2006 15:04"))
	pdf.Cell(contentW, 4, generated)
	pdf.Ln(8)
	pdf.SetTextColor(0, 0, 0)

	// Divider
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.3)
	pdf.Line(lm, pdf.GetY(), pageW-rm, pdf.GetY())
	pdf.Ln(6)

	// Summary Section
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(contentW, 6, utils.T("summary", language))
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(245, 245, 245)

	// Summary table
	summaryItems := []struct {
		label string
		value string
	}{
		{utils.T("total_revenue", language), formatRupiahSimple(data.TotalRevenue)},
		{utils.T("total_transactions", language), fmt.Sprintf("%d", data.TotalTransactions)},
		{utils.T("average_order_value", language), formatRupiahSimple(data.AverageOrderValue)},
		{utils.T("total_customers", language), fmt.Sprintf("%d", data.TotalCustomers)},
	}

	for _, item := range summaryItems {
		pdf.CellFormat(contentW*0.6, 6, item.label, "1", 0, "L", true, 0, "")
		pdf.CellFormat(contentW*0.4, 6, item.value, "1", 1, "R", true, 0, "")
	}

	pdf.Ln(8)

	// Sales by Customer section
	if len(data.SalesByCustomer) > 0 {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(contentW, 6, utils.T("sales_by_customer", language))
		pdf.Ln(8)

		// Table header
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(245, 245, 245)
		pdf.CellFormat(70, 6, utils.T("customer_name", language), "1", 0, "L", true, 0, "")
		pdf.CellFormat(25, 6, utils.T("transactions", language), "1", 0, "R", true, 0, "")
		pdf.CellFormat(50, 6, utils.T("total_sales", language), "1", 0, "R", true, 0, "")
		pdf.CellFormat(45, 6, utils.T("average_order", language), "1", 1, "R", true, 0, "")

		// Customer rows
		pdf.SetFont("Arial", "", 8)
		limit := len(data.SalesByCustomer)
		if limit > 20 {
			limit = 20
		}
		for i := 0; i < limit; i++ {
			c := data.SalesByCustomer[i]

			// Check if we need a new page
			if pdf.GetY() > pageH-30 {
				pdf.AddPage()
				// Reprint header
				pdf.SetFont("Arial", "B", 8)
				pdf.SetFillColor(245, 245, 245)
				pdf.CellFormat(70, 6, utils.T("customer_name", language), "1", 0, "L", true, 0, "")
				pdf.CellFormat(25, 6, utils.T("transactions", language), "1", 0, "R", true, 0, "")
				pdf.CellFormat(50, 6, utils.T("total_sales", language), "1", 0, "R", true, 0, "")
				pdf.CellFormat(45, 6, utils.T("average_order", language), "1", 1, "R", true, 0, "")
				pdf.SetFont("Arial", "", 8)
			}

			customerName := c.CustomerName
			if len(customerName) > 35 {
				customerName = customerName[:32] + "..."
			}

			pdf.CellFormat(70, 6, customerName, "1", 0, "L", false, 0, "")
			pdf.CellFormat(25, 6, fmt.Sprintf("%d", c.TransactionCount), "1", 0, "R", false, 0, "")
			pdf.CellFormat(50, 6, formatRupiahSimple(c.TotalAmount), "1", 0, "R", false, 0, "")
			pdf.CellFormat(45, 6, formatRupiahSimple(c.AverageOrder), "1", 1, "R", false, 0, "")
		}

		pdf.Ln(6)
	}

	// Items Sold section
	hasItems := false
	for _, c := range data.SalesByCustomer {
		if len(c.Items) > 0 {
			hasItems = true
			break
		}
	}

	if hasItems {
		// Check if we need a new page for items section
		if pdf.GetY() > pageH-80 {
			pdf.AddPage()
		}

		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, utils.T("items_sold", language))
		pdf.Ln(10)

		// Loop through customers and their items
		for _, customer := range data.SalesByCustomer {
			if len(customer.Items) == 0 {
				continue
			}

			// Check if we need a new page for this customer
			if pdf.GetY() > pageH-60 {
				pdf.AddPage()
			}

			// Customer name header
			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(230, 245, 255) // Light blue
			customerName := customer.CustomerName
			if len(customerName) > 40 {
				customerName = customerName[:37] + "..."
			}
			pdf.CellFormat(140, 7, customerName, "1", 0, "L", true, 0, "")
			pdf.CellFormat(50, 7, fmt.Sprintf("%d items", len(customer.Items)), "1", 1, "R", true, 0, "")

			// Items table header
			pdf.SetFont("Arial", "B", 8)
			pdf.SetFillColor(245, 245, 245)
			pdf.CellFormat(70, 6, utils.T("product", language), "1", 0, "L", true, 0, "")
			pdf.CellFormat(20, 6, utils.T("qty", language), "1", 0, "R", true, 0, "")
			pdf.CellFormat(35, 6, utils.T("unit_price", language), "1", 0, "R", true, 0, "")
			pdf.CellFormat(35, 6, utils.T("total", language), "1", 0, "R", true, 0, "")
			pdf.CellFormat(30, 6, utils.T("date", language), "1", 1, "C", true, 0, "")

			// Items rows
			pdf.SetFont("Arial", "", 8)
			for _, item := range customer.Items {
				// Check if we need a new page
				if pdf.GetY() > pageH-25 {
					pdf.AddPage()
					// Reprint header
					pdf.SetFont("Arial", "B", 8)
					pdf.SetFillColor(245, 245, 245)
					pdf.CellFormat(70, 6, utils.T("product", language), "1", 0, "L", true, 0, "")
					pdf.CellFormat(20, 6, utils.T("qty", language), "1", 0, "R", true, 0, "")
					pdf.CellFormat(35, 6, utils.T("unit_price", language), "1", 0, "R", true, 0, "")
					pdf.CellFormat(35, 6, utils.T("total", language), "1", 0, "R", true, 0, "")
					pdf.CellFormat(30, 6, utils.T("date", language), "1", 1, "C", true, 0, "")
					pdf.SetFont("Arial", "", 8)
				}

				productName := item.ProductName
				if len(productName) > 35 {
					productName = productName[:32] + "..."
				}

				qtyStr := fmt.Sprintf("%.0f %s", item.Quantity, item.Unit)

				pdf.CellFormat(70, 6, productName, "1", 0, "L", false, 0, "")
				pdf.CellFormat(20, 6, qtyStr, "1", 0, "R", false, 0, "")
				pdf.CellFormat(35, 6, formatRupiahSimple(item.UnitPrice), "1", 0, "R", false, 0, "")
				pdf.CellFormat(35, 6, formatRupiahSimple(item.TotalPrice), "1", 0, "R", false, 0, "")
				pdf.CellFormat(30, 6, item.SaleDate.Format("02/01/2006"), "1", 1, "C", false, 0, "")
			}

			// Customer subtotal
			pdf.SetFont("Arial", "B", 9)
			pdf.SetFillColor(230, 245, 255)
			pdf.CellFormat(125, 6, fmt.Sprintf("Subtotal (%s)", customer.CustomerName), "1", 0, "R", true, 0, "")
			pdf.CellFormat(65, 6, formatRupiahSimple(customer.TotalAmount), "1", 1, "R", true, 0, "")
			pdf.Ln(4)
		}
	}

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, fmt.Errorf("failed to generate sales report PDF: %v", err)
	}
	return out.Bytes(), nil
}

