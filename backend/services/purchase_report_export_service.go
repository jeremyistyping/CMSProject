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

// PurchaseReportExportService handles export functionality for Purchase Report
type PurchaseReportExportService struct{ db *gorm.DB }

// NewPurchaseReportExportService creates a new purchase report export service
func NewPurchaseReportExportService(db *gorm.DB) *PurchaseReportExportService {
	return &PurchaseReportExportService{db: db}
}

// getCompanyInfo from settings with defaults
func (s *PurchaseReportExportService) getCompanyInfo() *models.Settings {
	if s.db == nil {
		return &models.Settings{CompanyName: "PT. Sistem Akuntansi Indonesia"}
	}
	var settings models.Settings
	if err := s.db.First(&settings).Error; err != nil {
		return &models.Settings{CompanyName: "PT. Sistem Akuntansi Indonesia"}
	}
	return &settings
}

// ExportToCSV exports purchase report to CSV bytes with localization
func (s *PurchaseReportExportService) ExportToCSV(data *PurchaseReportData, userID uint) ([]byte, error) {
	// Get user language preference
	language := utils.GetUserLanguageFromSettings(s.db)
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header with localization
	w.Write([]string{utils.T("purchase_report", language)})
	w.Write([]string{data.Company.Name})
	w.Write([]string{utils.T("period", language) + ":", data.StartDate.Format("2006-01-02"), utils.T("to", language), data.EndDate.Format("2006-01-02")})
	w.Write([]string{utils.T("generated_on", language) + ":", data.GeneratedAt.In(time.Local).Format("2006-01-02 15:04")})
	w.Write([]string{})

	// Summary with localization
	w.Write([]string{utils.T("summary", language)})
	w.Write([]string{utils.T("total_purchases", language), fmt.Sprintf("%d", data.TotalPurchases)})
	w.Write([]string{utils.T("completed_purchases", language), fmt.Sprintf("%d", data.CompletedPurchases)})
	w.Write([]string{utils.T("total_amount", language), fmt.Sprintf("%.2f", data.TotalAmount)})
	w.Write([]string{utils.T("total_paid", language), fmt.Sprintf("%.2f", data.TotalPaid)})
	w.Write([]string{utils.T("outstanding_payables", language), fmt.Sprintf("%.2f", data.OutstandingPayables)})
	w.Write([]string{})

	// Purchases by vendor (top 20) with localization
	if len(data.PurchasesByVendor) > 0 {
		w.Write([]string{utils.T("purchases_by_vendor", language)})
		headers := utils.GetCSVHeaders("purchase_report", language)
		w.Write(headers)
		limit := len(data.PurchasesByVendor)
		if limit > 20 { limit = 20 }
		for i := 0; i < limit; i++ {
			v := data.PurchasesByVendor[i]
			w.Write([]string{
				fmt.Sprintf("%d", v.VendorID), v.VendorName,
				fmt.Sprintf("%d", v.TotalPurchases),
				fmt.Sprintf("%.2f", v.TotalAmount),
				fmt.Sprintf("%.2f", v.TotalPaid),
				fmt.Sprintf("%.2f", v.Outstanding),
				v.LastPurchaseDate.Format("2006-01-02"),
				v.PaymentMethod, v.Status,
			})
		}
		w.Write([]string{})
	}

	// Items Purchased section
	hasItems := false
	for _, v := range data.PurchasesByVendor {
		if len(v.Items) > 0 {
			hasItems = true
			break
		}
	}
	
	if hasItems {
		w.Write([]string{"Items Purchased"})
		w.Write([]string{"Vendor", "Product Code", "Product Name", "Quantity", "Unit", "Unit Price", "Total Price", "Purchase Date", "Invoice"})
		
		for _, vendor := range data.PurchasesByVendor {
			if len(vendor.Items) == 0 {
				continue
			}
			
			for _, item := range vendor.Items {
				w.Write([]string{
					vendor.VendorName,
					item.ProductCode,
					item.ProductName,
					fmt.Sprintf("%.2f", item.Quantity),
					item.Unit,
					fmt.Sprintf("%.2f", item.UnitPrice),
					fmt.Sprintf("%.2f", item.TotalPrice),
					item.PurchaseDate.Format("2006-01-02"),
					item.InvoiceNumber,
				})
			}
			
			// Vendor subtotal
			w.Write([]string{
				fmt.Sprintf("Subtotal (%s)", vendor.VendorName),
				"", "", "", "", "",
				fmt.Sprintf("%.2f", vendor.TotalAmount),
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

// ExportToPDF exports purchase report to PDF bytes with localization
func (s *PurchaseReportExportService) ExportToPDF(data *PurchaseReportData, userID uint) ([]byte, error) {
	// Get user language preference
	language := utils.GetUserLanguageFromSettings(s.db)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	lm, tm, rm, _ := pdf.GetMargins()
	pageW, _ := pdf.GetPageSize()
	contentW := pageW - lm - rm

	// Company settings for consistent letterhead
	settings := s.getCompanyInfo()

	// Try to render real logo at top-left
	logoW := 35.0
	logoPath := strings.TrimSpace(settings.CompanyLogo)
	logoDrawn := false
	if logoPath != "" {
		if strings.HasPrefix(logoPath, "/") { logoPath = "." + logoPath }
		if _, err := os.Stat(logoPath); err != nil {
			alt := filepath.Clean("./" + strings.TrimPrefix(settings.CompanyLogo, "/"))
			if _, err2 := os.Stat(alt); err2 == nil { logoPath = alt } else { logoPath = "" }
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
	if companyName == "" { companyName = data.Company.Name }
	pdf.SetFont("Arial", "B", 12)
	w := pdf.GetStringWidth(companyName)
	pdf.SetXY(pageW-rm-w, tm)
	pdf.Cell(w, 6, companyName)
	pdf.SetFont("Arial", "", 9)
	addr := strings.TrimSpace(settings.CompanyAddress)
	if addr == "" { addr = strings.TrimSpace(data.Company.Address) }
	if addr != "" {
		pdf.SetXY(pageW-rm-pdf.GetStringWidth(addr), tm+8)
		pdf.Cell(0, 4, addr)
	}
	phoneVal := strings.TrimSpace(settings.CompanyPhone)
	if phoneVal == "" { phoneVal = strings.TrimSpace(data.Company.Phone) }
	if phoneVal != "" {
		phone := fmt.Sprintf("Phone: %s", phoneVal)
		pdf.SetXY(pageW-rm-pdf.GetStringWidth(phone), tm+14)
		pdf.Cell(0, 4, phone)
	}
	emailVal := strings.TrimSpace(settings.CompanyEmail)
	if emailVal == "" { emailVal = strings.TrimSpace(data.Company.Email) }
	if emailVal != "" {
		email := fmt.Sprintf("Email: %s", emailVal)
		pdf.SetXY(pageW-rm-pdf.GetStringWidth(email), tm+20)
		pdf.Cell(0, 4, email)
	}

	// Separator
	pdf.SetDrawColor(238, 238, 238)
	pdf.SetLineWidth(0.2)
	pdf.Line(lm, tm+45, pageW-rm, tm+45)

	// Title and period with localization
	pdf.SetY(tm + 55)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(contentW, 10, utils.T("purchase_report", language))
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(contentW, 6, fmt.Sprintf("%s: %s %s %s", utils.T("period", language), data.StartDate.Format("2006-01-02"), utils.T("to", language), data.EndDate.Format("2006-01-02")))
	pdf.Ln(6)
	pdf.Cell(contentW, 6, fmt.Sprintf("%s: %s", utils.T("generated_on", language), data.GeneratedAt.In(time.Local).Format("2006-01-02 15:04")))
	pdf.Ln(10)

	// Summary block with localization
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(contentW, 6, utils.T("summary", language))
	pdf.Ln(7)
	
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(245, 245, 245) // Light gray background for better visibility
	
	// Summary items dengan background
	summaryItems := []struct {
		label string
		value string
	}{
		{utils.T("total_purchases", language), fmt.Sprintf("%d", data.TotalPurchases)},
		{utils.T("completed_purchases", language), fmt.Sprintf("%d", data.CompletedPurchases)},
		{utils.T("total_amount", language), formatRupiahSimple(data.TotalAmount)},
		{utils.T("total_paid", language), formatRupiahSimple(data.TotalPaid)},
		{utils.T("outstanding_payables", language), formatRupiahSimple(data.OutstandingPayables)},
	}
	
	for _, item := range summaryItems {
		pdf.CellFormat(contentW*0.6, 6, item.label, "1", 0, "L", true, 0, "")
		pdf.CellFormat(contentW*0.4, 6, item.value, "1", 1, "R", true, 0, "")
	}
	
	pdf.Ln(8)

	// Top vendors table with localization (limit to fit one page)
	if len(data.PurchasesByVendor) > 0 {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(contentW, 6, utils.T("top_vendors", language))
		pdf.Ln(8)
		
		// Table header dengan background yang lebih jelas
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(245, 245, 245) // Sama seperti sales report
		pdf.CellFormat(60, 6, utils.T("vendor", language), "1", 0, "L", true, 0, "")
		pdf.CellFormat(25, 6, utils.T("orders", language), "1", 0, "R", true, 0, "")
		pdf.CellFormat(35, 6, utils.T("amount", language), "1", 0, "R", true, 0, "")
		pdf.CellFormat(35, 6, utils.T("paid", language), "1", 0, "R", true, 0, "")
		pdf.CellFormat(35, 6, utils.T("outstanding", language), "1", 1, "R", true, 0, "")
		
		pdf.SetFont("Arial", "", 8)
		limit := len(data.PurchasesByVendor)
		if limit > 20 { limit = 20 }
		for i := 0; i < limit; i++ {
			v := data.PurchasesByVendor[i]
			name := v.VendorName
			if len(name) > 35 { name = name[:32] + "..." }
			pdf.CellFormat(60, 6, name, "1", 0, "L", false, 0, "")
			pdf.CellFormat(25, 6, fmt.Sprintf("%d", v.TotalPurchases), "1", 0, "R", false, 0, "")
			pdf.CellFormat(35, 6, formatRupiahSimple(v.TotalAmount), "1", 0, "R", false, 0, "")
			pdf.CellFormat(35, 6, formatRupiahSimple(v.TotalPaid), "1", 0, "R", false, 0, "")
			pdf.CellFormat(35, 6, formatRupiahSimple(v.Outstanding), "1", 1, "R", false, 0, "")
		}
		pdf.Ln(6)
	}

	// Items Purchased section - grouped by vendor
	hasItems := false
	for _, v := range data.PurchasesByVendor {
		if len(v.Items) > 0 {
			hasItems = true
			break
		}
	}
	
	if hasItems {
		// Check if we need a new page for items section
		_, pageH := pdf.GetPageSize()
		_, y := pdf.GetXY()
		if y > pageH-80 {
			pdf.AddPage()
		}
		
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, utils.T("items_purchased", language))
		pdf.Ln(10)
		
		// Loop through vendors and their items
		for _, vendor := range data.PurchasesByVendor {
			if len(vendor.Items) == 0 {
				continue
			}
			
			// Check if we need a new page for this vendor
			_, y := pdf.GetXY()
			if y > pageH-60 {
				pdf.AddPage()
			}
			
			// Vendor name header - menggunakan warna yang sama dengan sales customer header
			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(255, 250, 240) // Light orange/beige - mirip customer di sales
			vendorName := vendor.VendorName
			if len(vendorName) > 40 {
				vendorName = vendorName[:37] + "..."
			}
			pdf.CellFormat(140, 7, vendorName, "1", 0, "L", true, 0, "")
			pdf.CellFormat(50, 7, fmt.Sprintf("%d items", len(vendor.Items)), "1", 1, "R", true, 0, "")
			
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
			for _, item := range vendor.Items {
				// Check if we need a new page
				_, y := pdf.GetXY()
				if y > pageH-25 {
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
				pdf.CellFormat(30, 6, item.PurchaseDate.Format("02/01/2006"), "1", 1, "C", false, 0, "")
			}
			
			// Vendor subtotal - warna yang sama dengan customer subtotal
			pdf.SetFont("Arial", "B", 9)
			pdf.SetFillColor(255, 250, 240) // Sama dengan vendor header
			pdf.CellFormat(125, 6, fmt.Sprintf("Subtotal (%s)", vendor.VendorName), "1", 0, "R", true, 0, "")
			pdf.CellFormat(65, 6, formatRupiahSimple(vendor.TotalAmount), "1", 1, "R", true, 0, "")
			pdf.Ln(4)
		}
	}

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, fmt.Errorf("failed to generate purchase report PDF: %v", err)
	}
return out.Bytes(), nil
}


// formatRupiahSimple formats a number as Indonesian Rupiah (no decimals)
func formatRupiahSimple(amount float64) string {
	// Round to integer and format with thousand separators using simple logic
	s := fmt.Sprintf("%.0f", amount)
	if s == "0" { return "Rp 0" }
	var parts []string
	for i, r := range reverseString(s) {
		if i > 0 && i%3 == 0 { parts = append(parts, ".") }
		parts = append(parts, string(r))
	}
	formatted := reverseString(strings.Join(parts, ""))
	return "Rp " + formatted
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// detectImageType function is defined in pdf_service.go and uses file headers for more accurate detection

