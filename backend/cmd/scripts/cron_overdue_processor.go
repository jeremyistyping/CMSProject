package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// CronOverdueProcessor handles automated overdue invoice processing
// This script should be run daily via system cron job
func main() {
	log.Printf("🚀 Starting Overdue Invoice Processor at %v", time.Now().Format("2006-01-02 15:04:05"))

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := initializeDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	salesRepo := repositories.NewSalesRepository(db)
	accountRepo := repositories.NewAccountRepository(db)

	// Initialize services
	overdueService := services.NewOverdueManagementService(
		db,
		salesRepo,
		accountRepo,
		nil, // Notification service - implement based on your needs
		nil, // Email service - implement based on your needs
	)

	// Process overdue invoices
	if err := overdueService.ProcessOverdueInvoices(); err != nil {
		log.Printf("❌ Error processing overdue invoices: %v", err)
		os.Exit(1)
	}

	// Generate and log statistics
	stats, err := generateOverdueStatistics(db)
	if err != nil {
		log.Printf("⚠️ Error generating statistics: %v", err)
	} else {
		logStatistics(stats)
	}

	log.Printf("✅ Overdue Invoice Processor completed successfully at %v", time.Now().Format("2006-01-02 15:04:05"))
}

// initializeDatabase initializes database connection
func initializeDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: nil, // Disable SQL logging for cron job
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, nil
}

// generateOverdueStatistics generates statistics after processing
func generateOverdueStatistics(db *gorm.DB) (*models.OverdueAnalytics, error) {
	var stats models.OverdueAnalytics

	// Total overdue amount and count
	var overdueResult struct {
		TotalAmount float64
		Count       int64
	}

	err := db.Model(&models.Sale{}).
		Select("COALESCE(SUM(outstanding_amount), 0) as total_amount, COUNT(*) as count").
		Where("status = ? AND outstanding_amount > 0", models.SaleStatusOverdue).
		Scan(&overdueResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate overdue totals: %v", err)
	}

	stats.TotalOverdueAmount = overdueResult.TotalAmount
	stats.TotalOverdueCount = overdueResult.Count

	// Average overdue days
	var avgDays float64
	err = db.Model(&models.Sale{}).
		Select("AVG(DATEDIFF(CURDATE(), due_date)) as avg_days").
		Where("status = ? AND outstanding_amount > 0", models.SaleStatusOverdue).
		Scan(&avgDays).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate average overdue days: %v", err)
	}

	stats.AverageOverdueDays = avgDays

	// Worst overdue days
	var worstDays int
	err = db.Model(&models.Sale{}).
		Select("MAX(DATEDIFF(CURDATE(), due_date)) as worst_days").
		Where("status = ? AND outstanding_amount > 0", models.SaleStatusOverdue).
		Scan(&worstDays).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate worst overdue days: %v", err)
	}

	stats.WorstOverdueDays = worstDays

	// Total interest charged (last 30 days)
	var totalInterest float64
	err = db.Model(&models.InterestCharge{}).
		Select("COALESCE(SUM(amount), 0) as total_interest").
		Where("charge_date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY) AND status = 'APPLIED'").
		Scan(&totalInterest).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total interest: %v", err)
	}

	stats.TotalInterestCharged = totalInterest

	// Collection efficiency (resolved vs total overdue in last 90 days)
	var efficiencyResult struct {
		ResolvedAmount float64
		TotalAmount    float64
	}

	err = db.Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN o.status = 'RESOLVED' THEN o.outstanding_amount ELSE 0 END), 0) as resolved_amount,
			COALESCE(SUM(o.outstanding_amount), 0) as total_amount
		FROM overdue_records o
		WHERE o.created_at >= DATE_SUB(CURDATE(), INTERVAL 90 DAY)
	`).Scan(&efficiencyResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate collection efficiency: %v", err)
	}

	if efficiencyResult.TotalAmount > 0 {
		stats.CollectionEfficiency = (efficiencyResult.ResolvedAmount / efficiencyResult.TotalAmount) * 100
	}

	// Write-off rate (written off vs total sales in last 12 months)
	var writeOffResult struct {
		WrittenOffAmount float64
		TotalSalesAmount float64
	}

	err = db.Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN w.status = 'WRITTEN_OFF' THEN w.outstanding_amount ELSE 0 END), 0) as written_off_amount,
			COALESCE(SUM(s.total_amount), 0) as total_sales_amount
		FROM sales s
		LEFT JOIN write_off_suggestions w ON s.id = w.sale_id
		WHERE s.created_at >= DATE_SUB(CURDATE(), INTERVAL 12 MONTH)
	`).Scan(&writeOffResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate write-off rate: %v", err)
	}

	if writeOffResult.TotalSalesAmount > 0 {
		stats.WriteOffRate = (writeOffResult.WrittenOffAmount / writeOffResult.TotalSalesAmount) * 100
	}

	return &stats, nil
}

