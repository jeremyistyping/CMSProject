package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== SISTEM AKUNTANSI - RESET DATA TRANSAKSI ===")
	fmt.Println("")
	fmt.Println("⚠️  PERINGATAN: Script ini akan menghapus SEMUA data transaksi!")
	fmt.Println("✅ Yang akan DIPERTAHANKAN:")
	fmt.Println("   - Chart of Accounts (COA)")
	fmt.Println("   - Master data produk")
	fmt.Println("   - Data kontak/customer/vendor")
	fmt.Println("   - Data user dan permission")
	fmt.Println("   - Master data cash bank")
	fmt.Println("")
	fmt.Println("❌ Yang akan DIHAPUS:")
	fmt.Println("   - Semua transaksi penjualan")
	fmt.Println("   - Semua transaksi pembelian") 
	fmt.Println("   - Semua jurnal entry")
	fmt.Println("   - Semua payment records")
	fmt.Println("   - Semua inventory movements")
	fmt.Println("   - Semua expense records")
	fmt.Println("   - Balance accounts akan direset ke 0")
	fmt.Println("   - Stock produk akan direset ke 0")
	fmt.Println("")

	// Konfirmasi dari user
	if !confirmReset() {
		fmt.Println("Reset dibatalkan oleh user.")
		return
	}

	// Load config dan connect ke database
	_ = config.LoadConfig() // Load config untuk inisialisasi environment
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Gagal koneksi ke database")
	}

	fmt.Println("🔗 Berhasil terhubung ke database")

	// Step 1: Backup COA
	fmt.Println("\n📦 STEP 1: Backup Chart of Accounts...")
	if err := backupCOA(db); err != nil {
		log.Printf("❌ Gagal backup COA: %v", err)
		return
	}
	fmt.Println("✅ Backup COA berhasil")

	// Step 2: Show current data summary
	fmt.Println("\n📊 STEP 2: Summary data saat ini...")
	showCurrentDataSummary(db)

	// Final confirmation
	fmt.Println("\n⚠️  KONFIRMASI TERAKHIR:")
	fmt.Print("Ketik 'RESET SEKARANG' untuk melanjutkan: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input != "RESET SEKARANG" {
		fmt.Println("Reset dibatalkan. Input tidak sesuai.")
		return
	}

	// Step 3: Execute reset
	fmt.Println("\n🔄 STEP 3: Mengeksekusi reset data transaksi...")
	if err := executeTransactionReset(db); err != nil {
		log.Printf("❌ Gagal reset data transaksi: %v", err)
		return
	}

	// Step 4: Verify reset
	fmt.Println("\n✅ STEP 4: Verifikasi hasil reset...")
	showPostResetSummary(db)

	fmt.Println("\n🎉 RESET SELESAI!")
	fmt.Println("Database siap digunakan dengan COA yang bersih.")
	fmt.Println("Anda bisa mulai input transaksi baru dari 0.")
}

