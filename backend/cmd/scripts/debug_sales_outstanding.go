package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
)

// Debug script for sales outstanding update issue
func main() {
	fmt.Println("==========================================")
	fmt.Println("🔍 DEBUGGING SALES OUTSTANDING UPDATE")
	fmt.Println("==========================================")

	db := database.ConnectDB()

	// Step 1: Check sales with payments but outstanding not updated
	fmt.Println("\n📊 ANALYSIS: Sales vs Payments")
	
	var problematicSales []struct {
		SaleID            uint    `gorm:"column:sale_id"`
		SaleCode          string  `gorm:"column:sale_code"`
		CustomerName      string  `gorm:"column:customer_name"`
		TotalAmount       float64 `gorm:"column:total_amount"`
		OutstandingAmount float64 `gorm:"column:outstanding_amount"`
		TotalPayments     float64 `gorm:"column:total_payments"`
		PaymentCount      int     `gorm:"column:payment_count"`
		ExpectedOutstanding float64 `gorm:"column:expected_outstanding"`
	}

	query := `
		SELECT 
			s.id as sale_id,
			s.code as sale_code,
			c.name as customer_name,
			s.total_amount,
			s.outstanding_amount,
			COALESCE(SUM(pa.allocated_amount), 0) as total_payments,
			COUNT(pa.id) as payment_count,
			s.total_amount - COALESCE(SUM(pa.allocated_amount), 0) as expected_outstanding
		FROM sales s
		LEFT JOIN contacts c ON s.customer_id = c.id
		LEFT JOIN payment_allocations pa ON pa.invoice_id = s.id
		WHERE s.status IN ('INVOICED', 'PAID')
		GROUP BY s.id, s.code, c.name, s.total_amount, s.outstanding_amount
		HAVING s.outstanding_amount != (s.total_amount - COALESCE(SUM(pa.allocated_amount), 0))
		ORDER BY s.id
	`

	db.Raw(query).Scan(&problematicSales)

	if len(problematicSales) == 0 {
		fmt.Println("✅ No sales outstanding issues found")
		return
	}

	fmt.Printf("❌ Found %d sales with outstanding amount issues:\n\n", len(problematicSales))

	for i, sale := range problematicSales {
		fmt.Printf("  %d. Sale: %s (%s)\n", i+1, sale.SaleCode, sale.CustomerName)
		fmt.Printf("     Total: Rp %.2f\n", sale.TotalAmount)
		fmt.Printf("     Current Outstanding: Rp %.2f\n", sale.OutstandingAmount)
		fmt.Printf("     Total Payments: Rp %.2f (%d payments)\n", sale.TotalPayments, sale.PaymentCount)
		fmt.Printf("     Expected Outstanding: Rp %.2f\n", sale.ExpectedOutstanding)
		fmt.Printf("     Issue: Outstanding should be %.2f but shows %.2f\n\n", 
			sale.ExpectedOutstanding, sale.OutstandingAmount)
	}

	// Step 2: Check payment allocations for recent payments
	fmt.Println("📋 PAYMENT ALLOCATIONS ANALYSIS:")
	
	var recentPayments []struct {
		PaymentID     uint    `gorm:"column:payment_id"`
		PaymentCode   string  `gorm:"column:payment_code"`
		PaymentAmount float64 `gorm:"column:payment_amount"`
		ContactName   string  `gorm:"column:contact_name"`
		AllocationID  *uint   `gorm:"column:allocation_id"`
		InvoiceID     *uint   `gorm:"column:invoice_id"`
		AllocatedAmount *float64 `gorm:"column:allocated_amount"`
		SaleCode      *string `gorm:"column:sale_code"`
	}

	paymentQuery := `
		SELECT 
			p.id as payment_id,
			p.code as payment_code,
			p.amount as payment_amount,
			c.name as contact_name,
			pa.id as allocation_id,
			pa.invoice_id,
			pa.allocated_amount,
			s.code as sale_code
		FROM payments p
		LEFT JOIN contacts c ON p.contact_id = c.id
		LEFT JOIN payment_allocations pa ON pa.payment_id = p.id
		LEFT JOIN sales s ON pa.invoice_id = s.id
		WHERE p.created_at >= NOW() - INTERVAL '30 days'
		ORDER BY p.created_at DESC
		LIMIT 10
	`

	db.Raw(paymentQuery).Scan(&recentPayments)

	for i, payment := range recentPayments {
		fmt.Printf("  %d. Payment: %s (Rp %.2f) - %s\n", i+1, payment.PaymentCode, payment.PaymentAmount, payment.ContactName)
		if payment.AllocationID != nil {
			fmt.Printf("     ✅ Allocated: Rp %.2f to Sale %s\n", *payment.AllocatedAmount, *payment.SaleCode)
		} else {
			fmt.Printf("     ❌ NO ALLOCATION FOUND!\n")
		}
	}

	// Step 3: Offer to fix the issues
	fmt.Print("\n🔧 Do you want to fix the outstanding amounts? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" && response != "yes" {
		fmt.Println("🛑 No changes made")
		return
	}

	// Fix the issues
	fmt.Println("\n🔧 FIXING OUTSTANDING AMOUNTS...")
	
	totalFixed := 0
	for _, sale := range problematicSales {
		// Update the outstanding amount
		err := db.Exec(`
			UPDATE sales 
			SET outstanding_amount = ?,
				status = CASE 
					WHEN ? <= 0.01 THEN 'PAID' 
					ELSE 'INVOICED' 
				END,
				updated_at = NOW()
			WHERE id = ?
		`, sale.ExpectedOutstanding, sale.ExpectedOutstanding, sale.SaleID).Error

		if err != nil {
			log.Printf("❌ Failed to fix sale %s: %v", sale.SaleCode, err)
		} else {
			log.Printf("✅ Fixed sale %s: Outstanding %.2f → %.2f", 
				sale.SaleCode, sale.OutstandingAmount, sale.ExpectedOutstanding)
			totalFixed++
		}
	}

	fmt.Printf("\n🎉 Successfully fixed %d out of %d sales!\n", totalFixed, len(problematicSales))

	// Step 4: Verify the fix
	fmt.Println("\n🔍 VERIFICATION:")
	
	var remainingIssues int64
	db.Raw(`
		SELECT COUNT(*)
		FROM sales s
		LEFT JOIN payment_allocations pa ON pa.invoice_id = s.id
		WHERE s.status IN ('INVOICED', 'PAID')
		GROUP BY s.id
		HAVING s.outstanding_amount != (s.total_amount - COALESCE(SUM(pa.allocated_amount), 0))
	`).Scan(&remainingIssues)

	if remainingIssues == 0 {
		fmt.Println("✅ SUCCESS: All sales outstanding amounts are now correct!")
	} else {
		fmt.Printf("⚠️ Warning: %d sales still have issues\n", remainingIssues)
	}
}