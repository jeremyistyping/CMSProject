package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Account struct {
	ID       uint   `gorm:"primaryKey"`
	Code     string `gorm:"unique;not null"`
	Name     string `gorm:"not null"`
	Type     string
	Category string
}

type TaxAccountSettings struct {
	ID                        uint `gorm:"primaryKey"`
	CompanyID                 *uint
	PurchaseInputVATAccountID uint
	SalesOutputVATAccountID   uint
	IsActive                  bool
}

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🔍 Checking PPN Masukan account...")

	// Check if account 1240 exists
	var ppnMasukan Account
	if err := db.Where("code = ?", "1240").First(&ppnMasukan).Error; err != nil {
		log.Fatalf("❌ Account 1240 (PPN Masukan) not found: %v", err)
	}
	fmt.Printf("✅ Found account 1240: %s - %s (ID: %d)\n", ppnMasukan.Code, ppnMasukan.Name, ppnMasukan.ID)

	// Check if account 2103 exists
	var ppnKeluaran Account
	if err := db.Where("code = ?", "2103").First(&ppnKeluaran).Error; err != nil {
		log.Fatalf("❌ Account 2103 (PPN Keluaran) not found: %v", err)
	}
	fmt.Printf("✅ Found account 2103: %s - %s (ID: %d)\n", ppnKeluaran.Code, ppnKeluaran.Name, ppnKeluaran.ID)

	fmt.Println("\n🔍 Checking tax_account_settings...")

	// Get current tax settings
	var settings TaxAccountSettings
	if err := db.Where("is_active = ?", true).First(&settings).Error; err != nil {
		log.Fatalf("❌ Active tax_account_settings not found: %v", err)
	}

	fmt.Printf("📋 Current settings (ID: %d):\n", settings.ID)
	fmt.Printf("   - purchase_input_vat_account_id: %d\n", settings.PurchaseInputVATAccountID)
	fmt.Printf("   - sales_output_vat_account_id: %d\n", settings.SalesOutputVATAccountID)

	// Fix if necessary
	needsUpdate := false
	updates := map[string]interface{}{}

	if settings.PurchaseInputVATAccountID != ppnMasukan.ID {
		fmt.Printf("\n⚠️  PPN Masukan mapping is incorrect: %d (should be %d)\n", settings.PurchaseInputVATAccountID, ppnMasukan.ID)
		updates["purchase_input_vat_account_id"] = ppnMasukan.ID
		needsUpdate = true
	}

	if settings.SalesOutputVATAccountID != ppnKeluaran.ID {
		fmt.Printf("\n⚠️  PPN Keluaran mapping is incorrect: %d (should be %d)\n", settings.SalesOutputVATAccountID, ppnKeluaran.ID)
		updates["sales_output_vat_account_id"] = ppnKeluaran.ID
		needsUpdate = true
	}

	if needsUpdate {
		fmt.Println("\n🔧 Updating tax_account_settings...")
		if err := db.Model(&TaxAccountSettings{}).Where("id = ?", settings.ID).Updates(updates).Error; err != nil {
			log.Fatalf("❌ Failed to update: %v", err)
		}
		fmt.Println("✅ Tax account settings updated successfully!")
	} else {
		fmt.Println("\n✅ Tax account settings are already correct!")
	}

	// Verify
	var updatedSettings TaxAccountSettings
	db.Where("id = ?", settings.ID).First(&updatedSettings)
	fmt.Printf("\n📋 Final settings:\n")
	fmt.Printf("   - purchase_input_vat_account_id: %d (%s)\n", updatedSettings.PurchaseInputVATAccountID, ppnMasukan.Code)
	fmt.Printf("   - sales_output_vat_account_id: %d (%s)\n", updatedSettings.SalesOutputVATAccountID, ppnKeluaran.Code)

	// Mark accounts as system critical
	fmt.Println("\n🔧 Marking PPN accounts as system critical...")
	if err := db.Exec("UPDATE accounts SET is_system_critical = TRUE WHERE code IN ('1240', '2103') AND deleted_at IS NULL").Error; err != nil {
		log.Printf("⚠️  Warning: Failed to mark accounts as critical: %v", err)
	} else {
		fmt.Println("✅ PPN accounts marked as system critical")
	}

	fmt.Println("\n✅ All done! PPN Masukan and Keluaran are now properly configured.")
}
