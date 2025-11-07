package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

type PerformanceMetrics struct {
	Timestamp                time.Time              `json:"timestamp"`
	DatabaseMetrics         DatabaseMetrics        `json:"database_metrics"`
	SSOTMetrics             SSOTMetrics           `json:"ssot_metrics"`
	SystemMetrics           SystemMetrics         `json:"system_metrics"`
	QueryPerformanceMetrics map[string]QueryMetrics `json:"query_performance_metrics"`
	MaterializedViewMetrics MaterializedViewMetrics `json:"materialized_view_metrics"`
	AlertsTriggered         []Alert               `json:"alerts_triggered"`
}

type DatabaseMetrics struct {
	ActiveConnections   int           `json:"active_connections"`
	IdleConnections     int           `json:"idle_connections"`
	TotalConnections    int           `json:"total_connections"`
	QueryCount          int64         `json:"query_count"`
	SlowQueryCount      int64         `json:"slow_query_count"`
	AverageQueryTime    time.Duration `json:"average_query_time"`
	DatabaseSize        string        `json:"database_size"`
	IndexEfficiency     float64       `json:"index_efficiency"`
}

type SSOTMetrics struct {
	TotalJournalEntries    int64         `json:"total_journal_entries"`
	TotalJournalLines      int64         `json:"total_journal_lines"`
	TotalEventLogs         int64         `json:"total_event_logs"`
	PostedEntries          int64         `json:"posted_entries"`
	DraftEntries           int64         `json:"draft_entries"`
	ReversedEntries        int64         `json:"reversed_entries"`
	EntriesLast24h         int64         `json:"entries_last_24h"`
	BalanceAccuracy        float64       `json:"balance_accuracy"`
	DataIntegrityScore     float64       `json:"data_integrity_score"`
}

type SystemMetrics struct {
	CPUUsage     float64       `json:"cpu_usage"`
	MemoryUsage  float64       `json:"memory_usage"`
	DiskUsage    float64       `json:"disk_usage"`
	GoroutineCount int         `json:"goroutine_count"`
	HeapSize     uint64        `json:"heap_size_mb"`
	GCCount      uint32        `json:"gc_count"`
	Uptime       time.Duration  `json:"uptime_seconds"`
}

type QueryMetrics struct {
	QueryType          string        `json:"query_type"`
	AverageExecutionTime time.Duration `json:"average_execution_time_ms"`
	MaxExecutionTime   time.Duration `json:"max_execution_time_ms"`
	MinExecutionTime   time.Duration `json:"min_execution_time_ms"`
	TotalExecutions    int64         `json:"total_executions"`
	FailureRate        float64       `json:"failure_rate"`
}

type MaterializedViewMetrics struct {
	AccountBalancesLastRefresh time.Time     `json:"account_balances_last_refresh"`
	RefreshFrequency          time.Duration `json:"refresh_frequency"`
	RefreshDuration           time.Duration `json:"last_refresh_duration"`
	ViewSize                  string        `json:"view_size"`
	IsStale                   bool          `json:"is_stale"`
}

type Alert struct {
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	Metric      string    `json:"metric"`
	Threshold   float64   `json:"threshold"`
	ActualValue float64   `json:"actual_value"`
	Timestamp   time.Time `json:"timestamp"`
}

var startTime = time.Now()

func main() {
	fmt.Println("📊 SSOT Performance Monitoring System")
	fmt.Println("====================================")
	fmt.Println("Monitoring system performance with real-time metrics\n")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	// Start continuous monitoring
	monitoringInterval := 30 * time.Second
	reportInterval := 5 * time.Minute

	fmt.Printf("🔄 Starting monitoring (interval: %v, reporting: %v)\n", monitoringInterval, reportInterval)
	fmt.Println("Press Ctrl+C to stop monitoring")
	fmt.Println("")

	lastReport := time.Now()

	for {
		metrics := collectMetrics(db)
		
		// Save metrics to file
		saveMetrics(metrics)

		// Print summary
		if time.Since(lastReport) >= reportInterval {
			printMetricsSummary(metrics)
			lastReport = time.Now()
		}

		// Check for alerts
		alerts := checkAlerts(metrics)
		if len(alerts) > 0 {
			fmt.Printf("\n🚨 ALERTS TRIGGERED (%d):\n", len(alerts))
			for _, alert := range alerts {
				color := "Yellow"
				if alert.Severity == "CRITICAL" {
					color = "Red"
				}
				fmt.Printf("  %s [%s] %s (%.2f > %.2f)\n", 
					getColorCode(color), alert.Severity, alert.Message, alert.ActualValue, alert.Threshold)
			}
			fmt.Println("")
		}

		// Sleep until next monitoring cycle
		time.Sleep(monitoringInterval)
	}
}

