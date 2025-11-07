package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	log.Println("=== PAYMENT SYSTEM HEALTH MONITOR ===")
	
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	fmt.Printf("Monitor run at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("===============================================")

	// 1. Payment Status Overview
	checkPaymentStatus(db)

	// 2. Journal Entry Integrity
	checkJournalIntegrity(db)

	// 3. Cash/Bank Balance Health
	checkCashBankHealth(db)

	// 4. Recent Activity
	checkRecentActivity(db)

	// 5. System Health Score
	calculateHealthScore(db)

	fmt.Println("===============================================")
	fmt.Println("✅ Payment system monitoring complete")
}

func checkPaymentStatus(db *gorm.DB) {
	fmt.Println("\n🔍 PAYMENT STATUS OVERVIEW")
	fmt.Println("---------------------------")

	var totalPayments, completedPayments, pendingPayments, failedPayments int64
	var totalAmount, completedAmount, pendingAmount float64

	// Count payments by status
	db.Model(&models.Payment{}).Count(&totalPayments)
	db.Model(&models.Payment{}).Where("status = ?", "COMPLETED").Count(&completedPayments)
	db.Model(&models.Payment{}).Where("status = ?", "PENDING").Count(&pendingPayments)
	db.Model(&models.Payment{}).Where("status = ?", "FAILED").Count(&failedPayments)

	// Sum amounts
	db.Model(&models.Payment{}).Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount)
	db.Model(&models.Payment{}).Where("status = ?", "COMPLETED").Select("COALESCE(SUM(amount), 0)").Scan(&completedAmount)
	db.Model(&models.Payment{}).Where("status = ?", "PENDING").Select("COALESCE(SUM(amount), 0)").Scan(&pendingAmount)

	fmt.Printf("Total Payments      : %d (Rp %.2f)\n", totalPayments, totalAmount)
	fmt.Printf("├─ Completed        : %d (%.1f%%) - Rp %.2f\n", 
		completedPayments, float64(completedPayments)/float64(totalPayments)*100, completedAmount)
	fmt.Printf("├─ Pending          : %d (%.1f%%) - Rp %.2f", 
		pendingPayments, float64(pendingPayments)/float64(totalPayments)*100, pendingAmount)
	
	if pendingPayments > 0 {
		fmt.Printf(" ⚠️\n")
		// Show pending payments details
		var pendingDetails []struct {
			Code string
			Amount float64
			CreatedAt time.Time
		}
		db.Model(&models.Payment{}).Select("code, amount, created_at").
			Where("status = ?", "PENDING").Find(&pendingDetails)
		
		for _, p := range pendingDetails {
			age := time.Since(p.CreatedAt)
			fmt.Printf("    └─ %s: Rp %.2f (Age: %.0f hours)\n", 
				p.Code, p.Amount, age.Hours())
		}
	} else {
		fmt.Printf(" ✅\n")
	}
	
	fmt.Printf("└─ Failed           : %d (%.1f%%)\n", 
		failedPayments, float64(failedPayments)/float64(totalPayments)*100)
}

func checkJournalIntegrity(db *gorm.DB) {
	fmt.Println("\n📋 JOURNAL ENTRY INTEGRITY")
	fmt.Println("---------------------------")

	var paymentsWithJournals, paymentsWithoutJournals int64
	
	// Payments with journal entries
	db.Raw(`
		SELECT COUNT(DISTINCT p.id) 
		FROM payments p 
		INNER JOIN journal_entries je ON je.reference_type = 'PAYMENT' AND je.reference_id = p.id
		WHERE p.status = 'COMPLETED'
	`).Scan(&paymentsWithJournals)

	// Payments without journal entries
	db.Raw(`
		SELECT COUNT(*) 
		FROM payments p 
		LEFT JOIN journal_entries je ON je.reference_type = 'PAYMENT' AND je.reference_id = p.id
		WHERE p.status = 'COMPLETED' AND je.id IS NULL
	`).Scan(&paymentsWithoutJournals)

	var completedPayments int64
	db.Model(&models.Payment{}).Where("status = ?", "COMPLETED").Count(&completedPayments)

	fmt.Printf("Completed Payments  : %d\n", completedPayments)
	fmt.Printf("├─ With Journals    : %d (%.1f%%)", 
		paymentsWithJournals, float64(paymentsWithJournals)/float64(completedPayments)*100)
	
	if paymentsWithJournals == completedPayments {
		fmt.Printf(" ✅\n")
	} else {
		fmt.Printf(" ⚠️\n")
	}
	
	fmt.Printf("└─ Missing Journals : %d", paymentsWithoutJournals)
	if paymentsWithoutJournals > 0 {
		fmt.Printf(" ❌\n")
	} else {
		fmt.Printf(" ✅\n")
	}
}

