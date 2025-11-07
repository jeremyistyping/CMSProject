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
	log.Println("=== DEEP PAYMENT PROCESSING ANALYSIS ===")
	
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	fmt.Printf("Analysis started at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("=====================================================")

	// 1. Analyze transaction completeness
	analyzeTransactionCompleteness(db)

	// 2. Identify failure patterns
	identifyFailurePatterns(db)

	// 3. Check data consistency
	checkDataConsistency(db)

	// 4. Analyze timing issues
	analyzeTimingIssues(db)

	// 5. Resource availability check
	checkResourceAvailability(db)

	// 6. Generate recommendations
	generateRecommendations(db)

	fmt.Println("=====================================================")
	fmt.Println("✅ Deep analysis complete")
}

func analyzeTransactionCompleteness(db *gorm.DB) {
	fmt.Println("\n🔍 TRANSACTION COMPLETENESS ANALYSIS")
	fmt.Println("====================================")

	// Get all payments and check their completion status
	var payments []models.Payment
	db.Find(&payments)

	fmt.Printf("Total Payments Analyzed: %d\n", len(payments))

	type CompletionStatus struct {
		HasPaymentRecord    bool
		HasAllocations     bool
		HasJournalEntry    bool
		HasCashBankTx      bool
		IsStatusCompleted  bool
		CompletionScore    float64
	}

	var completionStats []CompletionStatus
	var totalScore float64

	for _, payment := range payments {
		status := CompletionStatus{
			HasPaymentRecord:  true, // Obviously true if we found it
			IsStatusCompleted: payment.Status == "COMPLETED",
		}

		// Check allocations
		var allocCount int64
		db.Model(&models.PaymentAllocation{}).Where("payment_id = ?", payment.ID).Count(&allocCount)
		status.HasAllocations = allocCount > 0

		// Check journal entries
		var journalCount int64
		db.Model(&models.JournalEntry{}).Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).Count(&journalCount)
		status.HasJournalEntry = journalCount > 0

		// Check cash bank transactions
		var cashBankCount int64
		db.Model(&models.CashBankTransaction{}).Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).Count(&cashBankCount)
		status.HasCashBankTx = cashBankCount > 0

		// Calculate completion score
		score := 0.0
		if status.HasPaymentRecord { score += 20 }
		if status.HasAllocations { score += 20 }
		if status.HasJournalEntry { score += 30 }
		if status.HasCashBankTx { score += 20 }
		if status.IsStatusCompleted { score += 10 }
		
		status.CompletionScore = score
		completionStats = append(completionStats, status)
		totalScore += score

		// Report individual payment status
		statusIcon := "✅"
		if score < 100 {
			statusIcon = "❌"
		} else if score < 90 {
			statusIcon = "⚠️"
		}
		
		fmt.Printf("  Payment %d (%s): %.0f%% complete %s\n", payment.ID, payment.Code, score, statusIcon)
		if score < 100 {
			if !status.HasAllocations { fmt.Printf("    - Missing: Allocations\n") }
			if !status.HasJournalEntry { fmt.Printf("    - Missing: Journal Entry\n") }
			if !status.HasCashBankTx { fmt.Printf("    - Missing: Cash/Bank Transaction\n") }
			if !status.IsStatusCompleted { fmt.Printf("    - Status: %s (not COMPLETED)\n", payment.Status) }
		}
	}

	avgScore := totalScore / float64(len(payments))
	fmt.Printf("\n📊 COMPLETION STATISTICS:\n")
	fmt.Printf("Average Completion Score: %.1f%%\n", avgScore)

	// Count by completion levels
	perfect := 0
	good := 0
	poor := 0
	for _, status := range completionStats {
		if status.CompletionScore == 100 {
			perfect++
		} else if status.CompletionScore >= 80 {
			good++
		} else {
			poor++
		}
	}

	fmt.Printf("Perfect (100%%): %d (%.1f%%)\n", perfect, float64(perfect)/float64(len(payments))*100)
	fmt.Printf("Good (80-99%%): %d (%.1f%%)\n", good, float64(good)/float64(len(payments))*100)
	fmt.Printf("Poor (<80%%): %d (%.1f%%)\n", poor, float64(poor)/float64(len(payments))*100)
}