func collectMetrics(db *gorm.DB) PerformanceMetrics {
	metrics := PerformanceMetrics{
		Timestamp: time.Now(),
		QueryPerformanceMetrics: make(map[string]QueryMetrics),
	}

	// Collect database metrics
	metrics.DatabaseMetrics = collectDatabaseMetrics(db)

	// Collect SSOT-specific metrics
	metrics.SSOTMetrics = collectSSOTMetrics(db)

	// Collect system metrics
	metrics.SystemMetrics = collectSystemMetrics()

	// Collect query performance metrics
	metrics.QueryPerformanceMetrics = collectQueryPerformanceMetrics(db)

	// Collect materialized view metrics
	metrics.MaterializedViewMetrics = collectMaterializedViewMetrics(db)

	// Check for alerts
	metrics.AlertsTriggered = checkAlerts(metrics)

	return metrics
}

func collectDatabaseMetrics(db *gorm.DB) DatabaseMetrics {
	sqlDB, _ := db.DB()
	stats := sqlDB.Stats()

	var dbSize string
	db.Raw("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize)

	// Calculate index efficiency (simplified)
	var indexHitRatio float64
	db.Raw(`
		SELECT COALESCE((
			SUM(idx_blks_hit) * 100.0 / NULLIF(SUM(idx_blks_hit + idx_blks_read), 0)
		), 0) as ratio
		FROM pg_stat_user_indexes
	`).Scan(&indexHitRatio)

	return DatabaseMetrics{
		ActiveConnections: stats.OpenConnections,
		IdleConnections:   stats.Idle,
		TotalConnections:  stats.OpenConnections,
		DatabaseSize:      dbSize,
		IndexEfficiency:   indexHitRatio,
	}
}

func collectSSOTMetrics(db *gorm.DB) SSOTMetrics {
	var metrics SSOTMetrics

	// Count journal entries by status
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&metrics.TotalJournalEntries)
	db.Raw("SELECT COUNT(*) FROM unified_journal_lines").Scan(&metrics.TotalJournalLines)
	db.Raw("SELECT COUNT(*) FROM journal_event_log").Scan(&metrics.TotalEventLogs)
	
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger WHERE status = 'POSTED'").Scan(&metrics.PostedEntries)
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger WHERE status = 'DRAFT'").Scan(&metrics.DraftEntries)
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger WHERE status = 'REVERSED'").Scan(&metrics.ReversedEntries)
	
	// Entries in last 24 hours
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger WHERE created_at >= NOW() - INTERVAL '24 hours'").Scan(&metrics.EntriesLast24h)

	// Calculate balance accuracy (all entries should be balanced)
	var unbalancedEntries int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger WHERE NOT is_balanced").Scan(&unbalancedEntries)
	
	if metrics.TotalJournalEntries > 0 {
		metrics.BalanceAccuracy = float64(metrics.TotalJournalEntries-unbalancedEntries) / float64(metrics.TotalJournalEntries) * 100
	}

	// Data integrity score (simplified check)
	var integrityIssues int64
	db.Raw(`
		SELECT COUNT(*) FROM unified_journal_ledger j
		LEFT JOIN unified_journal_lines l ON j.id = l.journal_id
		WHERE l.journal_id IS NULL
	`).Scan(&integrityIssues)
	
	if metrics.TotalJournalEntries > 0 {
		metrics.DataIntegrityScore = float64(metrics.TotalJournalEntries-integrityIssues) / float64(metrics.TotalJournalEntries) * 100
	}

	return metrics
}

func collectSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemMetrics{
		GoroutineCount: runtime.NumGoroutine(),
		HeapSize:       m.HeapAlloc / 1024 / 1024, // Convert to MB
		GCCount:        uint32(m.NumGC),
		Uptime:         time.Since(startTime),
	}
}

func collectQueryPerformanceMetrics(db *gorm.DB) map[string]QueryMetrics {
	metrics := make(map[string]QueryMetrics)

	// Benchmark key SSOT queries
	queries := map[string]string{
		"get_journals": "SELECT * FROM unified_journal_ledger LIMIT 100",
		"get_lines":    "SELECT * FROM unified_journal_lines LIMIT 100",
		"get_balances": "SELECT * FROM account_balances LIMIT 100",
		"count_entries": "SELECT COUNT(*) FROM unified_journal_ledger",
	}

	for queryType, sql := range queries {
		start := time.Now()
		
		// Handle different query types properly
		var result interface{}
		var err error
		
		if queryType == "count_entries" {
			var count int64
			err = db.Raw(sql).Scan(&count)
			result = count
		} else {
			err = db.Raw(sql).Find(&result).Error
		}
		
		duration := time.Since(start)

		metrics[queryType] = QueryMetrics{
			QueryType:             queryType,
			AverageExecutionTime:  duration,
			MaxExecutionTime:      duration,
			MinExecutionTime:      duration,
			TotalExecutions:       1,
			FailureRate:          func() float64 { if err != nil { return 100.0 } else { return 0.0 } }(),
		}
	}

	return metrics
}

func collectMaterializedViewMetrics(db *gorm.DB) MaterializedViewMetrics {
	var metrics MaterializedViewMetrics

	// Check if materialized view exists and get stats
	var viewExists bool
	db.Raw("SELECT EXISTS (SELECT 1 FROM pg_matviews WHERE matviewname = 'account_balances')").Scan(&viewExists)

	if viewExists {
		var viewSize string
		db.Raw("SELECT pg_size_pretty(pg_total_relation_size('account_balances'))").Scan(&viewSize)
		metrics.ViewSize = viewSize

		// Simple refresh test
		start := time.Now()
		db.Exec("REFRESH MATERIALIZED VIEW account_balances")
		metrics.RefreshDuration = time.Since(start)
		metrics.AccountBalancesLastRefresh = time.Now()
	}

	return metrics
}

func checkAlerts(metrics PerformanceMetrics) []Alert {
	var alerts []Alert

	// Memory usage alert
	if metrics.SystemMetrics.HeapSize > 500 { // 500MB
		alerts = append(alerts, Alert{
			Severity:    "WARNING",
			Message:     "High memory usage detected",
			Metric:      "heap_size_mb",
			Threshold:   500,
			ActualValue: float64(metrics.SystemMetrics.HeapSize),
			Timestamp:   time.Now(),
		})
	}

	// Unbalanced entries alert
	if metrics.SSOTMetrics.BalanceAccuracy < 99.0 {
		alerts = append(alerts, Alert{
			Severity:    "CRITICAL",
			Message:     "Low balance accuracy in journal entries",
			Metric:      "balance_accuracy",
			Threshold:   99.0,
			ActualValue: metrics.SSOTMetrics.BalanceAccuracy,
			Timestamp:   time.Now(),
		})
	}

	// Data integrity alert
	if metrics.SSOTMetrics.DataIntegrityScore < 95.0 {
		alerts = append(alerts, Alert{
			Severity:    "CRITICAL",
			Message:     "Data integrity issues detected",
			Metric:      "data_integrity_score",
			Threshold:   95.0,
			ActualValue: metrics.SSOTMetrics.DataIntegrityScore,
			Timestamp:   time.Now(),
		})
	}

	// Query performance alert
	for queryType, queryMetrics := range metrics.QueryPerformanceMetrics {
		if queryMetrics.AverageExecutionTime > 5*time.Second {
			alerts = append(alerts, Alert{
				Severity:    "WARNING",
				Message:     fmt.Sprintf("Slow query detected: %s", queryType),
				Metric:      "query_execution_time",
				Threshold:   5000, // 5 seconds in milliseconds
				ActualValue: float64(queryMetrics.AverageExecutionTime.Milliseconds()),
				Timestamp:   time.Now(),
			})
		}
	}

	return alerts
}