func confirmReset() bool {
	fmt.Print("\nApakah Anda yakin ingin melanjutkan? (ketik 'ya' untuk konfirmasi): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	
	return input == "ya" || input == "y" || input == "yes"
}

func backupCOA(db *gorm.DB) error {
	// Read and execute backup script (safe version)
	backupSQL, err := os.ReadFile("scripts/backup_coa_safe.sql")
	if err != nil {
		return fmt.Errorf("gagal membaca backup script: %v", err)
	}

	// Execute backup script
	if err := db.Exec(string(backupSQL)).Error; err != nil {
		return fmt.Errorf("gagal eksekusi backup: %v", err)
	}

	// Log backup completion manually (since audit_logs structure varies)
	fmt.Println("   📝 Logging backup activity...")
	logErr := db.Exec(`
		INSERT INTO audit_logs (action, table_name, record_id, old_values, new_values, user_id, created_at)
		VALUES ('CREATE', 'accounts_backup', 0, '', 'Backup COA before transaction data reset', 1, CURRENT_TIMESTAMP)
	`).Error
	
	if logErr != nil {
		fmt.Printf("   ⚠️  Warning: Gagal log ke audit_logs: %v\n", logErr)
		fmt.Println("   ℹ️  Backup tetap berhasil, hanya logging yang gagal")
	}

	return nil
}

func showCurrentDataSummary(db *gorm.DB) {
	type DataSummary struct {
		TotalAccounts      int64 `json:"total_accounts"`
		TotalSales         int64 `json:"total_sales"`
		TotalPurchases     int64 `json:"total_purchases"`
		TotalTransactions  int64 `json:"total_transactions"`
		TotalJournals      int64 `json:"total_journals"`
		TotalExpenses      int64 `json:"total_expenses"`
		TotalPayments      int64 `json:"total_payments"`
		TotalInventory     int64 `json:"total_inventory"`
		TotalProducts      int64 `json:"total_products"`
		TotalContacts      int64 `json:"total_contacts"`
		TotalCashBanks     int64 `json:"total_cash_banks"`
	}

	var summary DataSummary

	db.Model(&models.Account{}).Where("deleted_at IS NULL").Count(&summary.TotalAccounts)
	db.Model(&models.Sale{}).Count(&summary.TotalSales)
	db.Model(&models.Purchase{}).Count(&summary.TotalPurchases)
	db.Model(&models.Transaction{}).Count(&summary.TotalTransactions)
	db.Model(&models.Journal{}).Count(&summary.TotalJournals)
	db.Model(&models.Expense{}).Count(&summary.TotalExpenses)
	db.Model(&models.Payment{}).Count(&summary.TotalPayments)
	db.Model(&models.Inventory{}).Count(&summary.TotalInventory)
	db.Model(&models.Product{}).Where("deleted_at IS NULL").Count(&summary.TotalProducts)
	db.Model(&models.Contact{}).Where("deleted_at IS NULL").Count(&summary.TotalContacts)
	db.Model(&models.CashBank{}).Where("deleted_at IS NULL").Count(&summary.TotalCashBanks)

	fmt.Printf("📊 Data saat ini:\n")
	fmt.Printf("   COA Accounts: %d (akan DIPERTAHANKAN)\n", summary.TotalAccounts)
	fmt.Printf("   Products: %d (akan DIPERTAHANKAN, stock direset)\n", summary.TotalProducts)
	fmt.Printf("   Contacts: %d (akan DIPERTAHANKAN)\n", summary.TotalContacts)
	fmt.Printf("   Cash Banks: %d (akan DIPERTAHANKAN, balance direset)\n", summary.TotalCashBanks)
	fmt.Printf("   \n")
	fmt.Printf("   Sales: %d (akan DIHAPUS)\n", summary.TotalSales)
	fmt.Printf("   Purchases: %d (akan DIHAPUS)\n", summary.TotalPurchases)
	fmt.Printf("   Transactions: %d (akan DIHAPUS)\n", summary.TotalTransactions)
	fmt.Printf("   Journals: %d (akan DIHAPUS)\n", summary.TotalJournals)
	fmt.Printf("   Expenses: %d (akan DIHAPUS)\n", summary.TotalExpenses)
	fmt.Printf("   Payments: %d (akan DIHAPUS)\n", summary.TotalPayments)
	fmt.Printf("   Inventory: %d (akan DIHAPUS)\n", summary.TotalInventory)
}

func executeTransactionReset(db *gorm.DB) error {
	// Read and execute reset script (safe version)
	resetSQL, err := os.ReadFile("scripts/reset_transaction_data_safe.sql")
	if err != nil {
		return fmt.Errorf("gagal membaca reset script: %v", err)
	}

	start := time.Now()
	fmt.Printf("⏳ Mengeksekusi reset... (dimulai: %s)\n", start.Format("15:04:05"))
	
	// Execute reset script
	if err := db.Exec(string(resetSQL)).Error; err != nil {
		return fmt.Errorf("gagal eksekusi reset: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("✅ Reset selesai dalam waktu: %s\n", duration.String())

	return nil
}

func showPostResetSummary(db *gorm.DB) {
	type PostResetSummary struct {
		TotalAccounts      int64 `json:"total_accounts"`
		TotalProducts      int64 `json:"total_products"`  
		TotalContacts      int64 `json:"total_contacts"`
		TotalCashBanks     int64 `json:"total_cash_banks"`
		RemainingSales     int64 `json:"remaining_sales"`
		RemainingPurchases int64 `json:"remaining_purchases"`
		RemainingTransactions int64 `json:"remaining_transactions"`
		RemainingJournals  int64 `json:"remaining_journals"`
		AccountsWithBalance int64 `json:"accounts_with_balance"`
		ProductsWithStock  int64 `json:"products_with_stock"`
		CashBanksWithBalance int64 `json:"cash_banks_with_balance"`
	}

	var summary PostResetSummary

	// Count preserved data
	db.Model(&models.Account{}).Where("deleted_at IS NULL").Count(&summary.TotalAccounts)
	db.Model(&models.Product{}).Where("deleted_at IS NULL").Count(&summary.TotalProducts)
	db.Model(&models.Contact{}).Where("deleted_at IS NULL").Count(&summary.TotalContacts)
	db.Model(&models.CashBank{}).Where("deleted_at IS NULL").Count(&summary.TotalCashBanks)

	// Count remaining transaction data (should be 0)
	db.Model(&models.Sale{}).Count(&summary.RemainingSales)
	db.Model(&models.Purchase{}).Count(&summary.RemainingPurchases)
	db.Model(&models.Transaction{}).Count(&summary.RemainingTransactions)
	db.Model(&models.Journal{}).Count(&summary.RemainingJournals)

	// Count data with non-zero balances (should be 0)
	db.Model(&models.Account{}).Where("balance != 0 AND deleted_at IS NULL").Count(&summary.AccountsWithBalance)
	db.Model(&models.Product{}).Where("stock != 0 AND deleted_at IS NULL").Count(&summary.ProductsWithStock)
	db.Model(&models.CashBank{}).Where("balance != 0 AND deleted_at IS NULL").Count(&summary.CashBanksWithBalance)

	fmt.Printf("📊 Hasil reset:\n")
	fmt.Printf("✅ Data yang dipertahankan:\n")
	fmt.Printf("   - COA Accounts: %d\n", summary.TotalAccounts)
	fmt.Printf("   - Products: %d\n", summary.TotalProducts)
	fmt.Printf("   - Contacts: %d\n", summary.TotalContacts)
	fmt.Printf("   - Cash Banks: %d\n", summary.TotalCashBanks)
	fmt.Printf("\n")
	
	if summary.RemainingSales == 0 && summary.RemainingPurchases == 0 && 
	   summary.RemainingTransactions == 0 && summary.RemainingJournals == 0 {
		fmt.Printf("✅ Data transaksi berhasil dihapus:\n")
		fmt.Printf("   - Sales: %d\n", summary.RemainingSales)
		fmt.Printf("   - Purchases: %d\n", summary.RemainingPurchases)
		fmt.Printf("   - Transactions: %d\n", summary.RemainingTransactions)
		fmt.Printf("   - Journals: %d\n", summary.RemainingJournals)
	} else {
		fmt.Printf("⚠️  Peringatan - Sisa data transaksi:\n")
		fmt.Printf("   - Sales: %d\n", summary.RemainingSales)
		fmt.Printf("   - Purchases: %d\n", summary.RemainingPurchases)
		fmt.Printf("   - Transactions: %d\n", summary.RemainingTransactions)
		fmt.Printf("   - Journals: %d\n", summary.RemainingJournals)
	}

	fmt.Printf("\n")
	
	if summary.AccountsWithBalance == 0 && summary.ProductsWithStock == 0 && 
	   summary.CashBanksWithBalance == 0 {
		fmt.Printf("✅ Balance berhasil direset:\n")
		fmt.Printf("   - Accounts dengan balance > 0: %d\n", summary.AccountsWithBalance)
		fmt.Printf("   - Products dengan stock > 0: %d\n", summary.ProductsWithStock)
		fmt.Printf("   - Cash Banks dengan balance > 0: %d\n", summary.CashBanksWithBalance)
	} else {
		fmt.Printf("⚠️  Balance yang belum direset:\n")
		fmt.Printf("   - Accounts dengan balance > 0: %d\n", summary.AccountsWithBalance)
		fmt.Printf("   - Products dengan stock > 0: %d\n", summary.ProductsWithStock)
		fmt.Printf("   - Cash Banks dengan balance > 0: %d\n", summary.CashBanksWithBalance)
	}

	// Show backup tables info
	fmt.Println("\n📦 Tabel backup yang dibuat:")
	
	var backupTables []struct {
		TableName string `json:"table_name"`
		Count     int64  `json:"count"`
	}
	
	err := db.Raw(`
		SELECT 
			'accounts_backup' as table_name,
			COUNT(*) as count 
		FROM accounts_backup
		UNION ALL
		SELECT 
			'accounts_hierarchy_backup' as table_name,
			COUNT(*) as count 
		FROM accounts_hierarchy_backup
		UNION ALL
		SELECT 
			'accounts_original_balances' as table_name,
			COUNT(*) as count 
		FROM accounts_original_balances
	`).Scan(&backupTables).Error
	
	if err != nil {
		fmt.Printf("⚠️  Tidak dapat mengecek tabel backup: %v\n", err)
	} else {
		for _, backup := range backupTables {
			fmt.Printf("   - %s: %d records\n", backup.TableName, backup.Count)
		}
	}

	fmt.Println("\n🔧 Untuk restore COA dari backup (jika diperlukan):")
	fmt.Println("   go run cmd/restore_coa_from_backup.go")
}
