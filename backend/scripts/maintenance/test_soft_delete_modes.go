package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== TEST SOFT DELETE MODES ===")
	fmt.Println("Script untuk menguji kemampuan soft delete dan recovery")
	
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Gagal koneksi ke database")
	}
	
	fmt.Println("🔗 Berhasil terhubung ke database")
	
	// Test 1: Cek tabel yang memiliki soft delete
	fmt.Println("\n📋 TEST 1: Tabel dengan soft delete capability")
	testSoftDeleteCapability(db)
	
	// Test 2: Hitung data aktif dan soft deleted
	fmt.Println("\n📊 TEST 2: Summary data aktif vs soft deleted")
	testDataSummary(db)
	
	// Test 3: Simulasi soft delete pada 1 tabel
	fmt.Println("\n🧪 TEST 3: Simulasi soft delete (hanya untuk testing)")
	// Uncomment line di bawah jika ingin test simulasi
	// testSoftDeleteSimulation(db)
	
	fmt.Println("\n✅ Test selesai!")
}

func testSoftDeleteCapability(db *gorm.DB) {
	type tbl struct{ TableName string }
	var tables []tbl
	
	err := db.Raw(`
		SELECT table_name AS table_name
		FROM information_schema.columns
		WHERE table_schema = 'public' 
		AND column_name = 'deleted_at'
		ORDER BY table_name
	`).Scan(&tables).Error
	
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	
	fmt.Printf("🔍 Ditemukan %d tabel dengan soft delete capability:\n", len(tables))
	for i, t := range tables {
		fmt.Printf("   %d. %s\n", i+1, t.TableName)
	}
}

func testDataSummary(db *gorm.DB) {
	type tbl struct{ TableName string }
	var tables []tbl
	
	db.Raw(`
		SELECT table_name AS table_name
		FROM information_schema.columns
		WHERE table_schema = 'public' 
		AND column_name = 'deleted_at'
		ORDER BY table_name
	`).Scan(&tables)
	
	// Daftar pengecualian (tabel sistem/backup)
	exclude := map[string]bool{
		"accounts_backup":                 true,
		"accounts_hierarchy_backup":       true,
		"accounts_original_balances":      true,
		"schema_migrations":               true,
		"gorm_migrations":                 true,
	}
	
	totalActive := 0
	totalSoftDeleted := 0
	
	fmt.Printf("%-25s | %-10s | %-10s | %-10s\n", "Table", "Total", "Active", "Soft Del")
	fmt.Println("---------|-----------|-----------|----------")
	
	for _, t := range tables {
		name := t.TableName
		if exclude[name] {
			continue
		}
		
		// Hitung total records
		var total, active, softDeleted int64
		db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", name)).Scan(&total)
		db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL", name)).Scan(&active)
		db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NOT NULL", name)).Scan(&softDeleted)
		
		if total > 0 {
			fmt.Printf("%-25s | %-10d | %-10d | %-10d\n", name, total, active, softDeleted)
		}
		
		totalActive += int(active)
		totalSoftDeleted += int(softDeleted)
	}
	
	fmt.Println("---------|-----------|-----------|----------")
	fmt.Printf("%-25s | %-10s | %-10d | %-10d\n", "TOTAL", "-", totalActive, totalSoftDeleted)
	
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   - Records aktif: %d\n", totalActive)
	fmt.Printf("   - Records soft deleted: %d\n", totalSoftDeleted)
	fmt.Printf("   - Total records: %d\n", totalActive+totalSoftDeleted)
}

// Fungsi simulasi - hanya untuk testing (commented out by default)
func testSoftDeleteSimulation(db *gorm.DB) {
	fmt.Println("⚠️  SIMULASI: Soft delete pada tabel 'notifications' (jika ada)")
	
	// Cek apakah tabel notifications ada
	var exists bool
	db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'notifications')").Scan(&exists)
	
	if !exists {
		fmt.Println("   ℹ Tabel 'notifications' tidak ada, skip simulasi")
		return
	}
	
	// Hitung records aktif sebelum simulasi
	var countBefore int64
	db.Raw("SELECT COUNT(*) FROM notifications WHERE deleted_at IS NULL").Scan(&countBefore)
	fmt.Printf("   Records aktif sebelum: %d\n", countBefore)
	
	if countBefore == 0 {
		fmt.Println("   ℹ Tidak ada data aktif untuk simulasi soft delete")
		return
	}
	
	fmt.Println("   🧪 Melakukan soft delete pada 1 record pertama...")
	
	// Soft delete 1 record pertama
	err := db.Exec("UPDATE notifications SET deleted_at = NOW() WHERE deleted_at IS NULL LIMIT 1").Error
	if err != nil {
		fmt.Printf("   ❌ Error simulasi: %v\n", err)
		return
	}
	
	// Hitung records setelah simulasi
	var countAfter int64
	db.Raw("SELECT COUNT(*) FROM notifications WHERE deleted_at IS NULL").Scan(&countAfter)
	fmt.Printf("   Records aktif sesudah: %d\n", countAfter)
	
	if countAfter == countBefore-1 {
		fmt.Println("   ✅ Simulasi soft delete berhasil!")
		
		// Recovery otomatis
		fmt.Println("   🔄 Melakukan recovery otomatis...")
		err = db.Exec("UPDATE notifications SET deleted_at = NULL WHERE deleted_at IS NOT NULL").Error
		if err != nil {
			fmt.Printf("   ❌ Error recovery: %v\n", err)
			return
		}
		
		var countRecovered int64
		db.Raw("SELECT COUNT(*) FROM notifications WHERE deleted_at IS NULL").Scan(&countRecovered)
		fmt.Printf("   Records setelah recovery: %d\n", countRecovered)
		
		if countRecovered == countBefore {
			fmt.Println("   ✅ Recovery berhasil! Data kembali seperti semula")
		}
	}
}