func saveMetrics(metrics PerformanceMetrics) {
	// Save to JSON file with timestamp
	filename := fmt.Sprintf("ssot_metrics_%s.json", metrics.Timestamp.Format("20060102_1504"))
	
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal metrics: %v", err)
		return
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		log.Printf("Failed to save metrics: %v", err)
	}
}

func printMetricsSummary(metrics PerformanceMetrics) {
	fmt.Printf("📊 Performance Report - %s\n", metrics.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Println("=" + fmt.Sprintf("%80s", "="))
	
	// SSOT Metrics
	fmt.Printf("📝 SSOT Journal System:\n")
	fmt.Printf("   Entries: %d (Posted: %d, Draft: %d, Reversed: %d)\n", 
		metrics.SSOTMetrics.TotalJournalEntries, 
		metrics.SSOTMetrics.PostedEntries,
		metrics.SSOTMetrics.DraftEntries,
		metrics.SSOTMetrics.ReversedEntries)
	fmt.Printf("   Lines: %d | Event Logs: %d\n", 
		metrics.SSOTMetrics.TotalJournalLines, 
		metrics.SSOTMetrics.TotalEventLogs)
	fmt.Printf("   Balance Accuracy: %.2f%% | Data Integrity: %.2f%%\n", 
		metrics.SSOTMetrics.BalanceAccuracy,
		metrics.SSOTMetrics.DataIntegrityScore)
	fmt.Printf("   New Entries (24h): %d\n", metrics.SSOTMetrics.EntriesLast24h)

	// Database Metrics
	fmt.Printf("\n💾 Database Performance:\n")
	fmt.Printf("   Connections: %d active, %d idle\n", 
		metrics.DatabaseMetrics.ActiveConnections,
		metrics.DatabaseMetrics.IdleConnections)
	fmt.Printf("   Size: %s | Index Efficiency: %.2f%%\n", 
		metrics.DatabaseMetrics.DatabaseSize,
		metrics.DatabaseMetrics.IndexEfficiency)

	// System Metrics
	fmt.Printf("\n🖥️  System Resources:\n")
	fmt.Printf("   Memory: %d MB heap | Goroutines: %d\n", 
		metrics.SystemMetrics.HeapSize,
		metrics.SystemMetrics.GoroutineCount)
	fmt.Printf("   Uptime: %s\n", metrics.SystemMetrics.Uptime.String())

	// Query Performance
	fmt.Printf("\n⚡ Query Performance:\n")
	for queryType, queryMetrics := range metrics.QueryPerformanceMetrics {
		status := "✅"
		if queryMetrics.AverageExecutionTime > 1*time.Second {
			status = "⚠️ "
		}
		if queryMetrics.AverageExecutionTime > 5*time.Second {
			status = "❌"
		}
		fmt.Printf("   %s %-15s %6.2fms\n", 
			status, queryType, 
			float64(queryMetrics.AverageExecutionTime.Nanoseconds())/1000000)
	}

	// Materialized View
	fmt.Printf("\n📈 Materialized Views:\n")
	fmt.Printf("   Account Balances: %s (refresh: %.2fms)\n", 
		metrics.MaterializedViewMetrics.ViewSize,
		float64(metrics.MaterializedViewMetrics.RefreshDuration.Nanoseconds())/1000000)

	fmt.Printf("\n" + fmt.Sprintf("%80s", "=") + "\n")
}

func getColorCode(color string) string {
	switch color {
	case "Red":
		return "🔴"
	case "Yellow":
		return "🟡" 
	case "Green":
		return "🟢"
	default:
		return "⚪"
	}
}