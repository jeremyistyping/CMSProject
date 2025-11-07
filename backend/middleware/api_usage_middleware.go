package middleware

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// APIUsageStats stores API endpoint usage statistics
type APIUsageStats struct {
	Endpoint   string    `json:"endpoint"`
	Method     string    `json:"method"`
	Count      int64     `json:"count"`
	LastUsed   time.Time `json:"last_used"`
	FirstUsed  time.Time `json:"first_used"`
	AvgLatency float64   `json:"avg_latency"`
}

// APIUsageTracker tracks API endpoint usage
type APIUsageTracker struct {
	stats map[string]*APIUsageStats
	mutex sync.RWMutex
}

var globalUsageTracker = &APIUsageTracker{
	stats: make(map[string]*APIUsageStats),
}

// APIUsageMiddleware tracks API endpoint usage
func APIUsageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		// Process request
		c.Next()
		
		// Track usage
		endpointKey := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		latency := time.Since(startTime).Milliseconds()
		
		globalUsageTracker.recordUsage(endpointKey, c.Request.Method, c.Request.URL.Path, float64(latency))
		
		// Log usage for monitoring
		if gin.Mode() == gin.DebugMode {
			log.Printf("[API-USAGE] %s - %dms", endpointKey, latency)
		}
	}
}

// recordUsage records API endpoint usage statistics
func (tracker *APIUsageTracker) recordUsage(key, method, endpoint string, latency float64) {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	
	stat, exists := tracker.stats[key]
	if !exists {
		stat = &APIUsageStats{
			Endpoint:  endpoint,
			Method:    method,
			Count:     0,
			FirstUsed: time.Now(),
		}
		tracker.stats[key] = stat
	}
	
	stat.Count++
	stat.LastUsed = time.Now()
	
	// Calculate average latency
	stat.AvgLatency = (stat.AvgLatency*float64(stat.Count-1) + latency) / float64(stat.Count)
}

// GetAPIUsageStats returns current API usage statistics
func GetAPIUsageStats() map[string]*APIUsageStats {
	globalUsageTracker.mutex.RLock()
	defer globalUsageTracker.mutex.RUnlock()
	
	// Return a copy to avoid concurrent access issues
	result := make(map[string]*APIUsageStats)
	for k, v := range globalUsageTracker.stats {
		result[k] = &APIUsageStats{
			Endpoint:   v.Endpoint,
			Method:     v.Method,
			Count:      v.Count,
			LastUsed:   v.LastUsed,
			FirstUsed:  v.FirstUsed,
			AvgLatency: v.AvgLatency,
		}
	}
	
	return result
}

// GetUnusedEndpoints identifies potentially unused endpoints
func GetUnusedEndpoints(allEndpoints []string, minimumUsage int64) []string {
	globalUsageTracker.mutex.RLock()
	defer globalUsageTracker.mutex.RUnlock()
	
	var unused []string
	for _, endpoint := range allEndpoints {
		found := false
		for _, stat := range globalUsageTracker.stats {
			if endpoint == stat.Endpoint && stat.Count >= minimumUsage {
				found = true
				break
			}
		}
		if !found {
			unused = append(unused, endpoint)
		}
	}
	
	return unused
}

// ResetUsageStats resets all usage statistics
func ResetUsageStats() {
	globalUsageTracker.mutex.Lock()
	defer globalUsageTracker.mutex.Unlock()
	
	globalUsageTracker.stats = make(map[string]*APIUsageStats)
	log.Println("[API-USAGE] Usage statistics reset")
}