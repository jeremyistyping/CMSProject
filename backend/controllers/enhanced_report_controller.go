package controllers

import (
	"net/http"
	"time"

	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// EnhancedReportController handles comprehensive financial and operational reporting endpoints
// Updated to support SSOT P&L integration
type EnhancedReportController struct {
	db *gorm.DB
}

// getSSOTReportIntegrationController wires SSOT integration service and controller on-demand
func (erc *EnhancedReportController) getSSOTReportIntegrationController() *SSOTReportIntegrationController {
	// Instantiate dependencies required by EnhancedReportService
	accountRepo := repositories.NewAccountRepository(erc.db)
	salesRepo := repositories.NewSalesRepository(erc.db)
	purchaseRepo := repositories.NewPurchaseRepository(erc.db)
	productRepo := repositories.NewProductRepository(erc.db)
	contactRepo := repositories.NewContactRepository(erc.db)
	paymentRepo := repositories.NewPaymentRepository(erc.db)
	cashBankRepo := repositories.NewCashBankRepository(erc.db)
	cacheService := services.NewReportCacheService()

	// Core services
	enhancedReportService := services.NewEnhancedReportService(
		erc.db,
		accountRepo,
		salesRepo,
		purchaseRepo,
		productRepo,
		contactRepo,
		paymentRepo,
		cashBankRepo,
		cacheService,
	)
	unifiedJournalService := services.NewUnifiedJournalService(erc.db)

	integrationService := services.NewSSOTReportIntegrationService(
		erc.db,
		unifiedJournalService,
		enhancedReportService,
	)
	return NewSSOTReportIntegrationController(integrationService, erc.db)
}

// NewEnhancedReportController creates a new enhanced report controller
func NewEnhancedReportController(db *gorm.DB) *EnhancedReportController {
	return &EnhancedReportController{
		db: db,
	}
}

// GetComprehensiveBalanceSheet returns comprehensive balance sheet data from SSOT journal system
func (erc *EnhancedReportController) GetComprehensiveBalanceSheet(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	asOfDate := c.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	
	// Create SSOT Balance Sheet controller and delegate to it
	// This integrates the SSOT journal system with the enhanced report controller
	ssotController := NewSSOTBalanceSheetController(erc.db)
	
	// Set the format in the query parameters for the SSOT controller
	c.Request.URL.RawQuery = "as_of_date=" + asOfDate + "&format=" + format
	ssotController.GenerateSSOTBalanceSheet(c)
}

// GetComprehensiveProfitLoss generates P&L report using SSOT journal system
func (erc *EnhancedReportController) GetComprehensiveProfitLoss(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	// Create SSOT P&L controller and delegate to it
	// This integrates the SSOT journal system with the enhanced report controller
	ssotController := NewSSOTProfitLossController(erc.db)
	
	// Set the format in the query parameters for the SSOT controller
	c.Request.URL.RawQuery = c.Request.URL.RawQuery + "&format=" + format
	ssotController.GetSSOTProfitLoss(c)
}

// GetComprehensiveCashFlow now delegates to SSOT Cash Flow controller
func (erc *EnhancedReportController) GetComprehensiveCashFlow(c *gin.Context) {
	// Delegate to SSOTCashFlowController to ensure SSOT integration
	ssotCF := NewSSOTCashFlowController(erc.db)
	ssotCF.GetSSOTCashFlow(c)
}

// GetComprehensiveSalesSummary now delegates to SSOT integration controller
func (erc *EnhancedReportController) GetComprehensiveSalesSummary(c *gin.Context) {
	ctrl := erc.getSSOTReportIntegrationController()
	ctrl.GetSSOTSalesSummary(c)
}

// GetComprehensivePurchaseSummary now delegates to SSOT Purchase Report controller
func (erc *EnhancedReportController) GetComprehensivePurchaseSummary(c *gin.Context) {
	purchaseCtrl := NewSSOTPurchaseReportController(erc.db)
	purchaseCtrl.GetPurchaseSummary(c)
}

// GetVendorAnalysis is deprecated; redirect users to SSOT Purchase Report endpoints
func (erc *EnhancedReportController) GetVendorAnalysis(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error":   "This endpoint has been replaced by /api/v1/ssot-reports/purchase-report",
		"message": "Vendor Analysis has been replaced with more credible Purchase Report",
		"new_endpoints": gin.H{
			"purchase_report":  "/api/v1/ssot-reports/purchase-report",
			"purchase_summary": "/api/v1/ssot-reports/purchase-summary",
		},
		"migration_guide": "Use the new Purchase Report for accurate vendor and purchase analysis",
	})
}

// GetTrialBalance now delegates to SSOT integration controller
func (erc *EnhancedReportController) GetTrialBalance(c *gin.Context) {
	ctrl := erc.getSSOTReportIntegrationController()
	ctrl.GetSSOTTrialBalance(c)
}

// GetGeneralLedger now delegates to SSOT integration controller
func (erc *EnhancedReportController) GetGeneralLedger(c *gin.Context) {
	ctrl := erc.getSSOTReportIntegrationController()
	ctrl.GetSSOTGeneralLedger(c)
}

// GetJournalEntryAnalysis now delegates to SSOT integration controller
func (erc *EnhancedReportController) GetJournalEntryAnalysis(c *gin.Context) {
	ctrl := erc.getSSOTReportIntegrationController()
	ctrl.GetSSOTJournalAnalysis(c)
}

