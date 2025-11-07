package main

import (
	"fmt"
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Account struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Code        string `json:"code" gorm:"uniqueIndex;size:20"`
	Name        string `json:"name" gorm:"size:255"`
	Type        string `json:"type" gorm:"size:50"`
	Balance     float64 `json:"balance" gorm:"default:0"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	ParentID    *uint  `json:"parent_id"`
	IsHeader    bool   `json:"is_header" gorm:"default:false"`
	Level       int    `json:"level" gorm:"default:1"`
	Description string `json:"description"`
}

func main() {
	// Connect to database
	dsn := "accounting_user:Bismillah2024!@tcp(localhost:3306)/accounting_system?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔍 INVESTIGASI DUPLIKASI AKUN KAS")
	fmt.Println("================================")

	// 1. Cari semua akun yang mengandung kata "Kas"
	var kasAccounts []Account
	db.Where("name LIKE ?", "%Kas%").Order("code").Find(&kasAccounts)

	fmt.Printf("\n📋 SEMUA AKUN YANG MENGANDUNG 'KAS' (%d akun):\n", len(kasAccounts))
	fmt.Println("ID\tCode\t\tName\t\t\tType\t\tBalance\t\tActive\tHeader")
	fmt.Println("---\t----\t\t----\t\t\t----\t\t-------\t\t------\t------")
	for _, acc := range kasAccounts {
		fmt.Printf("%d\t%s\t\t%-20s\t%s\t\t%.2f\t\t%t\t%t\n", 
			acc.ID, acc.Code, acc.Name, acc.Type, acc.Balance, acc.IsActive, acc.IsHeader)
	}

	// 2. Cari duplikasi berdasarkan nama yang sama
	fmt.Printf("\n🔍 MENCARI DUPLIKASI BERDASARKAN NAMA:\n")
	var duplicates []struct {
		Name  string
		Count int64
	}
	
	db.Model(&Account{}).
		Select("name, COUNT(*) as count").
		Group("name").
		Having("COUNT(*) > 1").
		Scan(&duplicates)

	if len(duplicates) > 0 {
		fmt.Printf("Ditemukan %d nama akun yang duplikasi:\n", len(duplicates))
		for _, dup := range duplicates {
			fmt.Printf("- '%s' muncul %d kali\n", dup.Name, dup.Count)
			
			// Tampilkan detail akun dengan nama yang sama
			var accounts []Account
			db.Where("name = ?", dup.Name).Order("code").Find(&accounts)
			for _, acc := range accounts {
				fmt.Printf("  → ID: %d, Code: %s, Active: %t, Header: %t\n", 
					acc.ID, acc.Code, acc.IsActive, acc.IsHeader)
			}
		}
	} else {
		fmt.Println("Tidak ada duplikasi nama akun.")
	}

	// 3. Periksa struktur hierarki akun Kas
	fmt.Printf("\n🏗️ STRUKTUR HIERARKI AKUN KAS:\n")
	for _, acc := range kasAccounts {
		if acc.ParentID != nil {
			var parent Account
			db.First(&parent, *acc.ParentID)
			fmt.Printf("%s (%s) → Parent: %s (%s)\n", 
				acc.Code, acc.Name, parent.Code, parent.Name)
		} else {
			fmt.Printf("%s (%s) → Root account\n", acc.Code, acc.Name)
		}
	}

	// 4. Cek apakah ada akun non-aktif
	var inactiveKas []Account
	db.Where("name LIKE ? AND is_active = false", "%Kas%").Find(&inactiveKas)
	
	if len(inactiveKas) > 0 {
		fmt.Printf("\n⚠️ AKUN KAS TIDAK AKTIF (%d akun):\n", len(inactiveKas))
		for _, acc := range inactiveKas {
			fmt.Printf("- %s (%s) - Inactive\n", acc.Code, acc.Name)
		}
	}

	// 5. Rekomendasi perbaikan
	fmt.Printf("\n💡 ANALISIS & REKOMENDASI:\n")
	fmt.Println("========================")
	
	if len(kasAccounts) > 1 {
		fmt.Printf("Ditemukan %d akun dengan kata 'Kas'.\n", len(kasAccounts))
		
		// Identifikasi akun mana yang seharusnya digunakan
		var validKas *Account
		var duplicateKas []Account
		
		for i := range kasAccounts {
			acc := &kasAccounts[i]
			// Prioritas: aktif, bukan header, kode standar
			if acc.IsActive && !acc.IsHeader {
				if acc.Code == "1101" || acc.Code == "1100-001" {
					validKas = acc
				} else {
					duplicateKas = append(duplicateKas, *acc)
				}
			}
		}
		
		if validKas != nil {
			fmt.Printf("✅ Akun Kas yang valid: %s (%s)\n", validKas.Code, validKas.Name)
		}
		
		if len(duplicateKas) > 0 {
			fmt.Printf("❌ Akun Kas yang perlu dihapus/dinonaktifkan:\n")
			for _, acc := range duplicateKas {
				fmt.Printf("   - %s (%s) ID: %d\n", acc.Code, acc.Name, acc.ID)
			}
		}
	}

	// 6. Periksa journal entries yang menggunakan akun duplikat
	fmt.Printf("\n📊 PENGGUNAAN DALAM JOURNAL ENTRIES:\n")
	
	type JournalLineUsage struct {
		AccountID    uint
		AccountCode  string 
		AccountName  string
		EntryCount   int64
		TotalDebit   float64
		TotalCredit  float64
	}
	
	var usage []JournalLineUsage
	for _, acc := range kasAccounts {
		var count int64
		var totalDebit, totalCredit float64
		
		db.Table("journal_lines").
			Where("account_id = ?", acc.ID).
			Count(&count)
			
		if count > 0 {
			db.Table("journal_lines").
				Where("account_id = ?", acc.ID).
				Select("COALESCE(SUM(debit_amount), 0) as total_debit, COALESCE(SUM(credit_amount), 0) as total_credit").
				Row().Scan(&totalDebit, &totalCredit)
		}
		
		usage = append(usage, JournalLineUsage{
			AccountID:   acc.ID,
			AccountCode: acc.Code,
			AccountName: acc.Name,
			EntryCount:  count,
			TotalDebit:  totalDebit,
			TotalCredit: totalCredit,
		})
	}
	
	for _, u := range usage {
		fmt.Printf("%s (%s): %d entries, Debit: %.2f, Credit: %.2f\n", 
			u.AccountCode, u.AccountName, u.EntryCount, u.TotalDebit, u.TotalCredit)
	}
}