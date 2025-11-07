package services

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
	"context"

	"gorm.io/gorm"
)

// PerformanceProfiler helps identify bottlenecks in the application
type PerformanceProfiler struct {
	db *gorm.DB
	metrics map[string]*PerformanceMetric
	mu sync.RWMutex
}

// PerformanceMetric tracks performance data for specific operations
type PerformanceMetric struct {
	Name           string        `json:"name"`
	TotalCalls     int64         `json:"total_calls"`
	TotalDuration  time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	MinDuration    time.Duration `json:"min_duration"`
	MaxDuration    time.Duration `json:"max_duration"`
	ErrorCount     int64         `json:"error_count"`
	LastCalled     time.Time     `json:"last_called"`
}

// ProfilerReport contains comprehensive performance analysis
type ProfilerReport struct {
	SystemInfo       SystemInfo                    `json:"system_info"`
	DatabaseHealth   DatabaseHealth               `json:"database_health"`
	SlowOperations   []PerformanceMetric          `json:"slow_operations"`
	Bottlenecks      []Bottleneck                 `json:"bottlenecks"`
	Recommendations  []string                     `json:"recommendations"`
	Timestamp        time.Time                    `json:"timestamp"`
}

// SystemInfo contains basic system information
type SystemInfo struct {
	GoVersion      string    `json:"go_version"`
	NumCPU         int       `json:"num_cpu"`
	NumGoroutines  int       `json:"num_goroutines"`
	MemoryAlloc    uint64    `json:"memory_alloc_mb"`
	MemoryTotal    uint64    `json:"memory_total_mb"`
	MemorySys      uint64    `json:"memory_sys_mb"`
	GCCycles       uint32    `json:"gc_cycles"`
	Timestamp      time.Time `json:"timestamp"`
}

// DatabaseHealth tracks database performance metrics
type DatabaseHealth struct {
	ConnectionsOpen     int           `json:"connections_open"`
	ConnectionsInUse    int           `json:"connections_in_use"`
	ConnectionsIdle     int           `json:"connections_idle"`
	QueryDuration       time.Duration `json:"avg_query_duration"`
	LongRunningQueries  int           `json:"long_running_queries"`
	DeadlockCount       int64         `json:"deadlock_count"`
	SlowQueryCount      int64         `json:"slow_query_count"`
	DatabaseSize        string        `json:"database_size"`
}

// Bottleneck identifies specific performance issues
type Bottleneck struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // "low", "medium", "high", "critical"
	Impact      string    `json:"impact"`
	Suggestion  string    `json:"suggestion"`
	DetectedAt  time.Time `json:"detected_at"`
}

// NewPerformanceProfiler creates a new performance profiler
func NewPerformanceProfiler(db *gorm.DB) *PerformanceProfiler {
	return &PerformanceProfiler{
		db:      db,
		metrics: make(map[string]*PerformanceMetric),
	}
}

// TrackOperation records performance metrics for an operation
func (pp *PerformanceProfiler) TrackOperation(name string, duration time.Duration, err error) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	metric, exists := pp.metrics[name]
	if !exists {
		metric = &PerformanceMetric{
			Name:        name,
			MinDuration: duration,
			MaxDuration: duration,
		}
		pp.metrics[name] = metric
	}

	metric.TotalCalls++
	metric.TotalDuration += duration
	metric.AverageDuration = time.Duration(int64(metric.TotalDuration) / metric.TotalCalls)
	metric.LastCalled = time.Now()

	if duration < metric.MinDuration {
		metric.MinDuration = duration
	}
	if duration > metric.MaxDuration {
		metric.MaxDuration = duration
	}

	if err != nil {
		metric.ErrorCount++
	}

	// Log slow operations
	if duration > 5*time.Second {
		log.Printf("⚠️ SLOW OPERATION: %s took %v", name, duration)
	} else if duration > 1*time.Second {
		log.Printf("🐌 MODERATE OPERATION: %s took %v", name, duration)
	}
}

