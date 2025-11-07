package controllers

import (
	"net/http"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BalanceHealthController handles balance health monitoring and auto-healing
type BalanceHealthController struct {
	balanceValidationService *services.BalanceValidationService
}

// NewBalanceHealthController creates a new balance health controller
func NewBalanceHealthController(db *gorm.DB) *BalanceHealthController {
	return &BalanceHealthController{
		balanceValidationService: services.NewBalanceValidationService(db),
	}
}

// HealthCheck returns current balance health status
// @Summary Check balance health status
// @Description Check the current health of the balance sheet and accounting equation
// @Tags Balance Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Balance health status"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/balance-health/check [get]
func (ctrl *BalanceHealthController) HealthCheck(c *gin.Context) {
	validation, err := ctrl.balanceValidationService.ValidateRealTimeBalance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check balance health",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"message":          "Balance health check completed",
		"balance_status":   validation.IsValid,
		"data":            validation,
	})
}

// AutoHeal performs automatic healing of balance issues
// @Summary Auto-heal balance issues
// @Description Automatically detect and fix common balance synchronization issues
// @Tags Balance Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Auto-healing result"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/balance-health/auto-heal [post]
func (ctrl *BalanceHealthController) AutoHeal(c *gin.Context) {
	result, err := ctrl.balanceValidationService.AutoHealBalanceIssues()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Auto-healing failed",
			"message": err.Error(),
		})
		return
	}

	status := http.StatusOK
	message := "Auto-healing completed successfully"
	
	if !result.IsValid {
		status = http.StatusPartialContent
		message = "Auto-healing completed with warnings"
	}

	c.JSON(status, gin.H{
		"status":  "success",
		"message": message,
		"healing_result": result,
	})
}

// DetailedReport provides detailed balance validation report
// @Summary Get detailed balance report
// @Description Get detailed balance validation report with account breakdowns and recommendations
// @Tags Balance Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Detailed balance report"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/balance-health/detailed-report [get]
func (ctrl *BalanceHealthController) DetailedReport(c *gin.Context) {
	report, err := ctrl.balanceValidationService.GetDetailedValidationReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate detailed report",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Detailed balance report generated",
		"report":  report,
	})
}

// ScheduledMaintenance performs scheduled maintenance (for cron jobs)
// @Summary Perform scheduled maintenance
// @Description Perform scheduled balance maintenance and auto-healing (used by cron jobs)
// @Tags Balance Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Maintenance result"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/balance-health/scheduled-maintenance [post]
func (ctrl *BalanceHealthController) ScheduledMaintenance(c *gin.Context) {
	err := ctrl.balanceValidationService.ScheduledHealthCheck()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Scheduled maintenance failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Scheduled maintenance completed successfully",
		"timestamp": "Maintenance logged in database",
	})
}