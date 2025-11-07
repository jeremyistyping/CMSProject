package middleware

import (
	"net/http"
	"strings"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

type CashBankValidationMiddleware struct {
	validationService *services.CashBankValidationService
	enhancedService   *services.CashBankEnhancedService
}

func NewCashBankValidationMiddleware(
	validationService *services.CashBankValidationService,
	enhancedService *services.CashBankEnhancedService,
) *CashBankValidationMiddleware {
	return &CashBankValidationMiddleware{
		validationService: validationService,
		enhancedService:   enhancedService,
	}
}

// ValidateCashBankSync validates synchronization before critical cash bank operations
func (m *CashBankValidationMiddleware) ValidateCashBankSync() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for read-only operations and health checks
		if m.isReadOnlyOperation(c) {
			c.Next()
			return
		}

		// Skip validation for health/sync endpoints to avoid circular dependency
		if m.isSyncOrHealthEndpoint(c) {
			c.Next()
			return
		}

		// Perform sync validation for write operations
		if err := m.enhancedService.ValidateSyncMiddleware(); err != nil {
			c.JSON(http.StatusPreconditionFailed, gin.H{
				"error":   "CashBank synchronization validation failed",
				"details": err.Error(),
				"action":  "Please fix synchronization issues before proceeding",
				"help":    "Use GET /api/cashbank/health to check sync status and POST /api/cashbank/sync/fix to auto-fix issues",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// HealthCheck middleware for general health monitoring
func (m *CashBankValidationMiddleware) HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := m.enhancedService.GetHealthStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to check health status",
				"details": err.Error(),
			})
			return
		}

		httpStatus := http.StatusOK
		if needsAttention, ok := status["needs_attention"].(bool); ok && needsAttention {
			httpStatus = http.StatusPartialContent // 206 indicates some issues
		}

		c.JSON(httpStatus, gin.H{
			"status": "success",
			"data":   status,
		})
	}
}

// DetailedSyncStatus returns detailed sync status
func (m *CashBankValidationMiddleware) DetailedSyncStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := m.validationService.GetSyncStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to get sync status",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   status,
		})
	}
}

// AutoFixSync attempts to fix sync issues automatically
func (m *CashBankValidationMiddleware) AutoFixSync() gin.HandlerFunc {
	return func(c *gin.Context) {
		fixedCount, err := m.enhancedService.AutoFixSyncIssues()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to auto-fix sync issues",
				"details": err.Error(),
			})
			return
		}

		message := "All accounts are already synchronized"
		if fixedCount > 0 {
			message = "Successfully fixed sync issues"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "success",
			"message":     message,
			"fixed_count": fixedCount,
		})
	}
}

// ValidateAccountIntegrity validates specific cash bank account integrity
func (m *CashBankValidationMiddleware) ValidateAccountIntegrity() gin.HandlerFunc {
	return func(c *gin.Context) {
		cashBankID := c.GetUint("cash_bank_id")
		if cashBankID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "cash_bank_id parameter is required",
			})
			return
		}

		integrity, err := m.enhancedService.ValidateAndFixCashBank(cashBankID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to validate account integrity",
				"details": err.Error(),
			})
			return
		}

		status := "healthy"
		if integrity.Issue != "SYNC_OK" {
			status = "unhealthy"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "success",
			"sync_status": status,
			"data":      integrity,
		})
	}
}

// LinkCashBankAccount links cash bank to COA account
func (m *CashBankValidationMiddleware) LinkCashBankAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			CashBankID uint `json:"cash_bank_id" binding:"required"`
			AccountID  uint `json:"account_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		err := m.enhancedService.LinkCashBankToAccount(request.CashBankID, request.AccountID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Failed to link cash bank to account",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Cash bank account successfully linked to COA",
		})
	}
}

// UnlinkCashBankAccount unlinks cash bank from COA account
func (m *CashBankValidationMiddleware) UnlinkCashBankAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			CashBankID uint `json:"cash_bank_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		err := m.enhancedService.UnlinkCashBankFromAccount(request.CashBankID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Failed to unlink cash bank from account",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Cash bank account successfully unlinked from COA",
		})
	}
}

// RecalculateBalance recalculates cash bank balance from transactions
func (m *CashBankValidationMiddleware) RecalculateBalance() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			CashBankID uint `json:"cash_bank_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		err := m.enhancedService.RecalculateBalance(request.CashBankID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to recalculate balance",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Cash bank balance successfully recalculated from transactions",
		})
	}
}

// SyncAllBalances syncs all cash bank balances
func (m *CashBankValidationMiddleware) SyncAllBalances() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := m.enhancedService.SyncAllBalances()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to sync all balances",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "All cash bank balances successfully synchronized with COA",
		})
	}
}

// Helper methods
func (m *CashBankValidationMiddleware) isReadOnlyOperation(c *gin.Context) bool {
	method := c.Request.Method
	return method == "GET" || method == "HEAD" || method == "OPTIONS"
}

func (m *CashBankValidationMiddleware) isSyncOrHealthEndpoint(c *gin.Context) bool {
	path := c.Request.URL.Path
	
	syncPaths := []string{
		"/api/cashbank/health",
		"/api/cashbank/sync",
		"/api/health/cashbank",
		"/api/admin/cashbank/sync",
	}

	for _, syncPath := range syncPaths {
		if strings.Contains(path, syncPath) {
			return true
		}
	}

	return false
}

// AddRoutes adds validation and health check routes
func (m *CashBankValidationMiddleware) AddRoutes(router *gin.RouterGroup) {
	// Health check routes
	health := router.Group("/health")
	{
		health.GET("/cashbank", m.HealthCheck())
		health.GET("/cashbank/sync", m.DetailedSyncStatus())
	}

	// Sync management routes
	sync := router.Group("/cashbank/sync")
	{
		sync.GET("/status", m.DetailedSyncStatus())
		sync.POST("/fix", m.AutoFixSync())
		sync.POST("/link", m.LinkCashBankAccount())
		sync.POST("/unlink", m.UnlinkCashBankAccount())
		sync.POST("/recalculate", m.RecalculateBalance())
		sync.POST("/sync-all", m.SyncAllBalances())
		sync.GET("/validate/:id", func(c *gin.Context) {
			// Extract ID from URL parameter and set it in context
			c.Set("cash_bank_id", c.Param("id"))
			m.ValidateAccountIntegrity()(c)
		})
	}

	// Admin routes for advanced sync management
	admin := router.Group("/admin/cashbank")
	{
		admin.GET("/integrity", m.DetailedSyncStatus())
		admin.POST("/emergency-sync", m.SyncAllBalances())
		admin.POST("/force-fix", m.AutoFixSync())
	}
}