// GetFinancialDashboard returns safe empty dashboard data
func (erc *EnhancedReportController) GetFinancialDashboard(c *gin.Context) {
	emptyDashboardData := gin.H{
		"period": gin.H{
			"start_date": time.Now().AddDate(0, -1, 0).Format("2006-01-02"),
			"end_date":   time.Now().Format("2006-01-02"),
		},
		"balance_sheet": gin.H{
			"total_assets":      0,
			"total_liabilities": 0,
			"total_equity":      0,
		},
		"profit_loss": gin.H{
			"total_revenue": 0,
			"net_income":    0,
		},
		"cash_flow": gin.H{
			"net_cash_flow": 0,
		},
		"key_ratios": gin.H{
			"current_ratio":   0,
			"debt_to_equity": 0,
		},
		"message": "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyDashboardData,
	})
}

// GetAvailableReports returns metadata about available reports (updated statuses)
func (erc *EnhancedReportController) GetAvailableReports(c *gin.Context) {
	reports := []gin.H{
		{
			"id":          "comprehensive_balance_sheet",
			"name":        "Balance Sheet",
			"type":        "FINANCIAL",
			"description": "Enhanced Balance Sheet from SSOT Journal System",
			"endpoint":    "/api/reports/balance-sheet",
			"status":      "SSOT_INTEGRATED",
		},
		{
			"id":          "comprehensive_profit_loss",
			"name":        "Profit & Loss Statement",
			"type":        "FINANCIAL",
			"description": "Enhanced P&L statement from SSOT Journal System",
			"endpoint":    "/api/reports/profit-loss",
			"status":      "SSOT_INTEGRATED",
		},
		{
			"id":          "comprehensive_cash_flow",
			"name":        "Cash Flow Statement",
			"type":        "FINANCIAL",
			"description": "SSOT-integrated cash flow",
			"endpoint":    "/api/reports/cash-flow",
			"status":      "SSOT_INTEGRATED",
		},
		{
			"id":          "trial_balance",
			"name":        "Trial Balance",
			"type":        "FINANCIAL",
			"description": "Trial Balance from SSOT Journal System",
			"endpoint":    "/api/reports/trial-balance",
			"status":      "SSOT_INTEGRATED",
		},
		{
			"id":          "general_ledger",
			"name":        "General Ledger",
			"type":        "FINANCIAL",
			"description": "General Ledger from SSOT Journal System",
			"endpoint":    "/api/reports/general-ledger",
			"status":      "SSOT_INTEGRATED",
		},
		{
			"id":          "journal_analysis",
			"name":        "Journal Entry Analysis",
			"type":        "FINANCIAL",
			"description": "Journal analysis from SSOT Journal System",
			"endpoint":    "/api/reports/journal-entry-analysis",
			"status":      "SSOT_INTEGRATED",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    reports,
		"message": "Reports are powered by SSOT integration",
	})
}

// GetReportPreview returns safe empty preview data
func (erc *EnhancedReportController) GetReportPreview(c *gin.Context) {
	reportType := c.Param("type")

	previewData := gin.H{
		"report_type":  reportType,
		"company_name": "Sistema Akuntansi",
		"preview_data": gin.H{},
		"message":      "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   previewData,
		"meta":   gin.H{"is_preview": true, "safe_mode": true},
	})
}

// GetAccountsReceivable returns safe empty accounts receivable data
func (erc *EnhancedReportController) GetAccountsReceivable(c *gin.Context) {
	emptyARData := gin.H{
		"report_date":        time.Now().Format("2006-01-02"),
		"company_name":       "Sistema Akuntansi",
		"total_receivables":  0,
		"customer_balances":  []gin.H{},
		"aging_summary":      gin.H{"current": 0, "30_days": 0, "60_days": 0, "90_days_plus": 0},
		"message":            "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyARData,
	})
}

// GetAccountsPayable returns safe empty accounts payable data
func (erc *EnhancedReportController) GetAccountsPayable(c *gin.Context) {
	emptyAPData := gin.H{
		"report_date":      time.Now().Format("2006-01-02"),
		"company_name":     "Sistema Akuntansi",
		"total_payables":   0,
		"vendor_balances":  []gin.H{},
		"aging_summary":    gin.H{"current": 0, "30_days": 0, "60_days": 0, "90_days_plus": 0},
		"message":          "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyAPData,
	})
}

// GetInventoryReport returns safe empty inventory report data
func (erc *EnhancedReportController) GetInventoryReport(c *gin.Context) {
	emptyInventoryData := gin.H{
		"report_date":       time.Now().Format("2006-01-02"),
		"company_name":      "Sistema Akuntansi",
		"total_inventory_value": 0,
		"products":          []gin.H{},
		"categories":        []gin.H{},
		"low_stock_items":   []gin.H{},
		"message":           "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyInventoryData,
	})
}

// GetFinancialRatios returns safe empty financial ratios data
func (erc *EnhancedReportController) GetFinancialRatios(c *gin.Context) {
	emptyRatiosData := gin.H{
		"report_date":     time.Now().Format("2006-01-02"),
		"company_name":    "Sistema Akuntansi",
		"liquidity_ratios": gin.H{"current_ratio": 0, "quick_ratio": 0, "cash_ratio": 0},
		"leverage_ratios":  gin.H{"debt_to_equity": 0, "debt_to_assets": 0, "interest_coverage": 0},
		"efficiency_ratios": gin.H{"inventory_turnover": 0, "receivables_turnover": 0, "asset_turnover": 0},
		"profitability_ratios": gin.H{"gross_margin": 0, "net_margin": 0, "roa": 0, "roe": 0},
		"message":         "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyRatiosData,
	})
}

// GetReportValidation returns safe validation data
func (erc *EnhancedReportController) GetReportValidation(c *gin.Context) {
	validationReport := gin.H{
		"validation_date": time.Now().Format("2006-01-02"),
		"company_name":    "Sistema Akuntansi",
		"status":          "SAFE_MODE",
		"checks":          []gin.H{},
		"message":         "Report module is in safe mode - validation disabled",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   validationReport,
	})
}
