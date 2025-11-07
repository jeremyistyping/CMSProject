package controllers

import (
	"net/http"
	"strconv"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

type BalanceMonitoringController struct {
	monitoringService *services.BalanceMonitoringService
}

func NewBalanceMonitoringController(monitoringService *services.BalanceMonitoringService) *BalanceMonitoringController {
	return &BalanceMonitoringController{
		monitoringService: monitoringService,
	}
}

// CheckBalanceSync godoc
// @Summary Check balance synchronization
// @Description Check synchronization between cash/bank accounts and GL accounts
// @Tags Balance Monitoring
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} services.BalanceMonitoringResult
// @Router /api/monitoring/balance-sync [get]
func (c *BalanceMonitoringController) CheckBalanceSync(ctx *gin.Context) {
	result, err := c.monitoringService.CheckBalanceSynchronization()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check balance synchronization",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "Balance synchronization check completed",
	})
}

// FixBalanceDiscrepancies godoc
// @Summary Fix balance discrepancies
// @Description Automatically fix balance discrepancies by updating GL account balances
// @Tags Balance Monitoring
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.APIResponse
// @Router /api/monitoring/fix-discrepancies [post]
func (c *BalanceMonitoringController) FixBalanceDiscrepancies(ctx *gin.Context) {
	// First, check for discrepancies
	result, err := c.monitoringService.CheckBalanceSynchronization()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check balance synchronization",
			"details": err.Error(),
		})
		return
	}

	if len(result.Discrepancies) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No balance discrepancies found",
			"data": gin.H{
				"discrepancies_fixed": 0,
				"total_discrepancies": 0,
			},
		})
		return
	}

	// Auto-fix discrepancies
	err = c.monitoringService.AutoFixDiscrepancies(result)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fix balance discrepancies",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Balance discrepancies fixed successfully",
		"data": gin.H{
			"discrepancies_fixed": len(result.Discrepancies),
			"total_discrepancies": len(result.Discrepancies),
		},
	})
}

// GetBalanceHealth godoc
// @Summary Get balance health metrics
// @Description Get comprehensive balance health metrics and synchronization status
// @Tags Balance Monitoring
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.APIResponse
// @Router /api/monitoring/balance-health [get]
func (c *BalanceMonitoringController) GetBalanceHealth(ctx *gin.Context) {
	health, err := c.monitoringService.GetBalanceHealth()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get balance health metrics",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
		"message": "Balance health metrics retrieved successfully",
	})
}

// GetBalanceDiscrepancies godoc
// @Summary Get current balance discrepancies
// @Description Get list of current balance discrepancies between cash/bank and GL accounts
// @Tags Balance Monitoring
// @Accept json
// @Produce json
// @Security Bearer
// @Param limit query int false "Limit number of results" default(50)
// @Success 200 {array} services.BalanceDiscrepancy
// @Router /api/monitoring/discrepancies [get]
func (c *BalanceMonitoringController) GetBalanceDiscrepancies(ctx *gin.Context) {
	result, err := c.monitoringService.CheckBalanceSynchronization()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check balance synchronization",
			"details": err.Error(),
		})
		return
	}

	// Apply limit if specified
	limit := 50 // default limit
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	discrepancies := result.Discrepancies
	if len(discrepancies) > limit {
		discrepancies = discrepancies[:limit]
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    discrepancies,
		"meta": gin.H{
			"total_discrepancies":      len(result.Discrepancies),
			"returned_discrepancies":   len(discrepancies),
			"total_accounts_checked":   result.TotalAccountsChecked,
			"synchronized_accounts":    result.SynchronizedAccounts,
			"unsynchronized_accounts":  result.UnsynchronizedAccounts,
			"status":                   result.Status,
			"check_time":               result.CheckTime,
		},
	})
}

// GetSyncStatus godoc
// @Summary Get synchronization status summary
// @Description Get a quick summary of balance synchronization status
// @Tags Balance Monitoring
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.APIResponse
// @Router /api/monitoring/sync-status [get]
func (c *BalanceMonitoringController) GetSyncStatus(ctx *gin.Context) {
	result, err := c.monitoringService.CheckBalanceSynchronization()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check synchronization status",
			"details": err.Error(),
		})
		return
	}

	syncPercentage := float64(result.SynchronizedAccounts) / float64(result.TotalAccountsChecked) * 100

	status := gin.H{
		"status":                  result.Status,
		"total_accounts":          result.TotalAccountsChecked,
		"synchronized_accounts":   result.SynchronizedAccounts,
		"unsynchronized_accounts": result.UnsynchronizedAccounts,
		"sync_percentage":         syncPercentage,
		"has_discrepancies":       len(result.Discrepancies) > 0,
		"discrepancy_count":       len(result.Discrepancies),
		"check_time":              result.CheckTime,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
		"message": "Synchronization status retrieved successfully",
	})
}
