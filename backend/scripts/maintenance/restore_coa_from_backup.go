package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== SISTEM AKUNTANSI - RESTORE COA FROM BACKUP ===")
	fmt.Println("")
	fmt.Println("Script ini akan mengembalikan Chart of Accounts dari backup")
	fmt.Println("⚠️  PERINGATAN: Ini akan menimpa COA yang ada saat ini!")
	fmt.Println("")

	// Load config dan connect ke database
	_ = config.LoadConfig() // Load config untuk inisialisasi environment
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Gagal koneksi ke database")
	}

	fmt.Println("🔗 Berhasil terhubung ke database")

	// Check if backup exists
	if !checkBackupExists(db) {
		fmt.Println("❌ Backup COA tidak ditemukan!")
		fmt.Println("Jalankan reset terlebih dahulu untuk membuat backup.")
		return
	}

	// Show backup info
	showBackupInfo(db)

	// Konfirmasi dari user
	if !confirmRestore() {
		fmt.Println("Restore dibatalkan oleh user.")
		return
	}

	// Execute restore
	fmt.Println("\n🔄 Mengeksekusi restore COA...")
	if err := executeRestore(db); err != nil {
		log.Printf("❌ Gagal restore COA: %v", err)
		return
	}

	fmt.Println("✅ Restore COA selesai!")
}

func checkBackupExists(db *gorm.DB) bool {
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM accounts_backup").Scan(&count).Error
	return err == nil && count > 0
}

func showBackupInfo(db *gorm.DB) {
	var backupInfo struct {
		AccountsCount    int64 `json:"accounts_count"`
		HierarchyCount   int64 `json:"hierarchy_count"`
		BalancesCount    int64 `json:"balances_count"`
		CurrentCount     int64 `json:"current_count"`
	}

	db.Raw("SELECT COUNT(*) FROM accounts_backup").Scan(&backupInfo.AccountsCount)
	db.Raw("SELECT COUNT(*) FROM accounts_hierarchy_backup").Scan(&backupInfo.HierarchyCount)
	db.Raw("SELECT COUNT(*) FROM accounts_original_balances").Scan(&backupInfo.BalancesCount)
	db.Model(&models.Account{}).Where("deleted_at IS NULL").Count(&backupInfo.CurrentCount)

	fmt.Printf("📦 Informasi backup:\n")
	fmt.Printf("   - Accounts backup: %d records\n", backupInfo.AccountsCount)
	fmt.Printf("   - Hierarchy backup: %d records\n", backupInfo.HierarchyCount)
	fmt.Printf("   - Original balances: %d records\n", backupInfo.BalancesCount)
	fmt.Printf("   - COA saat ini: %d records\n", backupInfo.CurrentCount)
}

func confirmRestore() bool {
	fmt.Print("\nApakah Anda yakin ingin restore COA dari backup? (ketik 'ya' untuk konfirmasi): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	
	return input == "ya" || input == "y" || input == "yes"
}

func executeRestore(db *gorm.DB) error {
	tx := db.Begin()

	// 1. Clear current accounts table
	fmt.Println("   1. Menghapus COA saat ini...")
	if err := tx.Exec("DELETE FROM accounts").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal menghapus COA saat ini: %v", err)
	}

	// 2. Restore from backup
	fmt.Println("   2. Restore dari backup...")
	if err := tx.Exec(`
		INSERT INTO accounts (id, code, name, description, type, category, parent_id, level, is_header, is_active, balance, created_at, updated_at)
		SELECT id, code, name, description, type, category, parent_id, level, is_header, is_active, balance, created_at, updated_at
		FROM accounts_backup
	`).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal restore dari backup: %v", err)
	}

	// 3. Reset sequence
	fmt.Println("   3. Reset sequence...")
	var maxID uint
	tx.Raw("SELECT MAX(id) FROM accounts").Scan(&maxID)
	if err := tx.Exec(fmt.Sprintf("ALTER SEQUENCE accounts_id_seq RESTART WITH %d", maxID+1)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal reset sequence: %v", err)
	}

	// 4. Log restore activity
	if err := tx.Exec(`
		INSERT INTO audit_logs (action, table_name, record_id, old_values, new_values, user_id, created_at)
		VALUES ('UPDATE', 'accounts', 0, '', 'Restore COA from backup', 1, CURRENT_TIMESTAMP)
	`).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal log restore activity: %v", err)
	}

	return tx.Commit().Error
}
