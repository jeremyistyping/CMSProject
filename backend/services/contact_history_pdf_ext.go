package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	
	"github.com/jung-kurt/gofpdf"
)

// GenerateCustomerHistoryPDF generates PDF for customer transaction history
func (p *PDFService) GenerateCustomerHistoryPDF(historyData interface{}) ([]byte, error) {
	// Convert history data to map for easier access
	var dataMap map[string]interface{}
	b, _ := json.Marshal(historyData)
	_ = json.Unmarshal(b, &dataMap)

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	lm, tm, rm, _ := pdf.GetMargins()
	pageW, _ := pdf.GetPageSize()
	contentW := pageW - lm - rm

	// Company info
	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}

	// Logo and company header
	logoX, logoY, logoSize := lm, tm, 35.0
	logoAdded := false
	if strings.TrimSpace(companyInfo.CompanyLogo) != "" {
		logoPath := companyInfo.CompanyLogo
		if strings.HasPrefix(logoPath, "/") { logoPath = "." + logoPath }
		if _, err := os.Stat(logoPath); err == nil {
			if imgType := detectImageType(logoPath); imgType != "" {
				pdf.ImageOptions(logoPath, logoX, logoY, logoSize, 0, false, gofpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, 0, "")
				logoAdded = true
			}
		}
	}
	if !logoAdded {
		pdf.SetDrawColor(220,220,220)
		pdf.SetFillColor(248,249,250)
		pdf.SetLineWidth(0.3)
		pdf.Rect(logoX, logoY, logoSize, logoSize, "FD")
		pdf.SetFont("Arial","B",16)
		pdf.SetTextColor(120,120,120)
		pdf.SetXY(logoX+8, logoY+19)
		pdf.CellFormat(19,8,"</>","",0,"C",false,0,"")
		pdf.SetTextColor(0,0,0)
	}

	// Company name and address on right
	companyInfoX := pageW - rm
	pdf.SetFont("Arial","B",12)
	nameW := pdf.GetStringWidth(companyInfo.CompanyName)
	pdf.SetXY(companyInfoX-nameW, tm)
	pdf.Cell(nameW, 6, companyInfo.CompanyName)

	pdf.SetFont("Arial","",9)
	addrW := pdf.GetStringWidth(companyInfo.CompanyAddress)
	pdf.SetXY(companyInfoX-addrW, tm+8)
	pdf.Cell(addrW, 4, companyInfo.CompanyAddress)

	phoneText := fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone)
	phoneW := pdf.GetStringWidth(phoneText)
	pdf.SetXY(companyInfoX-phoneW, tm+14)
	pdf.Cell(phoneW, 4, phoneText)

	// Divider line
	pdf.SetDrawColor(238,238,238)
	pdf.SetLineWidth(0.2)
	pdf.Line(lm, tm+40, pageW-rm, tm+40)

	// Title
	pdf.SetY(tm + 45)
	pdf.SetFont("Arial","B",18)
	pdf.SetTextColor(51,51,51)
	pdf.Cell(contentW,8,"CUSTOMER TRANSACTION HISTORY")
	pdf.SetTextColor(0,0,0)
	pdf.Ln(10)

	// Customer info and period
	customerMap, _ := dataMap["customer"].(map[string]interface{})
	startDate := getString(dataMap, "start_date")
	endDate := getString(dataMap, "end_date")
	customerName := "Unknown Customer"
	customerCode := ""
	if customerMap != nil {
		customerName = getString(customerMap, "name")
		customerCode = getString(customerMap, "code")
	}

	pdf.SetFont("Arial","B",9)
	pdf.SetX(lm)
	pdf.Cell(30,5,"Customer:")
	pdf.SetFont("Arial","",9)
	pdf.SetTextColor(102,102,102)
	pdf.Cell(65,5,fmt.Sprintf("%s (%s)", customerName, customerCode))

	pdf.SetFont("Arial","B",9)
	pdf.SetTextColor(0,0,0)
	rightX := lm + contentW - 70
	pdf.SetX(rightX)
	pdf.Cell(22,5,"Period:")
	pdf.SetFont("Arial","",9)
	pdf.SetTextColor(102,102,102)
	pdf.Cell(48,5,fmt.Sprintf("%s to %s", startDate, endDate))
	pdf.Ln(8)

	// Summary box
	summaryMap, _ := dataMap["summary"].(map[string]interface{})
	if summaryMap != nil {
		pdf.SetFillColor(240, 248, 255)
		pdf.SetDrawColor(150, 180, 220)
		pdf.Rect(lm, pdf.GetY(), contentW, 25, "FD")
		
		pdf.SetFont("Arial","B",9)
		pdf.SetY(pdf.GetY()+3)
		pdf.Cell(contentW/4,5,"Total Transactions")
		pdf.Cell(contentW/4,5,"Total Amount")
		pdf.Cell(contentW/4,5,"Total Paid")
		pdf.Cell(contentW/4,5,"Outstanding")
		pdf.Ln(6)
		
		pdf.SetFont("Arial","",11)
		totalTx := int(getNumFrom(summaryMap["total_transactions"]))
		totalAmount := getNumFrom(summaryMap["total_amount"])
		totalPaid := getNumFrom(summaryMap["total_paid"])
		totalOutstanding := getNumFrom(summaryMap["total_outstanding"])
		
		pdf.Cell(contentW/4,5,fmt.Sprintf("%d", totalTx))
		pdf.Cell(contentW/4,5,p.formatRupiah(totalAmount))
		pdf.Cell(contentW/4,5,p.formatRupiah(totalPaid))
		pdf.Cell(contentW/4,5,p.formatRupiah(totalOutstanding))
		pdf.Ln(15)
	}

	// Transaction table headers
	pdf.SetFont("Arial", "B", 7)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(18, 7, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(18, 7, "Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Code", "1", 0, "C", true, 0, "")
	pdf.CellFormat(52, 7, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(22, 7, "Amount", "1", 0, "R", true, 0, "")
	pdf.CellFormat(22, 7, "Paid", "1", 0, "R", true, 0, "")
	pdf.CellFormat(23, 7, "Outstanding", "1", 0, "R", true, 0, "")
	pdf.Ln(7)

	// Transaction data
	pdf.SetFont("Arial", "", 6)
	if transactions, exists := dataMap["transactions"]; exists {
		if txSlice, ok := transactions.([]interface{}); ok {
			for _, tx := range txSlice {
				if txMap, ok := tx.(map[string]interface{}); ok {
					date := getString(txMap, "date")
					if len(date) > 10 { date = date[:10] }
					
					pdf.CellFormat(18, 6, date, "1", 0, "L", false, 0, "")
					pdf.CellFormat(18, 6, getString(txMap, "transaction_type"), "1", 0, "L", false, 0, "")
					pdf.CellFormat(25, 6, getString(txMap, "transaction_code"), "1", 0, "L", false, 0, "")
					desc := getString(txMap, "description")
					if len(desc) > 45 { desc = desc[:45] + "..." }
					pdf.CellFormat(52, 6, desc, "1", 0, "L", false, 0, "")
					pdf.CellFormat(22, 6, p.formatRupiah(getNumFrom(txMap["amount"])), "1", 0, "R", false, 0, "")
					pdf.CellFormat(22, 6, p.formatRupiah(getNumFrom(txMap["paid_amount"])), "1", 0, "R", false, 0, "")
					pdf.CellFormat(23, 6, p.formatRupiah(getNumFrom(txMap["outstanding"])), "1", 0, "R", false, 0, "")
					pdf.Ln(6)
				}
			}
		}
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(contentW, 4, fmt.Sprintf("Report generated on %s", time.Now().Format("02/01/2006 15:04")))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate customer history PDF: %v", err)
	}
	return buf.Bytes(), nil
}

// GenerateVendorHistoryPDF generates PDF for vendor transaction history
func (p *PDFService) GenerateVendorHistoryPDF(historyData interface{}) ([]byte, error) {
	// Convert history data to map for easier access
	var dataMap map[string]interface{}
	b, _ := json.Marshal(historyData)
	_ = json.Unmarshal(b, &dataMap)

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	lm, tm, rm, _ := pdf.GetMargins()
	pageW, _ := pdf.GetPageSize()
	contentW := pageW - lm - rm

	// Company info
	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}

	// Logo and company header (similar to General Ledger)
	logoX, logoY, logoSize := lm, tm, 35.0
	logoAdded := false
	if strings.TrimSpace(companyInfo.CompanyLogo) != "" {
		logoPath := companyInfo.CompanyLogo
		if strings.HasPrefix(logoPath, "/") { logoPath = "." + logoPath }
		if _, err := os.Stat(logoPath); err == nil {
			if imgType := detectImageType(logoPath); imgType != "" {
				pdf.ImageOptions(logoPath, logoX, logoY, logoSize, 0, false, gofpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, 0, "")
				logoAdded = true
			}
		}
	}
	if !logoAdded {
		pdf.SetDrawColor(220,220,220)
		pdf.SetFillColor(248,249,250)
		pdf.SetLineWidth(0.3)
		pdf.Rect(logoX, logoY, logoSize, logoSize, "FD")
		pdf.SetFont("Arial","B",16)
		pdf.SetTextColor(120,120,120)
		pdf.SetXY(logoX+8, logoY+19)
		pdf.CellFormat(19,8,"</>","",0,"C",false,0,"")
		pdf.SetTextColor(0,0,0)
	}

	// Company name and address on right
	companyInfoX := pageW - rm
	pdf.SetFont("Arial","B",12)
	nameW := pdf.GetStringWidth(companyInfo.CompanyName)
	pdf.SetXY(companyInfoX-nameW, tm)
	pdf.Cell(nameW, 6, companyInfo.CompanyName)

	pdf.SetFont("Arial","",9)
	addrW := pdf.GetStringWidth(companyInfo.CompanyAddress)
	pdf.SetXY(companyInfoX-addrW, tm+8)
	pdf.Cell(addrW, 4, companyInfo.CompanyAddress)

	phoneText := fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone)
	phoneW := pdf.GetStringWidth(phoneText)
	pdf.SetXY(companyInfoX-phoneW, tm+14)
	pdf.Cell(phoneW, 4, phoneText)

	// Divider line
	pdf.SetDrawColor(238,238,238)
	pdf.SetLineWidth(0.2)
	pdf.Line(lm, tm+40, pageW-rm, tm+40)

	// Title
	pdf.SetY(tm + 45)
	pdf.SetFont("Arial","B",18)
	pdf.SetTextColor(51,51,51)
	pdf.Cell(contentW,8,"VENDOR TRANSACTION HISTORY")
	pdf.SetTextColor(0,0,0)
	pdf.Ln(10)

	// Vendor info and period
	vendorMap, _ := dataMap["vendor"].(map[string]interface{})
	startDate := getString(dataMap, "start_date")
	endDate := getString(dataMap, "end_date")
	vendorName := "Unknown Vendor"
	vendorCode := ""
	if vendorMap != nil {
		vendorName = getString(vendorMap, "name")
		vendorCode = getString(vendorMap, "code")
	}

	pdf.SetFont("Arial","B",9)
	pdf.SetX(lm)
	pdf.Cell(30,5,"Vendor:")
	pdf.SetFont("Arial","",9)
	pdf.SetTextColor(102,102,102)
	pdf.Cell(65,5,fmt.Sprintf("%s (%s)", vendorName, vendorCode))

	pdf.SetFont("Arial","B",9)
	pdf.SetTextColor(0,0,0)
	rightX := lm + contentW - 70
	pdf.SetX(rightX)
	pdf.Cell(22,5,"Period:")
	pdf.SetFont("Arial","",9)
	pdf.SetTextColor(102,102,102)
	pdf.Cell(48,5,fmt.Sprintf("%s to %s", startDate, endDate))
	pdf.Ln(8)

	// Summary box
	summaryMap, _ := dataMap["summary"].(map[string]interface{})
	if summaryMap != nil {
		pdf.SetFillColor(255, 248, 240)
		pdf.SetDrawColor(220, 180, 150)
		pdf.Rect(lm, pdf.GetY(), contentW, 25, "FD")
		
		pdf.SetFont("Arial","B",9)
		pdf.SetY(pdf.GetY()+3)
		pdf.Cell(contentW/4,5,"Total Transactions")
		pdf.Cell(contentW/4,5,"Total Amount")
		pdf.Cell(contentW/4,5,"Total Paid")
		pdf.Cell(contentW/4,5,"Outstanding")
		pdf.Ln(6)
		
		pdf.SetFont("Arial","",11)
		totalTx := int(getNumFrom(summaryMap["total_transactions"]))
		totalAmount := getNumFrom(summaryMap["total_amount"])
		totalPaid := getNumFrom(summaryMap["total_paid"])
		totalOutstanding := getNumFrom(summaryMap["total_outstanding"])
		
		pdf.Cell(contentW/4,5,fmt.Sprintf("%d", totalTx))
		pdf.Cell(contentW/4,5,p.formatRupiah(totalAmount))
		pdf.Cell(contentW/4,5,p.formatRupiah(totalPaid))
		pdf.Cell(contentW/4,5,p.formatRupiah(totalOutstanding))
		pdf.Ln(15)
	}

	// Transaction table headers
	pdf.SetFont("Arial", "B", 7)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(18, 7, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(18, 7, "Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Code", "1", 0, "C", true, 0, "")
	pdf.CellFormat(52, 7, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(22, 7, "Amount", "1", 0, "R", true, 0, "")
	pdf.CellFormat(22, 7, "Paid", "1", 0, "R", true, 0, "")
	pdf.CellFormat(23, 7, "Outstanding", "1", 0, "R", true, 0, "")
	pdf.Ln(7)

	// Transaction data
	pdf.SetFont("Arial", "", 6)
	if transactions, exists := dataMap["transactions"]; exists {
		if txSlice, ok := transactions.([]interface{}); ok {
			for _, tx := range txSlice {
				if txMap, ok := tx.(map[string]interface{}); ok {
					date := getString(txMap, "date")
					if len(date) > 10 { date = date[:10] }
					
					pdf.CellFormat(18, 6, date, "1", 0, "L", false, 0, "")
					pdf.CellFormat(18, 6, getString(txMap, "transaction_type"), "1", 0, "L", false, 0, "")
					pdf.CellFormat(25, 6, getString(txMap, "transaction_code"), "1", 0, "L", false, 0, "")
					desc := getString(txMap, "description")
					if len(desc) > 45 { desc = desc[:45] + "..." }
					pdf.CellFormat(52, 6, desc, "1", 0, "L", false, 0, "")
					pdf.CellFormat(22, 6, p.formatRupiah(getNumFrom(txMap["amount"])), "1", 0, "R", false, 0, "")
					pdf.CellFormat(22, 6, p.formatRupiah(getNumFrom(txMap["paid_amount"])), "1", 0, "R", false, 0, "")
					pdf.CellFormat(23, 6, p.formatRupiah(getNumFrom(txMap["outstanding"])), "1", 0, "R", false, 0, "")
					pdf.Ln(6)
				}
			}
		}
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(contentW, 4, fmt.Sprintf("Report generated on %s", time.Now().Format("02/01/2006 15:04")))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate vendor history PDF: %v", err)
	}
	return buf.Bytes(), nil
}

// getString helper to get string from map with fallback
func getString(m map[string]interface{}, key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
