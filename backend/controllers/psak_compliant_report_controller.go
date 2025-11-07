package controllers

import (
	"net/http"
	"time"

	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"
	"github.com/gin-gonic/gin"
)

// PSAKCompliantReportController handles PSAK-compliant financial reports
type PSAKCompliantReportController struct {
	psakReportService *services.PSAKCompliantReportService
}

// NewPSAKCompliantReportController creates a new PSAK report controller
func NewPSAKCompliantReportController(psakReportService *services.PSAKCompliantReportService) *PSAKCompliantReportController {
	return &PSAKCompliantReportController{
		psakReportService: psakReportService,
	}
}

// PSAKBalanceSheetRequest represents request for PSAK-compliant balance sheet
type PSAKBalanceSheetRequest struct {
	AsOfDate string `json:"as_of_date" form:"as_of_date" binding:"required" example:"2024-12-31"`
}

// PSAKProfitLossRequest represents request for PSAK-compliant P&L
type PSAKProfitLossRequest struct {
	StartDate string `json:"start_date" form:"start_date" binding:"required" example:"2024-01-01"`
	EndDate   string `json:"end_date" form:"end_date" binding:"required" example:"2024-12-31"`
}

// PSAKCashFlowRequest represents request for PSAK-compliant cash flow
type PSAKCashFlowRequest struct {
	StartDate string `json:"start_date" form:"start_date" binding:"required" example:"2024-01-01"`
	EndDate   string `json:"end_date" form:"end_date" binding:"required" example:"2024-12-31"`
	Method    string `json:"method" form:"method" example:"INDIRECT"` // DIRECT or INDIRECT
}

// PSAKComplianceCheckRequest represents request for PSAK compliance check
type PSAKComplianceCheckRequest struct {
	ReportType string `json:"report_type" form:"report_type" binding:"required" example:"BALANCE_SHEET"` // BALANCE_SHEET, PROFIT_LOSS, CASH_FLOW
	AsOfDate   string `json:"as_of_date" form:"as_of_date" example:"2024-12-31"`
	StartDate  string `json:"start_date" form:"start_date" example:"2024-01-01"`
	EndDate    string `json:"end_date" form:"end_date" example:"2024-12-31"`
}