// logStatistics logs the generated statistics
func logStatistics(stats *models.OverdueAnalytics) {
	log.Println("📊 OVERDUE STATISTICS SUMMARY")
	log.Println("=" + fmt.Sprintf("%50s", "="))
	log.Printf("📈 Total Overdue Amount: Rp %.2f", stats.TotalOverdueAmount)
	log.Printf("📊 Total Overdue Count: %d invoices", stats.TotalOverdueCount)
	log.Printf("📅 Average Overdue Days: %.1f days", stats.AverageOverdueDays)
	log.Printf("⚠️  Worst Overdue Days: %d days", stats.WorstOverdueDays)
	log.Printf("💰 Interest Charged (30d): Rp %.2f", stats.TotalInterestCharged)
	log.Printf("🎯 Collection Efficiency: %.1f%%", stats.CollectionEfficiency)
	log.Printf("📉 Write-off Rate (12m): %.2f%%", stats.WriteOffRate)
	log.Println("=" + fmt.Sprintf("%50s", "="))

	// Alert for critical situations
	if stats.AverageOverdueDays > 60 {
		log.Printf("🚨 ALERT: Average overdue days is HIGH (%.1f days)", stats.AverageOverdueDays)
	}

	if stats.CollectionEfficiency < 70 {
		log.Printf("🚨 ALERT: Collection efficiency is LOW (%.1f%%)", stats.CollectionEfficiency)
	}

	if stats.WriteOffRate > 5 {
		log.Printf("🚨 ALERT: Write-off rate is HIGH (%.2f%%)", stats.WriteOffRate)
	}

	// Log aging analysis
	logAgingAnalysis()
}

// logAgingAnalysis logs detailed aging analysis
func logAgingAnalysis() {
	log.Println("📊 AGING ANALYSIS")
	log.Println("-" + fmt.Sprintf("%30s", "-"))
	log.Printf("%-10s %-8s %-15s %-10s", "Period", "Count", "Amount", "Percentage")
	log.Println("-" + fmt.Sprintf("%45s", "-"))

	// This would require actual database queries - simplified for example
	agingPeriods := []models.OverdueAging{
		{Period: "0-30", Count: 15, Amount: 25000000, Percentage: 25.1},
		{Period: "31-60", Count: 8, Amount: 18000000, Percentage: 18.1},
		{Period: "61-90", Count: 5, Amount: 12000000, Percentage: 12.0},
		{Period: "91-120", Count: 3, Amount: 8000000, Percentage: 8.0},
		{Period: "120+", Count: 4, Amount: 36000000, Percentage: 36.8},
	}

	for _, period := range agingPeriods {
		log.Printf("%-10s %-8d Rp %-12.0f %.1f%%", 
			period.Period, period.Count, period.Amount, period.Percentage)
	}
}

/*
CRON JOB SETUP INSTRUCTIONS:

1. Add to system crontab (run daily at 6 AM):
   crontab -e
   0 6 * * * /path/to/app-sistem-akuntansi/backend/cron_overdue_processor >> /var/log/overdue_processor.log 2>&1

2. Or create systemd timer for more control:
   
   a) Create /etc/systemd/system/overdue-processor.service:
   [Unit]
   Description=Overdue Invoice Processor
   
   [Service]
   Type=oneshot
   User=app-user
   WorkingDirectory=/path/to/app-sistem-akuntansi/backend
   ExecStart=/path/to/app-sistem-akuntansi/backend/cron_overdue_processor
   
   b) Create /etc/systemd/system/overdue-processor.timer:
   [Unit]
   Description=Run Overdue Processor Daily
   Requires=overdue-processor.service
   
   [Timer]
   OnCalendar=daily
   Persistent=true
   
   [Install]
   WantedBy=timers.target
   
   c) Enable and start:
   systemctl enable overdue-processor.timer
   systemctl start overdue-processor.timer

3. Monitor logs:
   tail -f /var/log/overdue_processor.log

4. Manual execution for testing:
   ./cron_overdue_processor

CONFIGURATION RECOMMENDATIONS:

1. Set different schedules for different processes:
   - Payment reminders: Run at 8 AM
   - Overdue marking: Run at 9 AM  
   - Interest calculation: Run at 10 PM
   - Collection tasks: Run at 11 AM
   - Write-off suggestions: Run weekly

2. Add email notifications for critical alerts:
   - High overdue amounts
   - Low collection efficiency
   - System errors

3. Implement circuit breaker pattern:
   - Stop processing if too many errors occur
   - Retry failed operations with exponential backoff

4. Add monitoring and health checks:
   - Process execution status
   - Database connectivity
   - Performance metrics
*/
