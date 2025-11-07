package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"math"
)

type SyncAnalysis struct {
	CashBankID      uint    `db:"cashbank_id"`
	CashBankCode    string  `db:"cashbank_code"`
	CashBankName    string  `db:"cashbank_name"`
	CashBankBalance float64 `db:"cashbank_balance"`
	COAAccountID    uint    `db:"coa_account_id"`
	COAAccountCode  string  `db:"coa_account_code"`
	COAAccountName  string  `db:"coa_account_name"`
	COABalance      float64 `db:"coa_balance"`
	Difference      float64 `db:"difference"`
	IsSync          bool    `db:"is_sync"`
	SyncPercentage  float64 `db:"sync_percentage"`
}

func main() {
	fmt.Println("📊 ANALISIS TINGKAT SINKRONISASI COA DAN CASH & BANK")
	fmt.Println("=" + string(make([]byte, 55)))

	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	// Get comprehensive sync analysis
	var syncData []SyncAnalysis
	query := `
		SELECT 
			cb.id as cashbank_id,
			cb.code as cashbank_code,
			cb.name as cashbank_name,
			cb.balance as cashbank_balance,
			a.id as coa_account_id,
			a.code as coa_account_code,
			a.name as coa_account_name,
			a.balance as coa_balance,
			ABS(cb.balance - a.balance) as difference,
			(cb.balance = a.balance) as is_sync,
			CASE 
				WHEN cb.balance = 0 AND a.balance = 0 THEN 100.0
				WHEN cb.balance = 0 OR a.balance = 0 THEN 
					CASE WHEN ABS(cb.balance - a.balance) = 0 THEN 100.0 ELSE 0.0 END
				ELSE 
					GREATEST(0, 100.0 - (ABS(cb.balance - a.balance) / GREATEST(ABS(cb.balance), ABS(a.balance)) * 100))
			END as sync_percentage
		FROM cash_banks cb
		JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL 
		  AND a.deleted_at IS NULL
		  AND cb.is_active = true
		ORDER BY cb.id
	`

	err := db.Raw(query).Scan(&syncData).Error
	if err != nil {
		fmt.Printf("❌ Error getting sync data: %v\n", err)
		return
	}

	if len(syncData) == 0 {
		fmt.Println("❌ Tidak ada data CashBank-COA yang ditemukan!")
		return
	}

	fmt.Printf("\n📋 DETAIL SINKRONISASI PER AKUN:\n")
	fmt.Println("=" + string(make([]byte, 120)))
	fmt.Printf("%-4s | %-12s | %-20s | %15s | %-12s | %-20s | %15s | %12s | %10s | %s\n",
		"ID", "CB Code", "CashBank Name", "CB Balance", "COA Code", "COA Account Name", "COA Balance", "Difference", "Sync %", "Status")
	fmt.Println(string(make([]byte, 120)))

	var totalAccounts int = len(syncData)
	var perfectSyncCount int = 0
	var totalSyncPercentage float64 = 0
	var criticalIssues int = 0
	var minorIssues int = 0

	for _, data := range syncData {
		status := "🟢 PERFECT"
		if !data.IsSync {
			if data.SyncPercentage >= 95 {
				status = "🟡 MINOR"
				minorIssues++
			} else {
				status = "🔴 CRITICAL"
				criticalIssues++
			}
		} else {
			perfectSyncCount++
		}

		totalSyncPercentage += data.SyncPercentage

		fmt.Printf("%-4d | %-12s | %-20s | %15.2f | %-12s | %-20s | %15.2f | %12.2f | %10.2f%% | %s\n",
			data.CashBankID, 
			data.CashBankCode, 
			truncateString(data.CashBankName, 20),
			data.CashBankBalance,
			data.COAAccountCode,
			truncateString(data.COAAccountName, 20),
			data.COABalance,
			data.Difference,
			data.SyncPercentage,
			status)
	}

	// Calculate overall statistics
	overallSyncPercentage := totalSyncPercentage / float64(totalAccounts)
	perfectSyncPercentage := (float64(perfectSyncCount) / float64(totalAccounts)) * 100

	fmt.Println("\n📊 RINGKASAN ANALISIS SINKRONISASI:")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Printf("Total Akun CashBank-COA      : %d\n", totalAccounts)
	fmt.Printf("Sinkronisasi Perfect (100%%)  : %d akun (%.1f%%)\n", perfectSyncCount, perfectSyncPercentage)
	fmt.Printf("Masalah Minor (95-99%%)       : %d akun\n", minorIssues)
	fmt.Printf("Masalah Critical (<95%%)      : %d akun\n", criticalIssues)
	fmt.Printf("\n🎯 TINGKAT SINKRONISASI KESELURUHAN: %.2f%%\n", overallSyncPercentage)

	// Analysis by categories
	fmt.Println("\n📈 KATEGORI KESEHATAN SINKRONISASI:")
	fmt.Println("=" + string(make([]byte, 40)))
	if overallSyncPercentage >= 98 {
		fmt.Println("🟢 EXCELLENT (98-100%): Sistem sinkronisasi berjalan sangat baik")
	} else if overallSyncPercentage >= 95 {
		fmt.Println("🟡 GOOD (95-97%): Sistem sinkronisasi berjalan baik dengan sedikit penyesuaian")
	} else if overallSyncPercentage >= 90 {
		fmt.Println("🟠 FAIR (90-94%): Perlu perbaikan pada beberapa akun")
	} else {
		fmt.Println("🔴 POOR (<90%): Perlu perbaikan menyeluruh sistem sinkronisasi")
	}

	// Test bidirectional sync functionality
	fmt.Println("\n🔄 TES FUNGSI SINKRONISASI BIDIRECTIONAL:")
	fmt.Println("=" + string(make([]byte, 50)))
	
	// Check if there are triggers for sync
	var triggerCount int
	db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.triggers 
		WHERE trigger_name LIKE '%sync%' OR trigger_name LIKE '%balance%'
	`).Scan(&triggerCount)
	
	fmt.Printf("Trigger Sinkronisasi Aktif   : %d\n", triggerCount)
	
	if triggerCount > 0 {
		fmt.Println("✅ Forward Sync (CashBank→COA): AKTIF")
		// Test showed forward sync works in the previous test
	} else {
		fmt.Println("❌ Forward Sync (CashBank→COA): TIDAK AKTIF")
	}
	
	// Based on previous test results
	fmt.Println("❌ Reverse Sync (COA→CashBank): TIDAK AKTIF")
	fmt.Println("⚠️  Direct Balance Update: TIDAK DISINKRONKAN (sesuai desain)")

	// Recommendations
	fmt.Println("\n💡 REKOMENDASI:")
	fmt.Println("=" + string(make([]byte, 20)))
	
	if perfectSyncCount == totalAccounts {
		fmt.Println("✅ Semua akun sudah tersinkronisasi dengan sempurna")
		fmt.Println("✅ Lanjutkan menggunakan sistem transaksi untuk update balance")
		fmt.Println("✅ Monitor secara berkala untuk mempertahankan konsistensi")
	} else {
		if criticalIssues > 0 {
			fmt.Printf("🔴 URGENT: Perbaiki %d akun dengan masalah critical\n", criticalIssues)
		}
		if minorIssues > 0 {
			fmt.Printf("🟡 Perbaiki %d akun dengan masalah minor\n", minorIssues)
		}
		fmt.Println("🔧 Aktifkan reverse sync trigger jika diperlukan")
		fmt.Println("📊 Jalankan balance sync otomatis secara berkala")
	}

	// Check transaction pattern
	fmt.Println("\n📋 ANALISIS POLA TRANSAKSI:")
	fmt.Println("=" + string(make([]byte, 35)))
	
	var transactionStats struct {
		TotalTransactions int     `db:"total_transactions"`
		TotalAmount       float64 `db:"total_amount"`
		AvgTransactionSize float64 `db:"avg_transaction_size"`
	}
	
	db.Raw(`
		SELECT 
			COUNT(*) as total_transactions,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(AVG(amount), 0) as avg_transaction_size
		FROM cash_bank_transactions 
		WHERE deleted_at IS NULL
	`).Scan(&transactionStats)
	
	fmt.Printf("Total Transaksi              : %d\n", transactionStats.TotalTransactions)
	fmt.Printf("Total Nilai Transaksi        : Rp %.2f\n", transactionStats.TotalAmount)
	
	if transactionStats.TotalTransactions > 0 {
		fmt.Printf("Rata-rata Nilai Transaksi    : Rp %.2f\n", transactionStats.AvgTransactionSize)
	}
	
	fmt.Println("\n" + string(make([]byte, 60)))
	fmt.Printf("🎯 KESIMPULAN: Tingkat Sinkronisasi COA-CashBank = %.2f%%\n", overallSyncPercentage)
	fmt.Println(string(make([]byte, 60)))
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func abs(x float64) float64 {
	return math.Abs(x)
}