package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BalanceValidationMiddleware provides middleware for validating balance consistency
type BalanceValidationMiddleware struct {
	db                 *gorm.DB
	autoSyncService    *services.AutoBalanceSyncService
	validationConfig   *BalanceValidationConfig
	lastValidationTime time.Time
	lastReport         *services.BalanceConsistencyReport
}

// BalanceValidationConfig configuration for balance validation middleware
type BalanceValidationConfig struct {
	// EnableAutoFix automatically fixes detected balance issues
	EnableAutoFix bool `json:"enable_auto_fix"`
	
	// ValidateOnWrite validates balances after write operations (POST, PUT, DELETE)
	ValidateOnWrite bool `json:"validate_on_write"`
	
	// ValidateOnRead validates balances before read operations (GET)
	ValidateOnRead bool `json:"validate_on_read"`
	
	// CacheValidationDuration how long to cache validation results
	CacheValidationDuration time.Duration `json:"cache_validation_duration"`
	
	// IncludeValidationInResponse include validation results in API response
	IncludeValidationInResponse bool `json:"include_validation_in_response"`
	
	// OnlyValidateRelatedAccounts only validate accounts related to the current operation
	OnlyValidateRelatedAccounts bool `json:"only_validate_related_accounts"`
	
	// FailOnInconsistency return error response if balance inconsistency is detected
	FailOnInconsistency bool `json:"fail_on_inconsistency"`
	
	// ValidatePaths API paths that should be validated (if empty, validate all)
	ValidatePaths []string `json:"validate_paths"`
}

// NewBalanceValidationMiddleware creates a new balance validation middleware
func NewBalanceValidationMiddleware(db *gorm.DB, config *BalanceValidationConfig) *BalanceValidationMiddleware {
	accountRepo := repositories.NewAccountRepository(db)
	autoSyncService := services.NewAutoBalanceSyncService(db, accountRepo)

	if config == nil {
		config = &BalanceValidationConfig{
			EnableAutoFix:               false, // Default to safe mode
			ValidateOnWrite:             true,
			ValidateOnRead:              false,
			CacheValidationDuration:     5 * time.Minute,
			IncludeValidationInResponse: false,
			OnlyValidateRelatedAccounts: true,
			FailOnInconsistency:         false,
			ValidatePaths:               []string{},
		}
	}

	return &BalanceValidationMiddleware{
		db:               db,
		autoSyncService:  autoSyncService,
		validationConfig: config,
	}
}

// BalanceValidationHandler main middleware handler
func (m *BalanceValidationMiddleware) BalanceValidationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if this path should be validated
		if !m.shouldValidatePath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Determine if we should validate based on HTTP method
		shouldValidate := m.shouldValidateForMethod(c.Request.Method)
		
		if !shouldValidate {
			c.Next()
			return
		}

		// For write operations, validate after the operation
		if m.isWriteOperation(c.Request.Method) {
			c.Next() // Execute the main handler first
			m.validateAfterWrite(c)
		} else {
			// For read operations, validate before
			m.validateBeforeRead(c)
			c.Next()
		}
	}
}

// shouldValidatePath checks if the current path should be validated
func (m *BalanceValidationMiddleware) shouldValidatePath(path string) bool {
	if len(m.validationConfig.ValidatePaths) == 0 {
		// If no specific paths configured, validate cash bank and account related paths
		return strings.Contains(path, "/cash-bank") || 
			   strings.Contains(path, "/account") ||
			   strings.Contains(path, "/transaction")
	}

	for _, validPath := range m.validationConfig.ValidatePaths {
		if strings.Contains(path, validPath) {
			return true
		}
	}
	return false
}

// shouldValidateForMethod determines if validation should occur for this HTTP method
func (m *BalanceValidationMiddleware) shouldValidateForMethod(method string) bool {
	if m.isWriteOperation(method) {
		return m.validationConfig.ValidateOnWrite
	}
	return m.validationConfig.ValidateOnRead
}

// isWriteOperation checks if the HTTP method is a write operation
func (m *BalanceValidationMiddleware) isWriteOperation(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || 
		   method == http.MethodPatch || method == http.MethodDelete
}

// validateAfterWrite validates balances after write operations
func (m *BalanceValidationMiddleware) validateAfterWrite(c *gin.Context) {
	log.Printf("🔍 Running balance validation after write operation: %s %s", c.Request.Method, c.Request.URL.Path)

	// Check if we should use cached results
	if m.shouldUseCachedValidation() {
		m.handleValidationResult(c, m.lastReport)
		return
	}

	// Run balance validation
	report, err := m.autoSyncService.ValidateBalanceConsistency()
	if err != nil {
		log.Printf("⚠️ Balance validation failed: %v", err)
		// Don't fail the request, just log the error
		return
	}

	// Update cache
	m.lastValidationTime = time.Now()
	m.lastReport = report

	// Handle validation results
	m.handleValidationResult(c, report)
}

// validateBeforeRead validates balances before read operations
func (m *BalanceValidationMiddleware) validateBeforeRead(c *gin.Context) {
	log.Printf("🔍 Running balance validation before read operation: %s %s", c.Request.Method, c.Request.URL.Path)

	// For read operations, we're more lenient and use cached results
	if m.shouldUseCachedValidation() {
		m.handleValidationResult(c, m.lastReport)
		return
	}

	// Only run validation if we don't have recent results
	report, err := m.autoSyncService.ValidateBalanceConsistency()
	if err != nil {
		log.Printf("⚠️ Balance validation failed: %v", err)
		return
	}

	// Update cache
	m.lastValidationTime = time.Now()
	m.lastReport = report

	// Handle validation results
	m.handleValidationResult(c, report)
}