func checkCashBankHealth(db *gorm.DB) {
	fmt.Println("\n💰 CASH/BANK ACCOUNT HEALTH")
	fmt.Println("-----------------------------")

	var cashBankAccounts []models.CashBank
	db.Where("is_active = ?", true).Find(&cashBankAccounts)

	var totalBalance float64
	negativeCount := 0
	lowBalanceCount := 0
	const lowBalanceThreshold = 1000000 // Rp 1 juta

	fmt.Printf("Active Accounts     : %d\n", len(cashBankAccounts))
	
	for _, account := range cashBankAccounts {
		totalBalance += account.Balance
		status := "✅"
		
		if account.Balance < 0 {
			negativeCount++
			status = "❌ NEGATIVE"
		} else if account.Balance < lowBalanceThreshold {
			lowBalanceCount++
			status = "⚠️ LOW"
		}
		
		fmt.Printf("├─ %-20s: Rp %12.2f %s\n", 
			account.Name[:min(20, len(account.Name))], account.Balance, status)
	}
	
	fmt.Printf("└─ Total Balance    : Rp %12.2f\n", totalBalance)

	if negativeCount > 0 {
		fmt.Printf("⚠️ WARNING: %d accounts with negative balance\n", negativeCount)
	}
	if lowBalanceCount > 0 {
		fmt.Printf("⚠️ WARNING: %d accounts with low balance\n", lowBalanceCount)
	}
}

func checkRecentActivity(db *gorm.DB) {
	fmt.Println("\n🕒 RECENT ACTIVITY (Last 24 Hours)")
	fmt.Println("-----------------------------------")

	yesterday := time.Now().AddDate(0, 0, -1)

	var recentPayments int64
	var recentAmount float64
	
	db.Model(&models.Payment{}).Where("created_at >= ?", yesterday).Count(&recentPayments)
	db.Model(&models.Payment{}).Where("created_at >= ?", yesterday).
		Select("COALESCE(SUM(amount), 0)").Scan(&recentAmount)

	fmt.Printf("New Payments        : %d\n", recentPayments)
	fmt.Printf("Total Amount        : Rp %.2f\n", recentAmount)

	// Recent completed payments
	var recentCompleted []struct {
		Code string
		Amount float64
		ContactName string
		CreatedAt time.Time
	}
	
	db.Raw(`
		SELECT p.code, p.amount, c.name as contact_name, p.created_at 
		FROM payments p 
		LEFT JOIN contacts c ON p.contact_id = c.id 
		WHERE p.created_at >= ? AND p.status = 'COMPLETED'
		ORDER BY p.created_at DESC 
		LIMIT 5
	`, yesterday).Scan(&recentCompleted)

	if len(recentCompleted) > 0 {
		fmt.Println("Recent Completions  :")
		for i, p := range recentCompleted {
			fmt.Printf("  %d. %-15s Rp %10.2f %s (%s)\n", 
				i+1, p.Code, p.Amount, p.ContactName, 
				p.CreatedAt.Format("15:04"))
		}
	} else {
		fmt.Println("Recent Completions  : None")
	}
}

func calculateHealthScore(db *gorm.DB) {
	fmt.Println("\n🏥 SYSTEM HEALTH SCORE")
	fmt.Println("----------------------")

	score := 100.0
	issues := []string{}

	// Check pending payments
	var pendingCount int64
	db.Model(&models.Payment{}).Where("status = ?", "PENDING").Count(&pendingCount)
	if pendingCount > 0 {
		score -= 20
		issues = append(issues, fmt.Sprintf("%d pending payments", pendingCount))
	}

	// Check journal integrity
	var paymentsWithoutJournals int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM payments p 
		LEFT JOIN journal_entries je ON je.reference_type = 'PAYMENT' AND je.reference_id = p.id
		WHERE p.status = 'COMPLETED' AND je.id IS NULL
	`).Scan(&paymentsWithoutJournals)
	
	if paymentsWithoutJournals > 0 {
		score -= 25
		issues = append(issues, fmt.Sprintf("%d payments missing journals", paymentsWithoutJournals))
	}

	// Check negative balances
	var negativeBalanceCount int64
	db.Model(&models.CashBank{}).Where("balance < 0 AND is_active = ?", true).Count(&negativeBalanceCount)
	if negativeBalanceCount > 0 {
		score -= 15
		issues = append(issues, fmt.Sprintf("%d accounts with negative balance", negativeBalanceCount))
	}

	// Check old pending payments (>24 hours)
	yesterday := time.Now().AddDate(0, 0, -1)
	var oldPendingCount int64
	db.Model(&models.Payment{}).Where("status = ? AND created_at < ?", "PENDING", yesterday).Count(&oldPendingCount)
	if oldPendingCount > 0 {
		score -= 20
		issues = append(issues, fmt.Sprintf("%d old pending payments", oldPendingCount))
	}

	// Display score
	fmt.Printf("Health Score        : %.0f/100", score)
	
	if score >= 95 {
		fmt.Printf(" 🟢 EXCELLENT\n")
	} else if score >= 80 {
		fmt.Printf(" 🟡 GOOD\n")
	} else if score >= 60 {
		fmt.Printf(" 🟠 WARNING\n")
	} else {
		fmt.Printf(" 🔴 CRITICAL\n")
	}

	if len(issues) > 0 {
		fmt.Println("Issues Found        :")
		for i, issue := range issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	} else {
		fmt.Println("Issues Found        : None ✅")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}