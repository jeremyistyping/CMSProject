package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterEnhancedReportRoutes registers SAFE financial reporting routes
// SIMPLIFIED VERSION: Safe routes that don't cause crashes
func RegisterEnhancedReportRoutes(router gin.IRouter, enhancedReportController *controllers.EnhancedReportController, jwtManager *middleware.JWTManager) {
	// Create report routes group with minimal authentication
	reportsGroup := router.Group("/reports")
	reportsGroup.Use(jwtManager.AuthRequired())

	// SIMPLIFIED ROUTES - Basic endpoints that return empty data safely
	// These routes are safe and won't cause crashes
	reportsGroup.GET("/balance-sheet", enhancedReportController.GetComprehensiveBalanceSheet)
	reportsGroup.GET("/profit-loss", enhancedReportController.GetComprehensiveProfitLoss)
	reportsGroup.GET("/cash-flow", enhancedReportController.GetComprehensiveCashFlow)
	reportsGroup.GET("/sales-summary", enhancedReportController.GetComprehensiveSalesSummary)
	reportsGroup.GET("/purchase-summary", enhancedReportController.GetComprehensivePurchaseSummary)
	reportsGroup.GET("/vendor-analysis", enhancedReportController.GetVendorAnalysis)
	reportsGroup.GET("/trial-balance", enhancedReportController.GetTrialBalance)
	reportsGroup.GET("/general-ledger", enhancedReportController.GetGeneralLedger)
	reportsGroup.GET("/journal-entry-analysis", enhancedReportController.GetJournalEntryAnalysis)
	
	// Additional report endpoints
	reportsGroup.GET("/accounts-receivable", enhancedReportController.GetAccountsReceivable)
	reportsGroup.GET("/accounts-payable", enhancedReportController.GetAccountsPayable)
	reportsGroup.GET("/inventory-report", enhancedReportController.GetInventoryReport)
	reportsGroup.GET("/financial-ratios", enhancedReportController.GetFinancialRatios)

	// Dashboard and metadata endpoints
	reportsGroup.GET("/financial-dashboard", enhancedReportController.GetFinancialDashboard)
	reportsGroup.GET("/available", enhancedReportController.GetAvailableReports)
	reportsGroup.GET("/preview/:type", enhancedReportController.GetReportPreview)
	reportsGroup.GET("/validation", enhancedReportController.GetReportValidation)

	// Comprehensive group (optional for backward compatibility)
	comprehensiveGroup := reportsGroup.Group("/comprehensive")
	{
		comprehensiveGroup.GET("/balance-sheet", enhancedReportController.GetComprehensiveBalanceSheet)
		comprehensiveGroup.GET("/profit-loss", enhancedReportController.GetComprehensiveProfitLoss)
		comprehensiveGroup.GET("/cash-flow", enhancedReportController.GetComprehensiveCashFlow)
		comprehensiveGroup.GET("/sales-summary", enhancedReportController.GetComprehensiveSalesSummary)
		comprehensiveGroup.GET("/purchase-summary", enhancedReportController.GetComprehensivePurchaseSummary)
		comprehensiveGroup.GET("/vendor-analysis", enhancedReportController.GetVendorAnalysis)
		comprehensiveGroup.GET("/trial-balance", enhancedReportController.GetTrialBalance)
		comprehensiveGroup.GET("/general-ledger", enhancedReportController.GetGeneralLedger)
		comprehensiveGroup.GET("/journal-analysis", enhancedReportController.GetJournalEntryAnalysis)
	}
}

// RegisterEnhancedReportRoutesWithPrefix registers enhanced report routes with a custom prefix
// SIMPLIFIED VERSION: Safe routes with custom prefix
func RegisterEnhancedReportRoutesWithPrefix(router gin.IRouter, prefix string, enhancedReportController *controllers.EnhancedReportController, jwtManager *middleware.JWTManager) {
	// Create report routes group with custom prefix
	reportsGroup := router.Group(prefix)
	reportsGroup.Use(jwtManager.AuthRequired())

	// Safe financial statements
	reportsGroup.GET("/balance-sheet", enhancedReportController.GetComprehensiveBalanceSheet)
	reportsGroup.GET("/profit-loss", enhancedReportController.GetComprehensiveProfitLoss)
	reportsGroup.GET("/cash-flow", enhancedReportController.GetComprehensiveCashFlow)
	
	// Safe operational reports
	reportsGroup.GET("/sales-summary", enhancedReportController.GetComprehensiveSalesSummary)
	reportsGroup.GET("/purchase-summary", enhancedReportController.GetComprehensivePurchaseSummary)
	
	// Safe dashboard and metadata
	reportsGroup.GET("/dashboard", enhancedReportController.GetFinancialDashboard)
	reportsGroup.GET("/available", enhancedReportController.GetAvailableReports)
}

// RegisterEnhancedReportMiddleware registers minimal middleware for enhanced reporting
func RegisterEnhancedReportMiddleware(router gin.IRouter) {
	// Add minimal report-specific middleware
	router.Use(func(c *gin.Context) {
		// Add safe headers for reports
		if c.Request.URL.Path == "/api/reports" || c.Request.URL.Path == "/api/reports/" {
			c.Header("X-Report-Version", "2.0-safe")
			c.Header("X-Report-Mode", "safe")
		}
		c.Next()
	})
}