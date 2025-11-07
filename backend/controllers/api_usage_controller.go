package controllers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"app-sistem-akuntansi/middleware"

	"github.com/gin-gonic/gin"
)

// APIUsageController handles API usage monitoring endpoints
type APIUsageController struct{}

// NewAPIUsageController creates a new API usage controller
func NewAPIUsageController() *APIUsageController {
	return &APIUsageController{}
}

// GetAPIUsageStats returns API usage statistics
// @Summary Get API usage statistics
// @Description Get comprehensive API endpoint usage statistics
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/api-usage/stats [get]
func (ctrl *APIUsageController) GetAPIUsageStats(c *gin.Context) {
	stats := middleware.GetAPIUsageStats()
	
	// Convert to slice for easier sorting and filtering
	var usageList []middleware.APIUsageStats
	for _, stat := range stats {
		usageList = append(usageList, *stat)
	}
	
	// Sort by usage count (descending)
	sort.Slice(usageList, func(i, j int) bool {
		return usageList[i].Count > usageList[j].Count
	})
	
	// Calculate summary statistics
	totalEndpoints := len(usageList)
	totalRequests := int64(0)
	var avgLatency float64
	
	for _, stat := range usageList {
		totalRequests += stat.Count
		avgLatency += stat.AvgLatency
	}
	
	if totalEndpoints > 0 {
		avgLatency /= float64(totalEndpoints)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"summary": gin.H{
				"total_endpoints":    totalEndpoints,
				"total_requests":     totalRequests,
				"average_latency":    avgLatency,
				"monitoring_since":   getMonitoringStartTime(usageList),
			},
			"endpoints": usageList,
		},
	})
}

// GetTopEndpoints returns most frequently used endpoints
// @Summary Get top used endpoints
// @Description Get list of most frequently used API endpoints
// @Tags monitoring
// @Accept json
// @Produce json
// @Param limit query int false "Number of top endpoints to return" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/api-usage/top [get]
func (ctrl *APIUsageController) GetTopEndpoints(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	
	stats := middleware.GetAPIUsageStats()
	
	// Convert to slice and sort by usage count
	var usageList []middleware.APIUsageStats
	for _, stat := range stats {
		usageList = append(usageList, *stat)
	}
	
	sort.Slice(usageList, func(i, j int) bool {
		return usageList[i].Count > usageList[j].Count
	})
	
	// Limit results
	if limit > len(usageList) {
		limit = len(usageList)
	}
	topEndpoints := usageList[:limit]
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"limit":        limit,
			"total_found":  len(usageList),
			"endpoints":    topEndpoints,
		},
	})
}

// GetUnusedEndpoints returns potentially unused endpoints
// @Summary Get unused endpoints
// @Description Get list of potentially unused API endpoints
// @Tags monitoring
// @Accept json
// @Produce json
// @Param min_usage query int false "Minimum usage count to consider as used" default(1)
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/api-usage/unused [get]
func (ctrl *APIUsageController) GetUnusedEndpoints(c *gin.Context) {
	minUsageStr := c.DefaultQuery("min_usage", "1")
	minUsage, err := strconv.ParseInt(minUsageStr, 10, 64)
	if err != nil || minUsage < 0 {
		minUsage = 1
	}
	
	// This would ideally come from route registration
	// For now, we'll identify endpoints with 0 or low usage from tracked stats
	stats := middleware.GetAPIUsageStats()
	
	var lowUsageEndpoints []middleware.APIUsageStats
	var unusedCount int
	
	for _, stat := range stats {
		if stat.Count < minUsage {
			lowUsageEndpoints = append(lowUsageEndpoints, *stat)
			if stat.Count == 0 {
				unusedCount++
			}
		}
	}
	
	// Sort by count (ascending - least used first)
	sort.Slice(lowUsageEndpoints, func(i, j int) bool {
		return lowUsageEndpoints[i].Count < lowUsageEndpoints[j].Count
	})
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"minimum_usage":      minUsage,
			"low_usage_count":    len(lowUsageEndpoints),
			"completely_unused":  unusedCount,
			"endpoints":          lowUsageEndpoints,
			"suggestion":         "Consider reviewing endpoints with very low usage for potential removal",
		},
	})
}

// ResetUsageStats resets all usage statistics
// @Summary Reset API usage statistics
// @Description Reset all API usage tracking statistics (admin only)
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/api-usage/reset [post]
func (ctrl *APIUsageController) ResetUsageStats(c *gin.Context) {
	middleware.ResetUsageStats()
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "API usage statistics have been reset",
		"reset_at": time.Now(),
	})
}

// GetUsageAnalytics returns usage analytics and insights
// @Summary Get API usage analytics
// @Description Get detailed analytics and insights about API usage patterns
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/api-usage/analytics [get]
func (ctrl *APIUsageController) GetUsageAnalytics(c *gin.Context) {
	stats := middleware.GetAPIUsageStats()
	
	if len(stats) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"message": "No usage data available yet",
				"analytics": gin.H{},
			},
		})
		return
	}
	
	// Calculate analytics
	var totalRequests int64
	var totalLatency float64
	methodCounts := make(map[string]int)
	pathCounts := make(map[string]int64)
	
	for _, stat := range stats {
		totalRequests += stat.Count
		totalLatency += stat.AvgLatency * float64(stat.Count)
		methodCounts[stat.Method]++
		
		// Group by path patterns (simplified)
		pathCounts[stat.Endpoint] += stat.Count
	}
	
	avgLatency := totalLatency / float64(totalRequests)
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"overview": gin.H{
				"total_endpoints": len(stats),
				"total_requests":  totalRequests,
				"average_latency": avgLatency,
			},
			"methods_distribution": methodCounts,
			"most_used_paths": getTopPaths(pathCounts, 10),
			"performance_insights": gin.H{
				"avg_latency": avgLatency,
				"status": getPerformanceStatus(avgLatency),
			},
		},
	})
}

// Helper functions

func getMonitoringStartTime(usageList []middleware.APIUsageStats) *time.Time {
	if len(usageList) == 0 {
		return nil
	}
	
	earliest := usageList[0].FirstUsed
	for _, stat := range usageList {
		if stat.FirstUsed.Before(earliest) {
			earliest = stat.FirstUsed
		}
	}
	
	return &earliest
}

func getTopPaths(pathCounts map[string]int64, limit int) []map[string]interface{} {
	type pathCount struct {
		Path  string
		Count int64
	}
	
	var paths []pathCount
	for path, count := range pathCounts {
		paths = append(paths, pathCount{Path: path, Count: count})
	}
	
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Count > paths[j].Count
	})
	
	if limit > len(paths) {
		limit = len(paths)
	}
	
	var result []map[string]interface{}
	for i := 0; i < limit; i++ {
		result = append(result, map[string]interface{}{
			"path":  paths[i].Path,
			"count": paths[i].Count,
		})
	}
	
	return result
}

func getPerformanceStatus(avgLatency float64) string {
	if avgLatency < 100 {
		return "excellent"
	} else if avgLatency < 500 {
		return "good"
	} else if avgLatency < 1000 {
		return "acceptable"
	} else {
		return "needs_attention"
	}
}