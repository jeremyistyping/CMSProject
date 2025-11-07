package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// PaymentTimeoutDiagnostics helps diagnose payment timeout issues
type PaymentTimeoutDiagnostics struct {
	db *gorm.DB
}

// NewPaymentTimeoutDiagnostics creates a new diagnostics instance
func NewPaymentTimeoutDiagnostics(db *gorm.DB) *PaymentTimeoutDiagnostics {
	return &PaymentTimeoutDiagnostics{db: db}
}

// DiagnosticResult contains diagnostic information
type DiagnosticResult struct {
	TestName        string        `json:"test_name"`
	Success         bool          `json:"success"`
	Duration        time.Duration `json:"duration"`
	Message         string        `json:"message"`
	Details         interface{}   `json:"details,omitempty"`
}

// DiagnosticReport contains full diagnostic report
type DiagnosticReport struct {
	Timestamp   time.Time          `json:"timestamp"`
	TotalTime   time.Duration      `json:"total_time"`
	Tests       []DiagnosticResult `json:"tests"`
	Summary     string             `json:"summary"`
	Passed      int                `json:"passed"`
	Failed      int                `json:"failed"`
}

// RunFullDiagnostics runs comprehensive payment timeout diagnostics
func (d *PaymentTimeoutDiagnostics) RunFullDiagnostics() *DiagnosticReport {
	startTime := time.Now()
	log.Println("🔍 Starting Payment Timeout Diagnostics...")

	report := &DiagnosticReport{
		Timestamp: startTime,
		Tests:     make([]DiagnosticResult, 0),
	}

	// Test 1: Basic database connectivity
	report.Tests = append(report.Tests, d.testDatabaseConnectivity())
	
	// Test 2: Query performance for sales table
	report.Tests = append(report.Tests, d.testSalesQueryPerformance())
	
	// Test 3: Query performance for payments table
	report.Tests = append(report.Tests, d.testPaymentQueryPerformance())
	
	// Test 4: Cash bank query performance
	report.Tests = append(report.Tests, d.testCashBankQueryPerformance())
	
	// Test 5: Journal system performance
	report.Tests = append(report.Tests, d.testJournalSystemPerformance())
	
	// Test 6: Index effectiveness
	report.Tests = append(report.Tests, d.testIndexEffectiveness())
	
	// Test 7: Connection pool status
	report.Tests = append(report.Tests, d.testConnectionPoolStatus())
	
	// Test 8: System resource usage
	report.Tests = append(report.Tests, d.testSystemResources())

	// Calculate summary
	for _, test := range report.Tests {
		if test.Success {
			report.Passed++
		} else {
			report.Failed++
		}
	}

	report.TotalTime = time.Since(startTime)
	
	if report.Failed == 0 {
		report.Summary = "✅ All diagnostic tests passed"
	} else {
		report.Summary = fmt.Sprintf("⚠️ %d test(s) failed, %d passed", report.Failed, report.Passed)
	}

	log.Printf("🎯 Payment Diagnostics completed in %v: %s", report.TotalTime, report.Summary)
	return report
}

// testDatabaseConnectivity tests basic database connectivity
func (d *PaymentTimeoutDiagnostics) testDatabaseConnectivity() DiagnosticResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result int
	err := d.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	
	duration := time.Since(start)
	
	if err != nil {
		return DiagnosticResult{
			TestName: "Database Connectivity",
			Success:  false,
			Duration: duration,
			Message:  fmt.Sprintf("Failed to connect: %v", err),
		}
	}

	return DiagnosticResult{
		TestName: "Database Connectivity",
		Success:  true,
		Duration: duration,
		Message:  "Database connection successful",
	}
}

// testSalesQueryPerformance tests sales table query performance
func (d *PaymentTimeoutDiagnostics) testSalesQueryPerformance() DiagnosticResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var count int64
	err := d.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM sales 
		WHERE status IN ('INVOICED', 'OVERDUE') 
		AND outstanding_amount > 0
	`).Scan(&count).Error
	
	duration := time.Since(start)
	
	if err != nil {
		return DiagnosticResult{
			TestName: "Sales Query Performance",
			Success:  false,
			Duration: duration,
			Message:  fmt.Sprintf("Sales query failed: %v", err),
		}
	}

	success := duration < 2*time.Second
	message := fmt.Sprintf("Found %d pending sales in %v", count, duration)
	if !success {
		message += " (SLOW - consider adding indexes)"
	}

	return DiagnosticResult{
		TestName: "Sales Query Performance",
		Success:  success,
		Duration: duration,
		Message:  message,
		Details:  map[string]interface{}{"pending_sales": count},
	}
}

// testPaymentQueryPerformance tests payment table query performance
func (d *PaymentTimeoutDiagnostics) testPaymentQueryPerformance() DiagnosticResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var count int64
	err := d.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM payments 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
	`).Scan(&count).Error
	
	duration := time.Since(start)
	
	if err != nil {
		return DiagnosticResult{
			TestName: "Payment Query Performance",
			Success:  false,
			Duration: duration,
			Message:  fmt.Sprintf("Payment query failed: %v", err),
		}
	}

	success := duration < 1*time.Second
	message := fmt.Sprintf("Found %d recent payments in %v", count, duration)
	if !success {
		message += " (SLOW - consider adding indexes)"
	}

	return DiagnosticResult{
		TestName: "Payment Query Performance",
		Success:  success,
		Duration: duration,
		Message:  message,
		Details:  map[string]interface{}{"recent_payments": count},
	}
}