// shouldUseCachedValidation checks if we should use cached validation results
func (m *BalanceValidationMiddleware) shouldUseCachedValidation() bool {
	if m.lastReport == nil || m.lastValidationTime.IsZero() {
		return false
	}

	return time.Since(m.lastValidationTime) < m.validationConfig.CacheValidationDuration
}

// handleValidationResult processes validation results and determines response
func (m *BalanceValidationMiddleware) handleValidationResult(c *gin.Context, report *services.BalanceConsistencyReport) {
	if report == nil {
		return
	}

	// Auto-fix issues if enabled
	if !report.IsConsistent && m.validationConfig.EnableAutoFix {
		log.Println("🔧 Auto-fixing detected balance issues...")
		
		if err := m.autoSyncService.FixAllBalanceIssues(); err != nil {
			log.Printf("❌ Auto-fix failed: %v", err)
		} else {
			log.Println("✅ Auto-fix completed successfully")
			
			// Re-validate after fix
			if newReport, err := m.autoSyncService.ValidateBalanceConsistency(); err == nil {
				report = newReport
				m.lastReport = newReport
				m.lastValidationTime = time.Now()
			}
		}
	}

	// Include validation results in response if configured
	if m.validationConfig.IncludeValidationInResponse {
		c.Header("X-Balance-Validation", m.getValidationSummary(report))
	}

	// Fail request if inconsistency detected and configured to do so
	if !report.IsConsistent && m.validationConfig.FailOnInconsistency {
		c.JSON(http.StatusConflict, gin.H{
			"error":              "Balance inconsistency detected",
			"validation_report":  report,
			"message":            "Please resolve balance issues before continuing",
		})
		c.Abort()
		return
	}

	// Log validation results
	m.logValidationResults(report)
}

// getValidationSummary creates a summary string for HTTP headers
func (m *BalanceValidationMiddleware) getValidationSummary(report *services.BalanceConsistencyReport) string {
	if report.IsConsistent {
		return "consistent"
	}

	summary := fmt.Sprintf("inconsistent;cashbank_issues=%d;parent_child_issues=%d;balance_diff=%.2f", 
		len(report.CashBankIssues), len(report.ParentChildIssues), report.BalanceEquationDifference)
	
	return summary
}

// logValidationResults logs the validation results
func (m *BalanceValidationMiddleware) logValidationResults(report *services.BalanceConsistencyReport) {
	if report.IsConsistent {
		log.Println("✅ Balance validation passed - all balances are consistent")
		return
	}

	log.Printf("⚠️ Balance validation found issues:")
	if len(report.CashBankIssues) > 0 {
		log.Printf("  - %d cash bank sync issues", len(report.CashBankIssues))
	}
	if len(report.ParentChildIssues) > 0 {
		log.Printf("  - %d parent-child balance issues", len(report.ParentChildIssues))
	}
	if report.BalanceEquationDifference != 0 {
		log.Printf("  - Balance sheet difference: %.2f", report.BalanceEquationDifference)
	}
}

// Balance validation endpoint handlers

// GetBalanceValidationStatus provides an endpoint to check balance validation status
func (m *BalanceValidationMiddleware) GetBalanceValidationStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		report, err := m.autoSyncService.ValidateBalanceConsistency()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to validate balances",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":             "success",
			"validation_report":  report,
			"timestamp":          time.Now(),
		})
	}
}

// FixBalanceIssues provides an endpoint to manually fix balance issues
func (m *BalanceValidationMiddleware) FixBalanceIssues() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("🔧 Manual balance issue fix requested")

		// Validate first to show current state
		beforeReport, err := m.autoSyncService.ValidateBalanceConsistency()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to validate balances",
				"message": err.Error(),
			})
			return
		}

		if beforeReport.IsConsistent {
			c.JSON(http.StatusOK, gin.H{
				"status":  "no_action_needed",
				"message": "All balances are already consistent",
				"report":  beforeReport,
			})
			return
		}

		// Fix the issues
		if err := m.autoSyncService.FixAllBalanceIssues(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":        "Failed to fix balance issues",
				"message":      err.Error(),
				"before_fix":   beforeReport,
			})
			return
		}

		// Validate after fix
		afterReport, err := m.autoSyncService.ValidateBalanceConsistency()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "Fix completed but validation failed",
				"message":    err.Error(),
				"before_fix": beforeReport,
			})
			return
		}

		// Update cache
		m.lastValidationTime = time.Now()
		m.lastReport = afterReport

		c.JSON(http.StatusOK, gin.H{
			"status":     "success",
			"message":    "Balance issues fixed successfully",
			"before_fix": beforeReport,
			"after_fix":  afterReport,
			"timestamp":  time.Now(),
		})
	}
}

// GetBalanceValidationConfig returns the current validation configuration
func (m *BalanceValidationMiddleware) GetBalanceValidationConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"config": m.validationConfig,
		})
	}
}

// UpdateBalanceValidationConfig updates the validation configuration
func (m *BalanceValidationMiddleware) UpdateBalanceValidationConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		var newConfig BalanceValidationConfig
		if err := c.ShouldBindJSON(&newConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid configuration",
				"message": err.Error(),
			})
			return
		}

		// Update configuration
		oldConfig := *m.validationConfig
		m.validationConfig = &newConfig

		c.JSON(http.StatusOK, gin.H{
			"status":     "success",
			"message":    "Validation configuration updated",
			"old_config": oldConfig,
			"new_config": newConfig,
		})
	}
}