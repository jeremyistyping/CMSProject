package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ProjectReportController struct {
	service *services.ProjectReportService
}

func NewProjectReportController(service *services.ProjectReportService) *ProjectReportController {
	return &ProjectReportController{service: service}
}

// parseReportParams - Helper to parse common report parameters
func (c *ProjectReportController) parseReportParams(ctx *gin.Context) (models.ProjectReportParams, error) {
	var params models.ProjectReportParams

	// Parse start_date
	startDateStr := ctx.Query("start_date")
	if startDateStr == "" {
		// Default to first day of current month
		now := time.Now()
		params.StartDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	} else {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return params, err
		}
		params.StartDate = startDate
	}

	// Parse end_date
	endDateStr := ctx.Query("end_date")
	if endDateStr == "" {
		// Default to today
		params.EndDate = time.Now()
	} else {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return params, err
		}
		params.EndDate = endDate
	}

	// Parse project_id (optional)
	projectIDStr := ctx.Query("project_id")
	if projectIDStr != "" {
		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			return params, err
		}
		id := uint(projectID)
		params.ProjectID = &id
	}

	// Parse format
	params.Format = ctx.DefaultQuery("format", "json")

	return params, nil
}

// GetBudgetVsActual - GET /api/v1/project-reports/budget-vs-actual
func (c *ProjectReportController) GetBudgetVsActual(ctx *gin.Context) {
	params, err := c.parseReportParams(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid parameters: " + err.Error(),
		})
		return
	}

	report, err := c.service.GenerateBudgetVsActualReport(params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate report: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetProfitability - GET /api/v1/project-reports/profitability
func (c *ProjectReportController) GetProfitability(ctx *gin.Context) {
	params, err := c.parseReportParams(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid parameters: " + err.Error(),
		})
		return
	}

	report, err := c.service.GenerateProfitabilityReport(params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate report: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetCashFlow - GET /api/v1/project-reports/cash-flow
func (c *ProjectReportController) GetCashFlow(ctx *gin.Context) {
	params, err := c.parseReportParams(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid parameters: " + err.Error(),
		})
		return
	}

	report, err := c.service.GenerateCashFlowReport(params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate report: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetCostSummary - GET /api/v1/project-reports/cost-summary
func (c *ProjectReportController) GetCostSummary(ctx *gin.Context) {
	params, err := c.parseReportParams(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid parameters: " + err.Error(),
		})
		return
	}

	report, err := c.service.GenerateCostSummaryReport(params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate report: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetAvailableReports - GET /api/v1/project-reports/available
func (c *ProjectReportController) GetAvailableReports(ctx *gin.Context) {
	reports := []gin.H{
		{
			"id":          "budget-vs-actual",
			"name":        "Budget vs Actual by COA Group",
			"description": "Menampilkan total estimasi vs realisasi per akun",
			"endpoint":    "/api/v1/project-reports/budget-vs-actual",
			"parameters": []string{
				"start_date (required)",
				"end_date (required)",
				"project_id (optional)",
			},
			"type": "PROJECT",
		},
		{
			"id":          "profitability",
			"name":        "Profitability Report per Project",
			"description": "(Pendapatan) – (Total Beban Langsung + Operasional)",
			"endpoint":    "/api/v1/project-reports/profitability",
			"parameters": []string{
				"start_date (required)",
				"end_date (required)",
				"project_id (optional)",
			},
			"type": "PROJECT",
		},
		{
			"id":          "cash-flow",
			"name":        "Cash Flow per Project",
			"description": "Dari kas masuk & kas keluar sesuai COA tipe Asset/Expense",
			"endpoint":    "/api/v1/project-reports/cash-flow",
			"parameters": []string{
				"start_date (required)",
				"end_date (required)",
				"project_id (optional)",
			},
			"type": "PROJECT",
		},
		{
			"id":          "cost-summary",
			"name":        "Cost Summary Report",
			"description": "Rekap per kategori (Material, Sewa, Labour, dll)",
			"endpoint":    "/api/v1/project-reports/cost-summary",
			"parameters": []string{
				"start_date (required)",
				"end_date (required)",
				"project_id (optional)",
			},
			"type": "PROJECT",
		},
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   reports,
	})
}