// testCashBankQueryPerformance tests cash bank query performance
func (d *PaymentTimeoutDiagnostics) testCashBankQueryPerformance() DiagnosticResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	err := d.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM cash_banks 
		WHERE is_active = true
	`).Scan(&count).Error
	
	duration := time.Since(start)
	
	if err != nil {
		return DiagnosticResult{
			TestName: "Cash Bank Query Performance",
			Success:  false,
			Duration: duration,
			Message:  fmt.Sprintf("Cash bank query failed: %v", err),
		}
	}

	success := duration < 500*time.Millisecond
	message := fmt.Sprintf("Found %d active cash banks in %v", count, duration)
	if !success {
		message += " (SLOW)"
	}

	return DiagnosticResult{
		TestName: "Cash Bank Query Performance",
		Success:  success,
		Duration: duration,
		Message:  message,
		Details:  map[string]interface{}{"active_cash_banks": count},
	}
}

// testJournalSystemPerformance tests journal system performance
func (d *PaymentTimeoutDiagnostics) testJournalSystemPerformance() DiagnosticResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var count int64
	err := d.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM ssot_journal_entries 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
	`).Scan(&count).Error
	
	duration := time.Since(start)
	
	if err != nil {
		return DiagnosticResult{
			TestName: "Journal System Performance",
			Success:  false,
			Duration: duration,
			Message:  fmt.Sprintf("Journal query failed: %v", err),
		}
	}

	success := duration < 3*time.Second
	message := fmt.Sprintf("Found %d journal entries (7 days) in %v", count, duration)
	if !success {
		message += " (VERY SLOW - this may be causing timeouts)"
	}

	return DiagnosticResult{
		TestName: "Journal System Performance",
		Success:  success,
		Duration: duration,
		Message:  message,
		Details:  map[string]interface{}{"journal_entries": count},
	}
}

// testIndexEffectiveness tests if indexes are being used effectively
func (d *PaymentTimeoutDiagnostics) testIndexEffectiveness() DiagnosticResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test if the sales status index exists and is effective
	var indexCount int64
	err := d.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		AND table_name = 'sales' 
		AND column_name = 'status'
	`).Scan(&indexCount).Error
	
	duration := time.Since(start)
	
	if err != nil {
		return DiagnosticResult{
			TestName: "Index Effectiveness",
			Success:  false,
			Duration: duration,
			Message:  fmt.Sprintf("Index check failed: %v", err),
		}
	}

	success := indexCount > 0
	message := fmt.Sprintf("Sales status index check completed in %v", duration)
	if !success {
		message += " (Missing important indexes - this WILL cause timeouts)"
	} else {
		message += " (Indexes found)"
	}

	return DiagnosticResult{
		TestName: "Index Effectiveness",
		Success:  success,
		Duration: duration,
		Message:  message,
		Details:  map[string]interface{}{"sales_status_indexes": indexCount},
	}
}

// testConnectionPoolStatus tests database connection pool status
func (d *PaymentTimeoutDiagnostics) testConnectionPoolStatus() DiagnosticResult {
	start := time.Now()
	
	sqlDB, err := d.db.DB()
	if err != nil {
		return DiagnosticResult{
			TestName: "Connection Pool Status",
			Success:  false,
			Duration: time.Since(start),
			Message:  fmt.Sprintf("Failed to get SQL DB: %v", err),
		}
	}

	stats := sqlDB.Stats()
	duration := time.Since(start)

	// Check if connection pool is healthy
	success := stats.OpenConnections <= 20 && stats.OpenConnections > 0
	message := fmt.Sprintf("Pool: %d open, %d idle, %d in use", 
		stats.OpenConnections, stats.Idle, stats.InUse)
	
	if !success {
		if stats.OpenConnections > 20 {
			message += " (HIGH CONNECTION COUNT - may cause timeouts)"
		} else {
			message += " (No active connections)"
		}
	}

	return DiagnosticResult{
		TestName: "Connection Pool Status",
		Success:  success,
		Duration: duration,
		Message:  message,
		Details:  map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"idle":            stats.Idle,
			"in_use":          stats.InUse,
			"wait_count":      stats.WaitCount,
		},
	}
}

// testSystemResources tests system resource availability
func (d *PaymentTimeoutDiagnostics) testSystemResources() DiagnosticResult {
	start := time.Now()
	
	// Simple test - measure response time for a basic operation
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var dummy int
	err := d.db.WithContext(ctx).Raw("SELECT 1 + 1").Scan(&dummy).Error
	duration := time.Since(start)

	if err != nil {
		return DiagnosticResult{
			TestName: "System Resources",
			Success:  false,
			Duration: duration,
			Message:  fmt.Sprintf("System resource test failed: %v", err),
		}
	}

	success := duration < 100*time.Millisecond
	message := fmt.Sprintf("Basic operation completed in %v", duration)
	if !success {
		message += " (SYSTEM MAY BE OVERLOADED)"
	}

	return DiagnosticResult{
		TestName: "System Resources",
		Success:  success,
		Duration: duration,
		Message:  message,
	}
}

// GetQuickHealthCheck provides a quick health check for payment processing
func (d *PaymentTimeoutDiagnostics) GetQuickHealthCheck() map[string]interface{} {
	start := time.Now()
	
	result := map[string]interface{}{
		"timestamp": start,
		"status":    "checking",
	}

	// Quick database ping
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := d.db.WithContext(ctx).Raw("SELECT 1").Error
	pingDuration := time.Since(start)

	if err != nil {
		result["status"] = "unhealthy"
		result["error"] = err.Error()
		result["ping_duration"] = pingDuration.String()
		return result
	}

	if pingDuration > 1*time.Second {
		result["status"] = "slow"
		result["warning"] = "Database response is slow"
	} else {
		result["status"] = "healthy"
	}

	result["ping_duration"] = pingDuration.String()
	result["total_check_time"] = time.Since(start).String()
	
	return result
}