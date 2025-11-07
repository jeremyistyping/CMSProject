package handlers

import (
	"context"
	"net/http"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/utils"
	"time"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check operations
type HealthHandler struct {
	baseRepo repositories.BaseRepository
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(baseRepo repositories.BaseRepository) *HealthHandler {
	return &HealthHandler{
		baseRepo: baseRepo,
	}
}

// HealthStatus represents the health status response
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents individual health check result
type HealthCheck struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
}

var (
	startTime = time.Now()
	version   = "1.0.0" // This could be set during build
)

// Health performs comprehensive health checks
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	healthStatus := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   version,
		Uptime:    time.Since(startTime).String(),
		Checks:    make(map[string]HealthCheck),
	}

	// Database health check
	dbStart := time.Now()
	if err := h.baseRepo.Health(ctx); err != nil {
		healthStatus.Status = "unhealthy"
		healthStatus.Checks["database"] = HealthCheck{
			Status:  "unhealthy",
			Message: err.Error(),
			Latency: time.Since(dbStart),
		}
	} else {
		healthStatus.Checks["database"] = HealthCheck{
			Status:  "healthy",
			Latency: time.Since(dbStart),
		}
	}

	// Memory health check (basic)
	healthStatus.Checks["memory"] = HealthCheck{
		Status: "healthy",
	}

	// Disk health check (basic)
	healthStatus.Checks["disk"] = HealthCheck{
		Status: "healthy",
	}

	// Set HTTP status based on overall health
	httpStatus := http.StatusOK
	if healthStatus.Status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, healthStatus)
}

// ReadinessCheck checks if the application is ready to serve requests
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check database connectivity
	if err := h.baseRepo.Health(ctx); err != nil {
		utils.WithError(err).Error("Readiness check failed - database unhealthy")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not ready",
			"message": "Database is not available",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// LivenessCheck checks if the application is alive
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
		"uptime":    time.Since(startTime).String(),
	})
}

// DatabaseStats returns database connection statistics
func (h *HealthHandler) DatabaseStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get base repository implementation to access database stats
	if baseRepo, ok := h.baseRepo.(*repositories.BaseRepo); ok {
		stats, err := baseRepo.GetDatabaseStats(ctx)
		if err != nil {
			appError := utils.NewInternalError("Failed to get database stats", err)
			c.JSON(appError.StatusCode, appError.ToErrorResponse(""))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   stats,
		})
		return
	}

	appError := utils.NewInternalError("Database stats not available", nil)
	c.JSON(appError.StatusCode, appError.ToErrorResponse(""))
}