// GetComprehensiveReport generates a comprehensive performance report
func (pp *PerformanceProfiler) GetComprehensiveReport(ctx context.Context) (*ProfilerReport, error) {
	report := &ProfilerReport{
		Timestamp: time.Now(),
	}

	// Get system information
	report.SystemInfo = pp.getSystemInfo()
	
	// Get database health
	dbHealth, err := pp.getDatabaseHealth(ctx)
	if err != nil {
		log.Printf("Warning: Could not get database health: %v", err)
		dbHealth = DatabaseHealth{}
	}
	report.DatabaseHealth = dbHealth

	// Get slow operations
	report.SlowOperations = pp.getSlowOperations()
	
	// Identify bottlenecks
	report.Bottlenecks = pp.identifyBottlenecks(report.SystemInfo, dbHealth, report.SlowOperations)
	
	// Generate recommendations
	report.Recommendations = pp.generateRecommendations(report)

	return report, nil
}

// getSystemInfo collects system performance information
func (pp *PerformanceProfiler) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:     runtime.Version(),
		NumCPU:        runtime.NumCPU(),
		NumGoroutines: runtime.NumGoroutine(),
		MemoryAlloc:   m.Alloc / 1024 / 1024,       // MB
		MemoryTotal:   m.TotalAlloc / 1024 / 1024,  // MB
		MemorySys:     m.Sys / 1024 / 1024,         // MB
		GCCycles:      m.NumGC,
		Timestamp:     time.Now(),
	}
}

// getDatabaseHealth analyzes database performance
func (pp *PerformanceProfiler) getDatabaseHealth(ctx context.Context) (DatabaseHealth, error) {
	health := DatabaseHealth{}
	
	sqlDB, err := pp.db.DB()
	if err != nil {
		return health, err
	}

	// Get connection pool stats
	stats := sqlDB.Stats()
	health.ConnectionsOpen = stats.OpenConnections
	health.ConnectionsInUse = stats.InUse
	health.ConnectionsIdle = stats.Idle

	// Test query performance
	start := time.Now()
	var count int64
	if err := pp.db.WithContext(ctx).Raw("SELECT 1").Scan(&count).Error; err == nil {
		health.QueryDuration = time.Since(start)
	}

	// Try to get database size (MySQL/PostgreSQL specific)
	var size string
	pp.db.WithContext(ctx).Raw(`
		SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) as size_mb 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
	`).Scan(&size)
	health.DatabaseSize = size + " MB"

	// Check for slow queries (approximate)
	if health.QueryDuration > 100*time.Millisecond {
		health.SlowQueryCount = 1
	}

	return health, nil
}

// getSlowOperations returns operations sorted by average duration
func (pp *PerformanceProfiler) getSlowOperations() []PerformanceMetric {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	var operations []PerformanceMetric
	for _, metric := range pp.metrics {
		if metric.AverageDuration > 500*time.Millisecond || metric.MaxDuration > 2*time.Second {
			operations = append(operations, *metric)
		}
	}

	return operations
}