func identifyFailurePatterns(db *gorm.DB) {
	fmt.Println("\n🔍 FAILURE PATTERN IDENTIFICATION")
	fmt.Println("=================================")

	// Analyze by payment method
	fmt.Println("By Payment Method:")
	var methodStats []struct {
		Method string
		Total int64
		Completed int64
	}
	
	db.Raw(`
		SELECT 
			method,
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) as completed
		FROM payments 
		GROUP BY method
	`).Scan(&methodStats)

	for _, stat := range methodStats {
		successRate := float64(stat.Completed) / float64(stat.Total) * 100
		fmt.Printf("  %-10s: %d total, %d completed (%.1f%% success)\n", 
			stat.Method, stat.Total, stat.Completed, successRate)
	}

	// Analyze by creation date
	fmt.Println("\nBy Creation Date:")
	var dateStats []struct {
		Date string
		Total int64
		Completed int64
	}
	
	db.Raw(`
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) as completed
		FROM payments 
		GROUP BY DATE(created_at)
		ORDER BY date
	`).Scan(&dateStats)

	for _, stat := range dateStats {
		successRate := float64(stat.Completed) / float64(stat.Total) * 100
		fmt.Printf("  %s: %d total, %d completed (%.1f%% success)\n", 
			stat.Date, stat.Total, stat.Completed, successRate)
	}

	// Check for user-specific patterns
	fmt.Println("\nBy User:")
	var userStats []struct {
		UserID uint
		Username string
		Total int64
		Completed int64
	}
	
	db.Raw(`
		SELECT 
			p.user_id,
			u.username,
			COUNT(*) as total,
			COUNT(CASE WHEN p.status = 'COMPLETED' THEN 1 END) as completed
		FROM payments p
		LEFT JOIN users u ON p.user_id = u.id
		GROUP BY p.user_id, u.username
	`).Scan(&userStats)

	for _, stat := range userStats {
		successRate := float64(stat.Completed) / float64(stat.Total) * 100
		fmt.Printf("  %-15s: %d total, %d completed (%.1f%% success)\n", 
			stat.Username, stat.Total, stat.Completed, successRate)
	}
}

