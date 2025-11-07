package routes

import (
	"net/http"
	"strconv"
	"time"

	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UltraFastPaymentRoutes provides the fastest possible payment endpoints
type UltraFastPaymentRoutes struct {
	service *services.UltraFastPaymentService
}

// NewUltraFastPaymentRoutes creates ultra-fast payment routes
func NewUltraFastPaymentRoutes(db *gorm.DB) *UltraFastPaymentRoutes {
	return &UltraFastPaymentRoutes{
		service: services.NewUltraFastPaymentService(db),
	}
}

// SetupUltraFastPaymentRoutes configures ultra-fast payment routes with minimal middleware
func (ufpr *UltraFastPaymentRoutes) SetupUltraFastPaymentRoutes(router *gin.Engine) {
	// Create ultra-fast group with only essential middleware
	ultraFast := router.Group("/api/ultra-fast")
	
	// Apply only critical middleware - remove everything non-essential
	ultraFast.Use(
		middleware.AuthRequired(), // Only auth, skip everything else for speed
	)

	// Ultra-fast payment endpoint
	ultraFast.POST("/payment", ufpr.recordPaymentUltraFast)
	
	// Health check endpoint (no auth needed)
	router.GET("/api/ultra-fast/health", ufpr.healthCheck)
}

// recordPaymentUltraFast handles ultra-fast payment recording
func (ufpr *UltraFastPaymentRoutes) recordPaymentUltraFast(c *gin.Context) {
	startTime := time.Now()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Parse request with minimal validation
	var req services.UltraFastPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// Set user ID from auth
	if userIDUint, ok := userID.(uint); ok {
		req.UserID = uint64(userIDUint)
	} else {
		// Try to convert from other types
		if userIDStr, ok := userID.(string); ok {
			if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
				req.UserID = id
			}
		} else if userIDFloat, ok := userID.(float64); ok {
			req.UserID = uint64(userIDFloat)
		}
	}

	if req.UserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	// Process payment with timeout context
	err := ufpr.service.RecordPaymentUltraFast(&req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":         false,
			"message":         "Payment processing failed",
			"error":           err.Error(),
			"processing_time": time.Since(startTime).String(),
		})
		return
	}

	// Start async journal creation (non-blocking)
	go ufpr.service.CreateJournalEntryAsync(&req)

	// Return immediate response
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Payment recorded (ultra-fast mode)",
		"processing_time": time.Since(startTime).String(),
	})
}

// healthCheck provides a simple health check endpoint
func (ufpr *UltraFastPaymentRoutes) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "ultra-fast-payment",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}