// identifyBottlenecks analyzes data to find performance issues
func (pp *PerformanceProfiler) identifyBottlenecks(sysInfo SystemInfo, dbHealth DatabaseHealth, slowOps []PerformanceMetric) []Bottleneck {
	var bottlenecks []Bottleneck

	// Check memory usage
	if sysInfo.MemoryAlloc > 500 { // More than 500MB
		bottlenecks = append(bottlenecks, Bottleneck{
			Type:        "Memory",
			Description: fmt.Sprintf("High memory usage: %d MB allocated", sysInfo.MemoryAlloc),
			Severity:    "medium",
			Impact:      "May cause GC pressure and slow response times",
			Suggestion:  "Consider implementing connection pooling or reducing memory allocations",
			DetectedAt:  time.Now(),
		})
	}

	// Check goroutine count
	if sysInfo.NumGoroutines > 1000 {
		bottlenecks = append(bottlenecks, Bottleneck{
			Type:        "Concurrency",
			Description: fmt.Sprintf("High goroutine count: %d", sysInfo.NumGoroutines),
			Severity:    "high",
			Impact:      "May indicate goroutine leaks or excessive concurrency",
			Suggestion:  "Review goroutine management and consider using worker pools",
			DetectedAt:  time.Now(),
		})
	}

	// Check database connections
	if dbHealth.ConnectionsOpen > 50 {
		bottlenecks = append(bottlenecks, Bottleneck{
			Type:        "Database",
			Description: fmt.Sprintf("High database connection count: %d", dbHealth.ConnectionsOpen),
			Severity:    "medium",
			Impact:      "May exhaust database connection pool",
			Suggestion:  "Optimize connection pooling settings or reduce connection usage",
			DetectedAt:  time.Now(),
		})
	}

	// Check query performance
	if dbHealth.QueryDuration > 1*time.Second {
		bottlenecks = append(bottlenecks, Bottleneck{
			Type:        "Database",
			Description: fmt.Sprintf("Slow database queries: avg %v", dbHealth.QueryDuration),
			Severity:    "critical",
			Impact:      "Direct cause of timeout issues",
			Suggestion:  "Add database indexes, optimize queries, or increase timeout values",
			DetectedAt:  time.Now(),
		})
	}

	// Check slow operations
	for _, op := range slowOps {
		if op.MaxDuration > 5*time.Second {
			bottlenecks = append(bottlenecks, Bottleneck{
				Type:        "Operation",
				Description: fmt.Sprintf("Slow operation '%s': max %v, avg %v", op.Name, op.MaxDuration, op.AverageDuration),
				Severity:    "high",
				Impact:      "May cause request timeouts",
				Suggestion:  "Optimize this operation or implement caching",
				DetectedAt:  time.Now(),
			})
		}
	}

	return bottlenecks
}

// generateRecommendations creates actionable recommendations
func (pp *PerformanceProfiler) generateRecommendations(report *ProfilerReport) []string {
	var recommendations []string

	// System-level recommendations
	if report.SystemInfo.MemoryAlloc > 300 {
		recommendations = append(recommendations, 
			"🔧 Consider implementing object pooling to reduce memory allocations")
	}

	if report.SystemInfo.NumGoroutines > 500 {
		recommendations = append(recommendations, 
			"⚡ Implement goroutine pools to limit concurrency")
	}

	// Database recommendations
	if report.DatabaseHealth.QueryDuration > 500*time.Millisecond {
		recommendations = append(recommendations,
			"🗃️ Add database indexes for frequently queried fields",
			"📊 Consider implementing query result caching",
			"⚡ Use raw SQL for complex queries instead of ORM")
	}

	if report.DatabaseHealth.ConnectionsOpen > 30 {
		recommendations = append(recommendations,
			"🔌 Optimize database connection pool settings",
			"⏰ Implement connection timeout and retry logic")
	}

	// Operation-specific recommendations
	if len(report.SlowOperations) > 0 {
		recommendations = append(recommendations,
			"🚀 Implement the Ultra-Fast Payment service for critical paths",
			"📊 Add performance monitoring to all slow operations",
			"⚡ Consider asynchronous processing for non-critical operations")
	}

	// General recommendations for timeout issues
	recommendations = append(recommendations,
		"⏰ Increase HTTP request timeout values",
		"🔄 Implement request retry logic with exponential backoff",
		"📈 Add detailed logging to identify exact timeout causes",
		"🏎️ Use the new Ultra-Fast endpoints for time-critical operations")

	return recommendations
}

// GetMetrics returns current performance metrics
func (pp *PerformanceProfiler) GetMetrics() map[string]PerformanceMetric {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	result := make(map[string]PerformanceMetric)
	for k, v := range pp.metrics {
		result[k] = *v
	}
	return result
}

// ClearMetrics resets all performance metrics
func (pp *PerformanceProfiler) ClearMetrics() {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.metrics = make(map[string]*PerformanceMetric)
	log.Println("🧹 Performance metrics cleared")
}

// LogSystemStatus prints current system status
func (pp *PerformanceProfiler) LogSystemStatus() {
	sysInfo := pp.getSystemInfo()
	log.Printf("📊 SYSTEM STATUS: Memory: %dMB, Goroutines: %d, GC Cycles: %d", 
		sysInfo.MemoryAlloc, sysInfo.NumGoroutines, sysInfo.GCCycles)
}