func checkDataConsistency(db *gorm.DB) {
	fmt.Println("\n🔍 DATA CONSISTENCY CHECK")
	fmt.Println("========================")

	// Check for orphaned allocations
	var orphanedAllocations int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM payment_allocations pa 
		LEFT JOIN payments p ON pa.payment_id = p.id 
		WHERE p.id IS NULL
	`).Scan(&orphanedAllocations)
	
	fmt.Printf("Orphaned Allocations: %d", orphanedAllocations)
	if orphanedAllocations > 0 { fmt.Printf(" ❌") } else { fmt.Printf(" ✅") }
	fmt.Println()

	// Check for journal entries without payments
	var orphanedJournals int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM journal_entries je 
		WHERE je.reference_type = 'PAYMENT' 
		AND je.reference_id NOT IN (SELECT id FROM payments)
	`).Scan(&orphanedJournals)
	
	fmt.Printf("Orphaned Journal Entries: %d", orphanedJournals)
	if orphanedJournals > 0 { fmt.Printf(" ❌") } else { fmt.Printf(" ✅") }
	fmt.Println()

	// Check for cash bank transactions without payments
	var orphanedCashBankTx int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_bank_transactions cbt 
		WHERE cbt.reference_type = 'PAYMENT' 
		AND cbt.reference_id NOT IN (SELECT id FROM payments)
	`).Scan(&orphanedCashBankTx)
	
	fmt.Printf("Orphaned Cash/Bank Tx: %d", orphanedCashBankTx)
	if orphanedCashBankTx > 0 { fmt.Printf(" ❌") } else { fmt.Printf(" ✅") }
	fmt.Println()

	// Check for amount mismatches
	var amountMismatches []struct {
		PaymentID uint
		PaymentCode string
		PaymentAmount float64
		AllocatedTotal float64
		Difference float64
	}
	
	db.Raw(`
		SELECT 
			p.id as payment_id,
			p.code as payment_code,
			p.amount as payment_amount,
			COALESCE(SUM(pa.allocated_amount), 0) as allocated_total,
			p.amount - COALESCE(SUM(pa.allocated_amount), 0) as difference
		FROM payments p
		LEFT JOIN payment_allocations pa ON p.id = pa.payment_id
		GROUP BY p.id, p.code, p.amount
		HAVING ABS(p.amount - COALESCE(SUM(pa.allocated_amount), 0)) > 0.01
	`).Scan(&amountMismatches)

	fmt.Printf("Amount Mismatches: %d", len(amountMismatches))
	if len(amountMismatches) > 0 { 
		fmt.Printf(" ❌\n")
		for _, mismatch := range amountMismatches {
			fmt.Printf("  %s: Payment %.2f vs Allocated %.2f (diff: %.2f)\n",
				mismatch.PaymentCode, mismatch.PaymentAmount, 
				mismatch.AllocatedTotal, mismatch.Difference)
		}
	} else { 
		fmt.Printf(" ✅\n") 
	}
}

func analyzeTimingIssues(db *gorm.DB) {
	fmt.Println("\n🔍 TIMING ANALYSIS")
	fmt.Println("==================")

	// Check processing time patterns
	var timingStats []struct {
		PaymentID uint
		Code string
		Status string
		CreatedAt time.Time
		UpdatedAt time.Time
		ProcessingTime float64
	}
	
	db.Raw(`
		SELECT 
			id as payment_id,
			code,
			status,
			created_at,
			updated_at,
			EXTRACT(EPOCH FROM (updated_at - created_at)) as processing_time
		FROM payments
		ORDER BY created_at
	`).Scan(&timingStats)

	var totalProcessingTime float64
	var completedCount int
	var stuckPayments []struct {
		Code string
		Status string
		Age float64
	}

	for _, stat := range timingStats {
		totalProcessingTime += stat.ProcessingTime
		
		if stat.Status == "COMPLETED" {
			completedCount++
		} else {
			age := time.Since(stat.CreatedAt).Hours()
			if age > 24 { // Consider stuck if pending > 24 hours
				stuckPayments = append(stuckPayments, struct {
					Code string
					Status string
					Age float64
				}{stat.Code, stat.Status, age})
			}
		}
	}

	if completedCount > 0 {
		avgProcessingTime := totalProcessingTime / float64(completedCount)
		fmt.Printf("Average Processing Time: %.2f seconds\n", avgProcessingTime)
	}

	fmt.Printf("Stuck Payments (>24h): %d\n", len(stuckPayments))
	for _, stuck := range stuckPayments {
		fmt.Printf("  %s (%s): %.1f hours old\n", stuck.Code, stuck.Status, stuck.Age)
	}
}

func checkResourceAvailability(db *gorm.DB) {
	fmt.Println("\n🔍 RESOURCE AVAILABILITY")
	fmt.Println("========================")

	// Check required accounts existence
	requiredAccounts := []string{"1201", "2101", "1101", "1104"} // AR, AP, Cash, Bank
	fmt.Println("Required Accounts:")
	for _, code := range requiredAccounts {
		var count int64
		db.Model(&models.Account{}).Where("code = ?", code).Count(&count)
		status := "✅"
		if count == 0 { status = "❌ MISSING" }
		fmt.Printf("  Account %s: %s\n", code, status)
	}

	// Check active cash/bank accounts
	var activeCashBanks int64
	db.Model(&models.CashBank{}).Where("is_active = ?", true).Count(&activeCashBanks)
	fmt.Printf("Active Cash/Bank Accounts: %d\n", activeCashBanks)

	// Check for negative balances that could cause failures
	var negativeBalances []struct {
		ID uint
		Name string
		Balance float64
	}
	db.Model(&models.CashBank{}).Select("id, name, balance").
		Where("balance < 0 AND is_active = ?", true).Find(&negativeBalances)

	fmt.Printf("Negative Balance Accounts: %d\n", len(negativeBalances))
	for _, acc := range negativeBalances {
		fmt.Printf("  %s: Rp %.2f ❌\n", acc.Name, acc.Balance)
	}
}

func generateRecommendations(db *gorm.DB) {
	fmt.Println("\n🎯 RECOMMENDATIONS FOR 100% SUCCESS RATE")
	fmt.Println("========================================")

	recommendations := []string{
		"1. IMMEDIATE FIXES:",
		"   • Fix all existing PENDING payments using repair script",
		"   • Resolve negative account balances",
		"   • Clean up any orphaned data",
		"",
		"2. TRANSACTION ROBUSTNESS:",
		"   • Implement database transaction with proper rollback",
		"   • Add retry mechanism for failed operations",
		"   • Use distributed transaction patterns if needed",
		"",
		"3. VALIDATION ENHANCEMENTS:",
		"   • Pre-validate all required accounts exist",
		"   • Check sufficient balance before processing",
		"   • Validate allocation amounts match payment amount",
		"",
		"4. ERROR HANDLING:",
		"   • Comprehensive error logging with context",
		"   • Graceful degradation for non-critical failures",
		"   • Dead letter queue for failed payments",
		"",
		"5. MONITORING & ALERTING:",
		"   • Real-time monitoring for stuck payments",
		"   • Daily health checks and reporting",
		"   • Automatic alerts for failures",
		"",
		"6. PROCESS IMPROVEMENTS:",
		"   • Async processing for non-critical operations",
		"   • Idempotent operations to prevent duplicates",
		"   • Circuit breaker pattern for external dependencies",
	}

	for _, rec := range recommendations {
		fmt.Println(rec)
	}
}

