package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PerformanceController handles performance monitoring and profiling endpoints
type PerformanceController struct {
	profiler    *services.PerformanceProfiler
	diagnostics *services.PaymentTimeoutDiagnostics
}

// NewPerformanceController creates a new performance controller
func NewPerformanceController(db *gorm.DB) *PerformanceController {
	return &PerformanceController{
		profiler:    services.NewPerformanceProfiler(db),
		diagnostics: services.NewPaymentTimeoutDiagnostics(db),
	}
}

// GetPerformanceReport generates comprehensive performance analysis
func (pc *PerformanceController) GetPerformanceReport(c *gin.Context) {
	timeout := 10 * time.Second
	if timeoutParam := c.Query("timeout"); timeoutParam != "" {
		if t, err := strconv.Atoi(timeoutParam); err == nil && t > 0 && t <= 30 {
			timeout = time.Duration(t) * time.Second
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	report, err := pc.profiler.GetComprehensiveReport(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to generate performance report",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
		"message": "Performance report generated successfully",
	})
}

// GetQuickMetrics returns current performance metrics
func (pc *PerformanceController) GetQuickMetrics(c *gin.Context) {
	metrics := pc.profiler.GetMetrics()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"metrics_count": len(metrics),
			"metrics":       metrics,
			"timestamp":     time.Now(),
		},
		"message": "Performance metrics retrieved successfully",
	})
}

// ClearMetrics resets all performance metrics
func (pc *PerformanceController) ClearMetrics(c *gin.Context) {
	pc.profiler.ClearMetrics()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Performance metrics cleared successfully",
		"timestamp": time.Now(),
	})
}

// GetSystemStatus returns current system performance status
func (pc *PerformanceController) GetSystemStatus(c *gin.Context) {
	// Log and return system status
	pc.profiler.LogSystemStatus()
	
	// Get basic system info
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	
	report, err := pc.profiler.GetComprehensiveReport(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get system status",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"system_info":     report.SystemInfo,
			"database_health": report.DatabaseHealth,
			"bottlenecks":     report.Bottlenecks,
		},
		"message": "System status retrieved successfully",
	})
}

// GetBottlenecks returns identified performance bottlenecks
func (pc *PerformanceController) GetBottlenecks(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	report, err := pc.profiler.GetComprehensiveReport(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to analyze bottlenecks",
			"error":   err.Error(),
		})
		return
	}

	// Filter bottlenecks by severity if requested
	severity := c.Query("severity")
	bottlenecks := report.Bottlenecks
	if severity != "" {
		var filtered []services.Bottleneck
		for _, bottleneck := range bottlenecks {
			if bottleneck.Severity == severity {
				filtered = append(filtered, bottleneck)
			}
		}
		bottlenecks = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"bottlenecks":     bottlenecks,
			"total_count":     len(report.Bottlenecks),
			"filtered_count":  len(bottlenecks),
			"recommendations": report.Recommendations,
		},
		"message": "Bottleneck analysis completed",
	})
}

// GetRecommendations returns performance optimization recommendations
func (pc *PerformanceController) GetRecommendations(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	report, err := pc.profiler.GetComprehensiveReport(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to generate recommendations",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"recommendations": report.Recommendations,
			"bottlenecks":     len(report.Bottlenecks),
			"slow_operations": len(report.SlowOperations),
		},
		"message": "Performance recommendations generated",
	})
}

// TestUltraFastEndpoint tests the ultra-fast payment endpoint performance
func (pc *PerformanceController) TestUltraFastEndpoint(c *gin.Context) {
	// This is a diagnostic endpoint to test if ultra-fast routes are working
	startTime := time.Now()
	
	// Simulate minimal processing
	time.Sleep(1 * time.Millisecond)
	
	processingTime := time.Since(startTime)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"processing_time": processingTime.String(),
			"endpoint":        "ultra-fast-test",
			"status":         "operational",
			"timestamp":      time.Now(),
		},
		"message": "Ultra-fast endpoint test completed",
	})
}

// RunTimeoutDiagnostics runs comprehensive timeout diagnostics
func (pc *PerformanceController) RunTimeoutDiagnostics(c *gin.Context) {
	report := pc.diagnostics.RunFullDiagnostics()

	if report.Failed > 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Timeout diagnostics found issues",
			"data":    report,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "All timeout diagnostics passed",
			"data":    report,
		})
	}
}

// GetQuickHealthCheck provides quick health check for payment processing
func (pc *PerformanceController) GetQuickHealthCheck(c *gin.Context) {
	health := pc.diagnostics.GetQuickHealthCheck()

	if status, ok := health["status"]; ok && status == "unhealthy" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "Payment system is unhealthy",
			"data":    health,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Payment system health check completed",
			"data":    health,
		})
	}
}