// Standard response wrapper
type PSAKReportResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// GetPSAKBalanceSheet godoc
// @Summary      Generate PSAK-compliant Balance Sheet
// @Description  Generate Balance Sheet sesuai PSAK 1 (Penyajian Laporan Keuangan)
// @Tags         PSAK Reports
// @Accept       json
// @Produce      json
// @Param        request body PSAKBalanceSheetRequest true "Balance Sheet Request"
// @Success      200 {object} PSAKReportResponse{data=services.PSAKBalanceSheetData}
// @Failure      400 {object} PSAKReportResponse
// @Failure      500 {object} PSAKReportResponse
// @Security     BearerAuth
// @Router       /api/v1/reports/psak/balance-sheet [post]
func (ctrl *PSAKCompliantReportController) GetPSAKBalanceSheet(c *gin.Context) {
	var req PSAKBalanceSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := utils.NewBadRequestError("Invalid request format: " + err.Error())
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Parse date
	asOfDate, err := time.Parse("2006-01-02", req.AsOfDate)
	if err != nil {
		appErr := utils.NewBadRequestError("Invalid date format: Date must be in YYYY-MM-DD format")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Generate PSAK-compliant balance sheet
	balanceSheet, err := ctrl.psakReportService.GeneratePSAKBalanceSheet(asOfDate)
	if err != nil {
		appErr := utils.NewInternalError("Failed to generate PSAK balance sheet", err)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	response := PSAKReportResponse{
		Success:   true,
		Message:   "PSAK-compliant Balance Sheet generated successfully",
		Data:      balanceSheet,
		Timestamp: time.Now(),
		RequestID: c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}

// GetPSAKProfitLoss godoc
// @Summary      Generate PSAK-compliant Profit & Loss Statement
// @Description  Generate P&L Statement sesuai PSAK 1 (Laporan Laba Rugi dan Penghasilan Komprehensif Lain)
// @Tags         PSAK Reports
// @Accept       json
// @Produce      json
// @Param        request body PSAKProfitLossRequest true "Profit & Loss Request"
// @Success      200 {object} PSAKReportResponse{data=services.PSAKProfitLossData}
// @Failure      400 {object} PSAKReportResponse
// @Failure      500 {object} PSAKReportResponse
// @Security     BearerAuth
// @Router       /api/v1/reports/psak/profit-loss [post]
func (ctrl *PSAKCompliantReportController) GetPSAKProfitLoss(c *gin.Context) {
	var req PSAKProfitLossRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := utils.NewBadRequestError("Invalid request format: " + err.Error())
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		appErr := utils.NewBadRequestError("Invalid start date format: Date must be in YYYY-MM-DD format")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		appErr := utils.NewBadRequestError("Invalid end date format: Date must be in YYYY-MM-DD format")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Validate date range
	if endDate.Before(startDate) {
		appErr := utils.NewBadRequestError("Invalid date range: End date must be after start date")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Generate PSAK-compliant P&L
	profitLoss, err := ctrl.psakReportService.GeneratePSAKProfitLoss(startDate, endDate)
	if err != nil {
		appErr := utils.NewInternalError("Failed to generate PSAK profit & loss", err)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	response := PSAKReportResponse{
		Success:   true,
		Message:   "PSAK-compliant Profit & Loss Statement generated successfully",
		Data:      profitLoss,
		Timestamp: time.Now(),
		RequestID: c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}

// GetPSAKCashFlow godoc
// @Summary      Generate PSAK-compliant Cash Flow Statement
// @Description  Generate Cash Flow Statement sesuai PSAK 2 (Laporan Arus Kas)
// @Tags         PSAK Reports
// @Accept       json
// @Produce      json
// @Param        request body PSAKCashFlowRequest true "Cash Flow Request"
// @Success      200 {object} PSAKReportResponse{data=services.PSAKCashFlowData}
// @Failure      400 {object} PSAKReportResponse
// @Failure      500 {object} PSAKReportResponse
// @Security     BearerAuth
// @Router       /api/v1/reports/psak/cash-flow [post]
func (ctrl *PSAKCompliantReportController) GetPSAKCashFlow(c *gin.Context) {
	var req PSAKCashFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := utils.NewBadRequestError("Invalid request format: " + err.Error())
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		appErr := utils.NewBadRequestError("Invalid start date format: Date must be in YYYY-MM-DD format")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		appErr := utils.NewBadRequestError("Invalid end date format: Date must be in YYYY-MM-DD format")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Validate date range
	if endDate.Before(startDate) {
		appErr := utils.NewBadRequestError("Invalid date range: End date must be after start date")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Validate method
	method := req.Method
	if method == "" {
		method = "INDIRECT" // Default to indirect method
	}
	if method != "DIRECT" && method != "INDIRECT" {
		appErr := utils.NewBadRequestError("Invalid method: Method must be DIRECT or INDIRECT")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Generate PSAK-compliant Cash Flow
	cashFlow, err := ctrl.psakReportService.GeneratePSAKCashFlow(startDate, endDate, method)
	if err != nil {
		appErr := utils.NewInternalError("Failed to generate PSAK cash flow", err)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	response := PSAKReportResponse{
		Success:   true,
		Message:   "PSAK-compliant Cash Flow Statement generated successfully",
		Data:      cashFlow,
		Timestamp: time.Now(),
		RequestID: c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}

// GetPSAKComplianceSummary godoc
// @Summary      Get PSAK Compliance Summary for all reports
// @Description  Generate compliance summary untuk semua laporan keuangan terhadap standar PSAK
// @Tags         PSAK Reports
// @Accept       json
// @Produce      json
// @Param        as_of_date query string false "As of date for balance sheet (YYYY-MM-DD)" example(2024-12-31)
// @Param        start_date query string false "Start date for P&L and cash flow (YYYY-MM-DD)" example(2024-01-01)
// @Param        end_date query string false "End date for P&L and cash flow (YYYY-MM-DD)" example(2024-12-31)
// @Success      200 {object} PSAKReportResponse{data=PSAKComplianceSummaryResponse}
// @Failure      400 {object} PSAKReportResponse
// @Failure      500 {object} PSAKReportResponse
// @Security     BearerAuth
// @Router       /api/v1/reports/psak/compliance-summary [get]
func (ctrl *PSAKCompliantReportController) GetPSAKComplianceSummary(c *gin.Context) {
	// Get dates from query params
	asOfDateStr := c.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -365).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	// Parse dates
	asOfDate, err := time.Parse("2006-01-02", asOfDateStr)
	if err != nil {
		appErr := utils.NewBadRequestError("Invalid as_of_date format: Date must be in YYYY-MM-DD format")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		appErr := utils.NewBadRequestError("Invalid start_date format: Date must be in YYYY-MM-DD format")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		appErr := utils.NewBadRequestError("Invalid end_date format: Date must be in YYYY-MM-DD format")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Generate compliance summary for all reports
	summary, err := ctrl.generateComplianceSummary(asOfDate, startDate, endDate)
	if err != nil {
		appErr := utils.NewInternalError("Failed to generate PSAK compliance summary", err)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	response := PSAKReportResponse{
		Success:   true,
		Message:   "PSAK compliance summary generated successfully",
		Data:      summary,
		Timestamp: time.Now(),
		RequestID: c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}

// CheckPSAKCompliance godoc
// @Summary      Check PSAK compliance for specific report
// @Description  Check compliance terhadap standar PSAK untuk laporan tertentu
// @Tags         PSAK Reports
// @Accept       json
// @Produce      json
// @Param        request body PSAKComplianceCheckRequest true "Compliance Check Request"
// @Success      200 {object} PSAKReportResponse{data=services.PSAKComplianceInfo}
// @Failure      400 {object} PSAKReportResponse
// @Failure      500 {object} PSAKReportResponse
// @Security     BearerAuth
// @Router       /api/v1/reports/psak/check-compliance [post]
func (ctrl *PSAKCompliantReportController) CheckPSAKCompliance(c *gin.Context) {
	var req PSAKComplianceCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := utils.NewBadRequestError("Invalid request format: " + err.Error())
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Validate report type
	validTypes := map[string]bool{
		"BALANCE_SHEET": true,
		"PROFIT_LOSS":   true,
		"CASH_FLOW":     true,
	}
	
	if !validTypes[req.ReportType] {
		appErr := utils.NewBadRequestError("Invalid report type: Report type must be BALANCE_SHEET, PROFIT_LOSS, or CASH_FLOW")
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	var compliance *services.PSAKComplianceInfo

	switch req.ReportType {
	case "BALANCE_SHEET":
		if req.AsOfDate == "" {
			appErr := utils.NewBadRequestError("Missing as_of_date: as_of_date is required for balance sheet compliance check")
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		asOfDate, parseErr := time.Parse("2006-01-02", req.AsOfDate)
		if parseErr != nil {
			appErr := utils.NewBadRequestError("Invalid as_of_date format: Date must be in YYYY-MM-DD format")
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		
		balanceSheet, genErr := ctrl.psakReportService.GeneratePSAKBalanceSheet(asOfDate)
		if genErr != nil {
			appErr := utils.NewInternalError("Failed to generate balance sheet for compliance check", genErr)
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		compliance = &balanceSheet.PSAKCompliance

	case "PROFIT_LOSS":
		if req.StartDate == "" || req.EndDate == "" {
			appErr := utils.NewBadRequestError("Missing date range: start_date and end_date are required for P&L compliance check")
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		
		startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
		if parseErr != nil {
			appErr := utils.NewBadRequestError("Invalid start_date format: Date must be in YYYY-MM-DD format")
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		
		endDate, parseErr := time.Parse("2006-01-02", req.EndDate)
		if parseErr != nil {
			appErr := utils.NewBadRequestError("Invalid end_date format: Date must be in YYYY-MM-DD format")
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		
		profitLoss, genErr := ctrl.psakReportService.GeneratePSAKProfitLoss(startDate, endDate)
		if genErr != nil {
			appErr := utils.NewInternalError("Failed to generate P&L for compliance check", genErr)
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		compliance = &profitLoss.PSAKCompliance

	case "CASH_FLOW":
		if req.StartDate == "" || req.EndDate == "" {
			appErr := utils.NewBadRequestError("Missing date range: start_date and end_date are required for cash flow compliance check")
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		
		startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
		if parseErr != nil {
			appErr := utils.NewBadRequestError("Invalid start_date format: Date must be in YYYY-MM-DD format")
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		
		endDate, parseErr := time.Parse("2006-01-02", req.EndDate)
		if parseErr != nil {
			appErr := utils.NewBadRequestError("Invalid end_date format: Date must be in YYYY-MM-DD format")
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		
		cashFlow, genErr := ctrl.psakReportService.GeneratePSAKCashFlow(startDate, endDate, "INDIRECT")
		if genErr != nil {
			appErr := utils.NewInternalError("Failed to generate cash flow for compliance check", genErr)
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
			return
		}
		compliance = &cashFlow.PSAKCompliance
	}

	response := PSAKReportResponse{
		Success:   true,
		Message:   "PSAK compliance check completed successfully",
		Data:      compliance,
		Timestamp: time.Now(),
		RequestID: c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}

// GetPSAKStandardsList godoc
// @Summary      Get list of supported PSAK standards
// @Description  Get daftar standar PSAK yang didukung dalam sistem
// @Tags         PSAK Reports
// @Accept       json
// @Produce      json
// @Success      200 {object} PSAKReportResponse{data=PSAKStandardsListResponse}
// @Security     BearerAuth
// @Router       /api/v1/reports/psak/standards [get]
func (ctrl *PSAKCompliantReportController) GetPSAKStandardsList(c *gin.Context) {
	standards := ctrl.getSupportedPSAKStandards()

	response := PSAKReportResponse{
		Success:   true,
		Message:   "PSAK standards list retrieved successfully",
		Data:      standards,
		Timestamp: time.Now(),
		RequestID: c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}

// Supporting response types

type PSAKComplianceSummaryResponse struct {
	OverallCompliance   string                           `json:"overall_compliance"`
	OverallScore        float64                          `json:"overall_score"`
	ReportCompliances   map[string]*services.PSAKComplianceInfo `json:"report_compliances"`
	GeneratedAt         time.Time                        `json:"generated_at"`
	Period              PSAKPeriodInfo                   `json:"period"`
	CriticalIssues      []services.PSAKIssue            `json:"critical_issues,omitempty"`
	Recommendations     []string                         `json:"recommendations,omitempty"`
}

type PSAKPeriodInfo struct {
	AsOfDate  string `json:"as_of_date"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type PSAKStandardsListResponse struct {
	Standards []PSAKStandardInfo `json:"standards"`
	TotalCount int               `json:"total_count"`
}

type PSAKStandardInfo struct {
	Code         string   `json:"code"`        // e.g., "PSAK 1"
	Title        string   `json:"title"`       // e.g., "Penyajian Laporan Keuangan"
	Description  string   `json:"description"`
	ApplicableTo []string `json:"applicable_to"` // e.g., ["BALANCE_SHEET", "PROFIT_LOSS"]
	Status       string   `json:"status"`        // e.g., "ACTIVE", "SUPERSEDED"
	EffectiveDate string  `json:"effective_date"`
}

// Helper methods

func (ctrl *PSAKCompliantReportController) generateComplianceSummary(asOfDate, startDate, endDate time.Time) (*PSAKComplianceSummaryResponse, error) {
	summary := &PSAKComplianceSummaryResponse{
		ReportCompliances: make(map[string]*services.PSAKComplianceInfo),
		GeneratedAt:       time.Now(),
		Period: PSAKPeriodInfo{
			AsOfDate:  asOfDate.Format("2006-01-02"),
			StartDate: startDate.Format("2006-01-02"),
			EndDate:   endDate.Format("2006-01-02"),
		},
	}

	var totalScore float64
	var reportCount int
	var allCriticalIssues []services.PSAKIssue
	var allRecommendations []string

	// Check Balance Sheet compliance
	if balanceSheet, err := ctrl.psakReportService.GeneratePSAKBalanceSheet(asOfDate); err == nil {
		summary.ReportCompliances["BALANCE_SHEET"] = &balanceSheet.PSAKCompliance
		score, _ := balanceSheet.PSAKCompliance.ComplianceScore.Float64()
		totalScore += score
		reportCount++

		// Collect critical issues
		for _, issue := range balanceSheet.PSAKCompliance.NonComplianceIssues {
			if issue.Severity == "HIGH" {
				allCriticalIssues = append(allCriticalIssues, issue)
			}
		}
		allRecommendations = append(allRecommendations, balanceSheet.PSAKCompliance.Recommendations...)
	}

	// Check P&L compliance
	if profitLoss, err := ctrl.psakReportService.GeneratePSAKProfitLoss(startDate, endDate); err == nil {
		summary.ReportCompliances["PROFIT_LOSS"] = &profitLoss.PSAKCompliance
		score, _ := profitLoss.PSAKCompliance.ComplianceScore.Float64()
		totalScore += score
		reportCount++

		// Collect critical issues
		for _, issue := range profitLoss.PSAKCompliance.NonComplianceIssues {
			if issue.Severity == "HIGH" {
				allCriticalIssues = append(allCriticalIssues, issue)
			}
		}
		allRecommendations = append(allRecommendations, profitLoss.PSAKCompliance.Recommendations...)
	}

	// Check Cash Flow compliance
	if cashFlow, err := ctrl.psakReportService.GeneratePSAKCashFlow(startDate, endDate, "INDIRECT"); err == nil {
		summary.ReportCompliances["CASH_FLOW"] = &cashFlow.PSAKCompliance
		score, _ := cashFlow.PSAKCompliance.ComplianceScore.Float64()
		totalScore += score
		reportCount++

		// Collect critical issues
		for _, issue := range cashFlow.PSAKCompliance.NonComplianceIssues {
			if issue.Severity == "HIGH" {
				allCriticalIssues = append(allCriticalIssues, issue)
			}
		}
		allRecommendations = append(allRecommendations, cashFlow.PSAKCompliance.Recommendations...)
	}

	// Calculate overall compliance
	if reportCount > 0 {
		summary.OverallScore = totalScore / float64(reportCount)
		
		switch {
		case summary.OverallScore >= 95:
			summary.OverallCompliance = "FULL"
		case summary.OverallScore >= 80:
			summary.OverallCompliance = "SUBSTANTIAL"
		case summary.OverallScore >= 60:
			summary.OverallCompliance = "PARTIAL"
		default:
			summary.OverallCompliance = "NON_COMPLIANT"
		}
	}

	summary.CriticalIssues = allCriticalIssues
	summary.Recommendations = removeDuplicateStrings(allRecommendations)

	return summary, nil
}

func (ctrl *PSAKCompliantReportController) getSupportedPSAKStandards() *PSAKStandardsListResponse {
	standards := []PSAKStandardInfo{
		{
			Code:          "PSAK 1",
			Title:         "Penyajian Laporan Keuangan",
			Description:   "Standar yang mengatur penyajian laporan keuangan untuk tujuan umum",
			ApplicableTo:  []string{"BALANCE_SHEET", "PROFIT_LOSS"},
			Status:        "ACTIVE",
			EffectiveDate: "2015-01-01",
		},
		{
			Code:          "PSAK 2",
			Title:         "Laporan Arus Kas",
			Description:   "Standar yang mengatur penyajian laporan arus kas",
			ApplicableTo:  []string{"CASH_FLOW"},
			Status:        "ACTIVE",
			EffectiveDate: "2015-01-01",
		},
		{
			Code:          "PSAK 14",
			Title:         "Persediaan",
			Description:   "Standar yang mengatur akuntansi untuk persediaan",
			ApplicableTo:  []string{"BALANCE_SHEET", "PROFIT_LOSS"},
			Status:        "ACTIVE",
			EffectiveDate: "2015-01-01",
		},
		{
			Code:          "PSAK 16",
			Title:         "Aset Tetap",
			Description:   "Standar yang mengatur akuntansi untuk aset tetap",
			ApplicableTo:  []string{"BALANCE_SHEET"},
			Status:        "ACTIVE",
			EffectiveDate: "2015-01-01",
		},
		{
			Code:          "PSAK 23",
			Title:         "Pendapatan dari Kontrak dengan Pelanggan",
			Description:   "Standar yang mengatur pengakuan pendapatan",
			ApplicableTo:  []string{"PROFIT_LOSS"},
			Status:        "ACTIVE",
			EffectiveDate: "2020-01-01",
		},
		{
			Code:          "PSAK 46",
			Title:         "Pajak Penghasilan",
			Description:   "Standar yang mengatur akuntansi untuk pajak penghasilan",
			ApplicableTo:  []string{"BALANCE_SHEET", "PROFIT_LOSS"},
			Status:        "ACTIVE",
			EffectiveDate: "2015-01-01",
		},
		{
			Code:          "PSAK 56",
			Title:         "Laba per Saham",
			Description:   "Standar yang mengatur penghitungan dan penyajian laba per saham",
			ApplicableTo:  []string{"PROFIT_LOSS"},
			Status:        "ACTIVE",
			EffectiveDate: "2015-01-01",
		},
	}

	return &PSAKStandardsListResponse{
		Standards:  standards,
		TotalCount: len(standards),
	}
}

// Utility functions
func removeDuplicateStrings(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}