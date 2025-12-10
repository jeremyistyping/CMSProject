package services

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type POPDFService struct {
	poService *PurchaseOrderService
}

func NewPOPDFService(poService *PurchaseOrderService) *POPDFService {
	return &POPDFService{poService: poService}
}

// GeneratePOPDF generates a PDF for a Purchase Order
func (s *POPDFService) GeneratePOPDF(poID uint) ([]byte, error) {
	// Get PO with all relations
	po, err := s.poService.GetByID(poID)
	if err != nil {
		return nil, err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "PURCHASE ORDER", "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Company Info
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 5, "PT. UNIPRO INDONESIA", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(0, 5, "Jl. Contoh No. 123, Jakarta", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, "Telp: (021) 1234567", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// PO Info Box
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(95, 6, "PO Number: "+po.Code, "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, "Date: "+po.OrderDate.Format("02 Jan 2006"), "1", 1, "L", false, 0, "")
	
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(95, 6, "Project: "+po.Project.ProjectName, "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, "Status: "+po.Status, "1", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Vendor Info
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, "VENDOR INFORMATION", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	if po.Vendor != nil {
		pdf.CellFormat(0, 5, "Name: "+po.Vendor.Name, "", 1, "L", false, 0, "")
		if po.Vendor.Address != "" {
			pdf.CellFormat(0, 5, "Address: "+po.Vendor.Address, "", 1, "L", false, 0, "")
		}
		if po.Vendor.Phone != "" {
			pdf.CellFormat(0, 5, "Phone: "+po.Vendor.Phone, "", 1, "L", false, 0, "")
		}
	} else {
		pdf.CellFormat(0, 5, "Vendor: Not specified", "", 1, "L", false, 0, "")
	}
	pdf.Ln(5)

	// Delivery Info
	if po.DeliveryAddress != "" || po.ExpectedDeliveryDate != nil {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(0, 6, "DELIVERY INFORMATION", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		if po.DeliveryAddress != "" {
			pdf.MultiCell(0, 5, "Address: "+po.DeliveryAddress, "", "L", false)
		}
		if po.ExpectedDeliveryDate != nil {
			pdf.CellFormat(0, 5, "Expected Date: "+po.ExpectedDeliveryDate.Format("02 Jan 2006"), "", 1, "L", false, 0, "")
		}
		if po.PaymentTerms != "" {
			pdf.CellFormat(0, 5, "Payment Terms: "+po.PaymentTerms, "", 1, "L", false, 0, "")
		}
		pdf.Ln(5)
	}

	// Items Table Header
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(10, 7, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(70, 7, "Item Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Unit", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Unit Price", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Total", "1", 1, "C", true, 0, "")

	// Items
	pdf.SetFont("Arial", "", 8)
	for i, item := range po.Items {
		pdf.CellFormat(10, 6, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(70, 6, item.ItemName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", item.Quantity), "1", 0, "R", false, 0, "")
		pdf.CellFormat(20, 6, item.Unit, "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 6, formatCurrency(item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 6, formatCurrency(item.TotalPrice), "1", 1, "R", false, 0, "")
	}

	// Totals
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(120, 6, "", "", 0, "L", false, 0, "")
	pdf.CellFormat(35, 6, "Subtotal:", "1", 0, "L", false, 0, "")
	pdf.CellFormat(35, 6, formatCurrency(po.Subtotal), "1", 1, "R", false, 0, "")

	if po.TaxAmount > 0 {
		pdf.CellFormat(120, 6, "", "", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, "Tax:", "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, formatCurrency(po.TaxAmount), "1", 1, "R", false, 0, "")
	}

	if po.DiscountAmount > 0 {
		pdf.CellFormat(120, 6, "", "", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, "Discount:", "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, formatCurrency(po.DiscountAmount), "1", 1, "R", false, 0, "")
	}

	pdf.CellFormat(120, 6, "", "", 0, "L", false, 0, "")
	pdf.CellFormat(35, 6, "TOTAL:", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 6, formatCurrency(po.TotalAmount), "1", 1, "R", true, 0, "")

	// Notes
	if po.Notes != "" {
		pdf.Ln(10)
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(0, 5, "Notes:", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8)
		pdf.MultiCell(0, 5, po.Notes, "", "L", false)
	}

	// Signatures
	pdf.Ln(15)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(95, 5, "Prepared By:", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 5, "Approved By:", "", 1, "L", false, 0, "")
	pdf.Ln(15)
	pdf.CellFormat(95, 5, "_____________________", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 5, "_____________________", "", 1, "L", false, 0, "")
	pdf.CellFormat(95, 5, po.Creator.FirstName+" "+po.Creator.LastName, "", 0, "L", false, 0, "")
	if po.Approver != nil {
		pdf.CellFormat(95, 5, po.Approver.FirstName+" "+po.Approver.LastName, "", 1, "L", false, 0, "")
	} else {
		pdf.CellFormat(95, 5, "", "", 1, "L", false, 0, "")
	}

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 10, fmt.Sprintf("Generated on %s", time.Now().Format("02 Jan 2006 15:04")), "", 0, "C", false, 0, "")

	// Get PDF bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GenerateGRPDF generates a PDF for Goods Receipt
func (s *POPDFService) GenerateGRPDF(poID uint) ([]byte, error) {
	// Get PO with all relations
	po, err := s.poService.GetByID(poID)
	if err != nil {
		return nil, err
	}

	// Get all goods receipts for this PO
	grs, err := s.poService.GetGoodsReceiptsByPOID(poID)
	if err != nil {
		return nil, err
	}

	if len(grs) == 0 {
		return nil, fmt.Errorf("no goods receipts found for this PO")
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "GOODS RECEIPT REPORT", "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Company Info
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 5, "PT. UNIPRO INDONESIA", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(0, 5, "Jl. Contoh No. 123, Jakarta", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// PO Reference Box
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(0, 7, "PURCHASE ORDER INFORMATION", "1", 1, "L", true, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(50, 6, "PO Number:", "1", 0, "L", false, 0, "")
	pdf.CellFormat(140, 6, po.Code, "1", 1, "L", false, 0, "")
	pdf.CellFormat(50, 6, "Project:", "1", 0, "L", false, 0, "")
	pdf.CellFormat(140, 6, po.Project.ProjectName, "1", 1, "L", false, 0, "")
	if po.Vendor != nil {
		pdf.CellFormat(50, 6, "Vendor:", "1", 0, "L", false, 0, "")
		pdf.CellFormat(140, 6, po.Vendor.Name, "1", 1, "L", false, 0, "")
	}
	pdf.Ln(5)

	// Summary Section
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(230, 240, 255)
	pdf.CellFormat(0, 7, "RECEIPT SUMMARY", "1", 1, "L", true, 0, "")
	
	// Create map of PO items for easy lookup
	poItemMap := make(map[uint]*struct {
		Name            string
		Unit            string
		OrderedQty      float64
		TotalReceived   float64
		TotalAccepted   float64
		TotalRejected   float64
		ReceiptCount    int
	})
	
	for _, item := range po.Items {
		poItemMap[item.ID] = &struct {
			Name            string
			Unit            string
			OrderedQty      float64
			TotalReceived   float64
			TotalAccepted   float64
			TotalRejected   float64
			ReceiptCount    int
		}{
			Name:       item.ItemName,
			Unit:       item.Unit,
			OrderedQty: item.Quantity,
		}
	}
	
	// Aggregate receipt data
	for _, gr := range grs {
		for _, grItem := range gr.Items {
			if summary, exists := poItemMap[grItem.POItemID]; exists {
				summary.TotalReceived += grItem.ReceivedQuantity
				summary.TotalAccepted += grItem.AcceptedQuantity
				summary.TotalRejected += grItem.RejectedQuantity
				summary.ReceiptCount++
			}
		}
	}
	
	// Summary Table Header
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(10, 7, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(70, 7, "Item Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Ordered", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Received", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Accepted", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Rejected", "1", 0, "C", true, 0, "")
	pdf.CellFormat(15, 7, "Unit", "1", 0, "C", true, 0, "")
	pdf.CellFormat(15, 7, "Status", "1", 1, "C", true, 0, "")
	
	// Summary Table Data
	pdf.SetFont("Arial", "", 8)
	itemNo := 1
	for _, item := range po.Items {
		if summary, exists := poItemMap[item.ID]; exists {
			// Determine status
			status := "Pending"
			statusColor := false
			if summary.TotalAccepted >= summary.OrderedQty {
				status = "Complete"
				pdf.SetFillColor(200, 255, 200) // Light green
				statusColor = true
			} else if summary.TotalReceived > 0 {
				status = "Partial"
				pdf.SetFillColor(255, 255, 200) // Light yellow
				statusColor = true
			}
			
			pdf.CellFormat(10, 6, fmt.Sprintf("%d", itemNo), "1", 0, "C", false, 0, "")
			pdf.CellFormat(70, 6, summary.Name, "1", 0, "L", false, 0, "")
			pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", summary.OrderedQty), "1", 0, "R", false, 0, "")
			pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", summary.TotalReceived), "1", 0, "R", false, 0, "")
			pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", summary.TotalAccepted), "1", 0, "R", false, 0, "")
			pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", summary.TotalRejected), "1", 0, "R", false, 0, "")
			pdf.CellFormat(15, 6, summary.Unit, "1", 0, "C", false, 0, "")
			pdf.CellFormat(15, 6, status, "1", 1, "C", statusColor, 0, "")
			itemNo++
		}
	}
	pdf.Ln(8)

	// Detailed Receipts Section
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(230, 240, 255)
	pdf.CellFormat(0, 7, "DETAILED RECEIPT HISTORY", "1", 1, "L", true, 0, "")
	pdf.Ln(3)

	// Loop through all goods receipts
	for grIdx, gr := range grs {
		if grIdx > 0 {
			pdf.Ln(5)
		}

		// GR Info Box
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(245, 245, 245)
		pdf.CellFormat(0, 6, fmt.Sprintf("Receipt #%d - %s", grIdx+1, gr.Code), "1", 1, "L", true, 0, "")
		
		pdf.SetFont("Arial", "", 8)
		pdf.CellFormat(60, 5, "Date: "+gr.ReceiptDate.Format("02 Jan 2006"), "LR", 0, "L", false, 0, "")
		pdf.CellFormat(65, 5, "Received By: "+gr.Receiver.FirstName+" "+gr.Receiver.LastName, "LR", 0, "L", false, 0, "")
		pdf.CellFormat(65, 5, "Status: "+gr.Status, "LR", 1, "L", false, 0, "")
		pdf.CellFormat(0, 0, "", "LRB", 1, "L", false, 0, "")

		// Items for this receipt
		pdf.SetFont("Arial", "", 7)
		for i, grItem := range gr.Items {
			// Find item name from PO
			itemName := "Unknown Item"
			itemUnit := ""
			for _, poItem := range po.Items {
				if poItem.ID == grItem.POItemID {
					itemName = poItem.ItemName
					itemUnit = poItem.Unit
					break
				}
			}
			
			pdf.CellFormat(10, 5, fmt.Sprintf("%d.", i+1), "LR", 0, "L", false, 0, "")
			pdf.CellFormat(80, 5, itemName, "R", 0, "L", false, 0, "")
			pdf.CellFormat(30, 5, fmt.Sprintf("Received: %.2f %s", grItem.ReceivedQuantity, itemUnit), "R", 0, "L", false, 0, "")
			pdf.CellFormat(30, 5, fmt.Sprintf("Accepted: %.2f", grItem.AcceptedQuantity), "R", 0, "L", false, 0, "")
			pdf.CellFormat(40, 5, fmt.Sprintf("Rejected: %.2f", grItem.RejectedQuantity), "R", 1, "L", false, 0, "")
			
			if grItem.RejectionReason != "" {
				pdf.SetFont("Arial", "I", 7)
				pdf.CellFormat(10, 4, "", "LR", 0, "L", false, 0, "")
				pdf.MultiCell(180, 4, "Reason: "+grItem.RejectionReason, "R", "L", false)
				pdf.SetFont("Arial", "", 7)
			}
		}
		pdf.CellFormat(0, 0, "", "LRB", 1, "L", false, 0, "")

		// Notes
		if gr.Notes != "" {
			pdf.SetFont("Arial", "I", 7)
			pdf.CellFormat(20, 4, "Notes:", "LR", 0, "L", false, 0, "")
			pdf.MultiCell(170, 4, gr.Notes, "R", "L", false)
			pdf.CellFormat(0, 0, "", "LRB", 1, "L", false, 0, "")
		}
	}

	// Signature
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(95, 5, "Warehouse Manager", "", 0, "C", false, 0, "")
	pdf.CellFormat(95, 5, "Verified By", "", 1, "C", false, 0, "")
	pdf.Ln(15)
	pdf.CellFormat(95, 5, "_____________________", "", 0, "C", false, 0, "")
	pdf.CellFormat(95, 5, "_____________________", "", 1, "C", false, 0, "")

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 10, fmt.Sprintf("Generated on %s | Total Receipts: %d", time.Now().Format("02 Jan 2006 15:04"), len(grs)), "", 0, "C", false, 0, "")

	// Get PDF bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Helper function to format currency with thousand separator
func formatCurrency(amount float64) string {
	// Format number with 2 decimal places
	str := fmt.Sprintf("%.2f", amount)
	
	// Split into integer and decimal parts
	parts := []rune(str)
	var intPart, decPart string
	dotIndex := -1
	for i, r := range parts {
		if r == '.' {
			dotIndex = i
			break
		}
	}
	
	if dotIndex >= 0 {
		intPart = string(parts[:dotIndex])
		decPart = string(parts[dotIndex:])
	} else {
		intPart = str
		decPart = ".00"
	}
	
	// Add thousand separators to integer part
	var result []rune
	for i, r := range []rune(intPart) {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, r)
	}
	
	return "Rp " + string(result) + decPart
}
/*  */