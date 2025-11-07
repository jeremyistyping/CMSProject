package main

import (
	"log"
	"os"
	"strconv"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Get sale ID from command line argument or use default
	saleID := 7 // Default to Sale #7
	if len(os.Args) > 1 {
		if id, err := strconv.Atoi(os.Args[1]); err == nil && id > 0 {
			saleID = id
		} else {
			log.Printf("Invalid sale ID, using default: %d", saleID)
		}
	}

	// Load config from environment
	cfg := config.LoadConfig()

	// Connect to database using config
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Connected to database successfully")

	log.Printf("=== Debugging Sale #%d Calculation ===\n", saleID)

	// 1. Get Sale with all relations
	var sale models.Sale
	if err := db.Preload("SaleItems.Product").Preload("Customer").First(&sale, saleID).Error; err != nil {
		log.Fatalf("Failed to load Sale #%d: %v", saleID, err)
	}

	log.Printf("📊 SALE #%d DATA:", sale.ID)
	log.Printf("  ID: %d", sale.ID)
	log.Printf("  Code: %s", sale.Code)
	log.Printf("  InvoiceNumber: %s", sale.InvoiceNumber)
	log.Printf("  Customer: %s", sale.Customer.Name)
	log.Printf("  Status: %s", sale.Status)
	log.Printf("  Date: %s", sale.Date)
	log.Println()

	// 2. Check all financial fields
	log.Println("💰 FINANCIAL FIELDS:")
	log.Printf("  Subtotal: %.2f", sale.Subtotal)
	log.Printf("  DiscountPercent: %.2f%%", sale.DiscountPercent)
	log.Printf("  DiscountAmount: %.2f", sale.DiscountAmount)
	log.Printf("  TaxableAmount: %.2f", sale.TaxableAmount)
	log.Printf("  NetBeforeTax: %.2f", sale.NetBeforeTax)
	log.Println()

	// 3. Tax fields (Legacy)
	log.Println("🏛️ LEGACY TAX FIELDS:")
	log.Printf("  Tax: %.2f", sale.Tax)
	log.Printf("  PPN: %.2f", sale.PPN)
	log.Printf("  PPNPercent: %.2f%%", sale.PPNPercent)
	log.Printf("  PPh: %.2f", sale.PPh)
	log.Printf("  PPhPercent: %.2f%%", sale.PPhPercent)
	log.Printf("  PPhType: %s", sale.PPhType)
	log.Printf("  TotalTax: %.2f", sale.TotalTax)
	log.Println()

	// 4. Enhanced Tax fields
	log.Println("🆕 ENHANCED TAX FIELDS:")
	log.Printf("  PPNRate: %.2f%%", sale.PPNRate)
	log.Printf("  PPNAmount: %.2f", sale.PPNAmount)
	log.Printf("  OtherTaxAdditions: %.2f", sale.OtherTaxAdditions)
	log.Printf("  TotalTaxAdditions: %.2f", sale.TotalTaxAdditions)
	log.Println()
	log.Printf("  PPh21Rate: %.2f%%", sale.PPh21Rate)
	log.Printf("  PPh21Amount: %.2f", sale.PPh21Amount)
	log.Printf("  PPh23Rate: %.2f%%", sale.PPh23Rate)
	log.Printf("  PPh23Amount: %.2f", sale.PPh23Amount)
	log.Printf("  OtherTaxDeductions: %.2f", sale.OtherTaxDeductions)
	log.Printf("  TotalTaxDeductions: %.2f", sale.TotalTaxDeductions)
	log.Println()

	// 5. Other fields
	log.Println("📦 OTHER FIELDS:")
	log.Printf("  ShippingCost: %.2f", sale.ShippingCost)
	log.Printf("  ShippingTaxable: %v", sale.ShippingTaxable)
	log.Printf("  PaymentMethodType: %s", sale.PaymentMethodType)
	if sale.CashBankID != nil {
		log.Printf("  CashBankID: %d", *sale.CashBankID)
	} else {
		log.Printf("  CashBankID: nil")
	}
	log.Println()

	// 6. Totals
	log.Println("💵 TOTALS:")
	log.Printf("  TotalAmount: %.2f", sale.TotalAmount)
	log.Printf("  PaidAmount: %.2f", sale.PaidAmount)
	log.Printf("  OutstandingAmount: %.2f", sale.OutstandingAmount)
	log.Println()

	// 7. Sale Items
	log.Println("📋 SALE ITEMS:")
	var itemsSubtotal float64
	for i, item := range sale.SaleItems {
		log.Printf("\n  Item #%d:", i+1)
		log.Printf("    Product: %s (ID: %d)", item.Product.Name, item.Product.ID)
		log.Printf("    Quantity: %d", item.Quantity)
		log.Printf("    UnitPrice: %.2f", item.UnitPrice)
		log.Printf("    DiscountPercent: %.2f%%", item.DiscountPercent)
		log.Printf("    DiscountAmount: %.2f", item.DiscountAmount)
		log.Printf("    LineTotal: %.2f", item.LineTotal)
		log.Printf("    Taxable: %v", item.Taxable)
		log.Printf("    PPNAmount: %.2f", item.PPNAmount)
		log.Printf("    PPhAmount: %.2f", item.PPhAmount)
		log.Printf("    TotalTax: %.2f", item.TotalTax)
		log.Printf("    FinalAmount: %.2f", item.FinalAmount)
		log.Printf("    Product CostPrice: %.2f", item.Product.CostPrice)
		
		itemsSubtotal += item.LineTotal
	}
	log.Printf("\n  Items Subtotal (sum of LineTotal): %.2f", itemsSubtotal)
	log.Println()

	// 8. Manual calculation
	log.Println("🔢 MANUAL CALCULATION:")
	
	// Calculate what the journal entry should be
	subtotal := sale.Subtotal
	ppn := sale.PPN
	otherTaxAdd := sale.OtherTaxAdditions
	shippingCost := sale.ShippingCost
	
	grossAmount := subtotal + ppn + otherTaxAdd + shippingCost
	log.Printf("  Gross = Subtotal(%.2f) + PPN(%.2f) + OtherTaxAdd(%.2f) + Shipping(%.2f) = %.2f", 
		subtotal, ppn, otherTaxAdd, shippingCost, grossAmount)
	
	pph := sale.PPh
	pph21 := sale.PPh21Amount
	pph23 := sale.PPh23Amount
	otherTaxDed := sale.OtherTaxDeductions
	totalDeductions := pph + pph21 + pph23 + otherTaxDed
	
	log.Printf("  Total Deductions = PPh(%.2f) + PPh21(%.2f) + PPh23(%.2f) + OtherTaxDed(%.2f) = %.2f",
		pph, pph21, pph23, otherTaxDed, totalDeductions)
	
	expectedTotal := grossAmount - totalDeductions
	log.Printf("  Expected TotalAmount = Gross(%.2f) - Deductions(%.2f) = %.2f", 
		grossAmount, totalDeductions, expectedTotal)
	log.Printf("  Actual TotalAmount in DB: %.2f", sale.TotalAmount)
	log.Printf("  DIFFERENCE: %.2f", expectedTotal-sale.TotalAmount)
	log.Println()

	// 9. Check COA accounts that will be used
	log.Println("🏦 COA ACCOUNTS CHECK:")
	checkAccount(db, "1101", "KAS (Cash)")
	checkAccount(db, "1102", "BANK")
	checkAccount(db, "1102-001", "BANK BCA")
	checkAccount(db, "1201", "PIUTANG USAHA (Receivables)")
	checkAccount(db, "2103", "PPN KELUARAN (VAT Output)")
	checkAccount(db, "2104", "PPh YANG DIPOTONG (Tax Withheld)")
	checkAccount(db, "2107", "PEMOTONGAN PAJAK LAINNYA (Other Tax Deductions)")
	checkAccount(db, "2108", "PENAMBAHAN PAJAK LAINNYA (Other Tax Additions)")
	checkAccount(db, "4101", "Pendapatan Penjualan (Revenue)")
	checkAccount(db, "4102", "PENDAPATAN JASA/ONGKIR (Shipping Revenue)")
	checkAccount(db, "5101", "HARGA POKOK PENJUALAN (COGS)")
	checkAccount(db, "1301", "PERSEDIAAN BARANG DAGANGAN (Inventory)")
	log.Println()

	// 10. Calculate COGS
	log.Println("💰 COGS CALCULATION:")
	var totalCOGS float64
	for i, item := range sale.SaleItems {
		itemCOGS := float64(item.Quantity) * item.Product.CostPrice
		totalCOGS += itemCOGS
		log.Printf("  Item #%d: %s - Qty: %d × CostPrice: %.2f = %.2f",
			i+1, item.Product.Name, item.Quantity, item.Product.CostPrice, itemCOGS)
	}
	log.Printf("  Total COGS: %.2f", totalCOGS)
	log.Println()

	// 11. Calculate journal entry balance
	log.Println("📊 JOURNAL ENTRY CALCULATION:")
	log.Println("  DEBIT side:")
	log.Printf("    Cash/Bank/Receivable: %.2f", grossAmount)
	log.Printf("    COGS: %.2f", totalCOGS)
	totalDebit := grossAmount + totalCOGS
	log.Printf("    TOTAL DEBIT: %.2f", totalDebit)
	log.Println()
	
	log.Println("  CREDIT side:")
	log.Printf("    Revenue (Subtotal): %.2f", subtotal)
	log.Printf("    PPN Keluaran: %.2f", ppn)
	log.Printf("    Other Tax Additions: %.2f", otherTaxAdd)
	log.Printf("    Shipping Revenue: %.2f", shippingCost)
	log.Printf("    Tax Deductions (PPh, etc): %.2f", totalDeductions)
	log.Printf("    Inventory: %.2f", totalCOGS)
	totalCredit := subtotal + ppn + otherTaxAdd + shippingCost + totalDeductions + totalCOGS
	log.Printf("    TOTAL CREDIT: %.2f", totalCredit)
	log.Println()
	
	difference := totalDebit - totalCredit
	log.Printf("  BALANCE: Debit(%.2f) - Credit(%.2f) = %.2f", totalDebit, totalCredit, difference)
	if difference == 0 {
		log.Println("  ✅ BALANCED!")
	} else {
		log.Printf("  ❌ IMBALANCED by %.2f!", difference)
	}
	log.Println()

	log.Printf("\n=== Analysis Complete for Sale #%d ===", sale.ID)
	
	// Print usage hint
	if saleID == 7 {
		log.Println("\n💡 TIP: To debug another sale, run: go run cmd/scripts/debug_sale_3_calc.go <sale_id>")
		log.Println("   Example: go run cmd/scripts/debug_sale_3_calc.go 5")
	}
}

func checkAccount(db *gorm.DB, code string, description string) {
	var account models.Account
	if err := db.Where("code = ?", code).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("  ❌ %s (%s): NOT FOUND", code, description)
		} else {
			log.Printf("  ⚠️ %s (%s): ERROR - %v", code, description, err)
		}
	} else {
		log.Printf("  ✅ %s - %s (ID: %d, Type: %s, Balance: %.2f)", 
			code, account.Name, account.ID, account.Type, account.Balance)
	}
}
