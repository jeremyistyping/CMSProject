package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"
)

// GenerateCustomerHistoryCSV generates CSV for customer transaction history
func (p *PDFService) GenerateCustomerHistoryCSV(historyData interface{}) ([]byte, error) {
	// Convert history data to map for easier access
	var dataMap map[string]interface{}
	b, _ := json.Marshal(historyData)
	_ = json.Unmarshal(b, &dataMap)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Get customer info
	customerMap, _ := dataMap["customer"].(map[string]interface{})
	startDate := getString(dataMap, "start_date")
	endDate := getString(dataMap, "end_date")
	customerName := "Unknown Customer"
	customerCode := ""
	if customerMap != nil {
		customerName = getString(customerMap, "name")
		customerCode = getString(customerMap, "code")
	}

	// Header
	w.Write([]string{"CUSTOMER TRANSACTION HISTORY"})
	w.Write([]string{"Customer:", fmt.Sprintf("%s (%s)", customerName, customerCode)})
	w.Write([]string{"Period:", fmt.Sprintf("%s to %s", startDate, endDate)})
	w.Write([]string{"Generated:", time.Now().Format("2006-01-02 15:04")})
	w.Write([]string{})

	// Summary
	summaryMap, _ := dataMap["summary"].(map[string]interface{})
	if summaryMap != nil {
		w.Write([]string{"SUMMARY"})
		totalTx := int(getNumFrom(summaryMap["total_transactions"]))
		totalAmount := getNumFrom(summaryMap["total_amount"])
		totalPaid := getNumFrom(summaryMap["total_paid"])
		totalOutstanding := getNumFrom(summaryMap["total_outstanding"])

		w.Write([]string{"Total Transactions:", fmt.Sprintf("%d", totalTx)})
		w.Write([]string{"Total Amount:", fmt.Sprintf("%.2f", totalAmount)})
		w.Write([]string{"Total Paid:", fmt.Sprintf("%.2f", totalPaid)})
		w.Write([]string{"Outstanding:", fmt.Sprintf("%.2f", totalOutstanding)})
		w.Write([]string{})
	}

	// Transaction table headers
	w.Write([]string{"TRANSACTIONS"})
	w.Write([]string{"Date", "Type", "Code", "Description", "Reference", "Amount", "Paid Amount", "Outstanding", "Status", "Due Date"})

	// Transaction data
	if transactions, exists := dataMap["transactions"]; exists {
		if txSlice, ok := transactions.([]interface{}); ok {
			for _, tx := range txSlice {
				if txMap, ok := tx.(map[string]interface{}); ok {
					date := getString(txMap, "date")
					if len(date) > 10 {
						date = date[:10]
					}

					dueDate := ""
					if dueDateVal, exists := txMap["due_date"]; exists && dueDateVal != nil {
						if dueDateStr, ok := dueDateVal.(string); ok && dueDateStr != "" {
							if len(dueDateStr) > 10 {
								dueDate = dueDateStr[:10]
							} else {
								dueDate = dueDateStr
							}
						}
					}

					w.Write([]string{
						date,
						getString(txMap, "transaction_type"),
						getString(txMap, "transaction_code"),
						getString(txMap, "description"),
						getString(txMap, "reference"),
						fmt.Sprintf("%.2f", getNumFrom(txMap["amount"])),
						fmt.Sprintf("%.2f", getNumFrom(txMap["paid_amount"])),
						fmt.Sprintf("%.2f", getNumFrom(txMap["outstanding"])),
						getString(txMap, "status"),
						dueDate,
					})
				}
			}
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("failed to write customer history CSV: %v", err)
	}
	return buf.Bytes(), nil
}

// GenerateVendorHistoryCSV generates CSV for vendor transaction history
func (p *PDFService) GenerateVendorHistoryCSV(historyData interface{}) ([]byte, error) {
	// Convert history data to map for easier access
	var dataMap map[string]interface{}
	b, _ := json.Marshal(historyData)
	_ = json.Unmarshal(b, &dataMap)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Get vendor info
	vendorMap, _ := dataMap["vendor"].(map[string]interface{})
	startDate := getString(dataMap, "start_date")
	endDate := getString(dataMap, "end_date")
	vendorName := "Unknown Vendor"
	vendorCode := ""
	if vendorMap != nil {
		vendorName = getString(vendorMap, "name")
		vendorCode = getString(vendorMap, "code")
	}

	// Header
	w.Write([]string{"VENDOR TRANSACTION HISTORY"})
	w.Write([]string{"Vendor:", fmt.Sprintf("%s (%s)", vendorName, vendorCode)})
	w.Write([]string{"Period:", fmt.Sprintf("%s to %s", startDate, endDate)})
	w.Write([]string{"Generated:", time.Now().Format("2006-01-02 15:04")})
	w.Write([]string{})

	// Summary
	summaryMap, _ := dataMap["summary"].(map[string]interface{})
	if summaryMap != nil {
		w.Write([]string{"SUMMARY"})
		totalTx := int(getNumFrom(summaryMap["total_transactions"]))
		totalAmount := getNumFrom(summaryMap["total_amount"])
		totalPaid := getNumFrom(summaryMap["total_paid"])
		totalOutstanding := getNumFrom(summaryMap["total_outstanding"])

		w.Write([]string{"Total Transactions:", fmt.Sprintf("%d", totalTx)})
		w.Write([]string{"Total Amount:", fmt.Sprintf("%.2f", totalAmount)})
		w.Write([]string{"Total Paid:", fmt.Sprintf("%.2f", totalPaid)})
		w.Write([]string{"Outstanding:", fmt.Sprintf("%.2f", totalOutstanding)})
		w.Write([]string{})
	}

	// Transaction table headers
	w.Write([]string{"TRANSACTIONS"})
	w.Write([]string{"Date", "Type", "Code", "Description", "Reference", "Amount", "Paid Amount", "Outstanding", "Status", "Due Date"})

	// Transaction data
	if transactions, exists := dataMap["transactions"]; exists {
		if txSlice, ok := transactions.([]interface{}); ok {
			for _, tx := range txSlice {
				if txMap, ok := tx.(map[string]interface{}); ok {
					date := getString(txMap, "date")
					if len(date) > 10 {
						date = date[:10]
					}

					dueDate := ""
					if dueDateVal, exists := txMap["due_date"]; exists && dueDateVal != nil {
						if dueDateStr, ok := dueDateVal.(string); ok && dueDateStr != "" {
							if len(dueDateStr) > 10 {
								dueDate = dueDateStr[:10]
							} else {
								dueDate = dueDateStr
							}
						}
					}

					w.Write([]string{
						date,
						getString(txMap, "transaction_type"),
						getString(txMap, "transaction_code"),
						getString(txMap, "description"),
						getString(txMap, "reference"),
						fmt.Sprintf("%.2f", getNumFrom(txMap["amount"])),
						fmt.Sprintf("%.2f", getNumFrom(txMap["paid_amount"])),
						fmt.Sprintf("%.2f", getNumFrom(txMap["outstanding"])),
						getString(txMap, "status"),
						dueDate,
					})
				}
			}
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("failed to write vendor history CSV: %v", err)
	}
	return buf.Bytes(), nil
}
