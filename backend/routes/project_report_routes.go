package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupProjectReportRoutes - Setup routes untuk project-based reports
func SetupProjectReportRoutes(router gin.IRouter, db *gorm.DB, jwtManager *middleware.JWTManager) {
	// Initialize service
	projectReportService := services.NewProjectReportService(db)
	
	// Initialize controller
	projectReportController := controllers.NewProjectReportController(projectReportService)
	
	// Create reports group with authentication
	reportsGroup := router.Group("/project-reports")
	reportsGroup.Use(jwtManager.AuthRequired())
	
	// Available reports list
	reportsGroup.GET("/available", projectReportController.GetAvailableReports)
	
	// Budget vs Actual Report
	reportsGroup.GET("/budget-vs-actual", projectReportController.GetBudgetVsActual)
	
	// Profitability Report
	reportsGroup.GET("/profitability", projectReportController.GetProfitability)
	
	// Cash Flow Report
	reportsGroup.GET("/cash-flow", projectReportController.GetCashFlow)
	
	// Cost Summary Report
	reportsGroup.GET("/cost-summary", projectReportController.GetCostSummary)

	// Portfolio Budget vs Actual per Project
	reportsGroup.GET("/portfolio-budget-vs-actual", projectReportController.GetPortfolioBudgetVsActual)

	// Progress vs Cost per Project (time-series)
	reportsGroup.GET("/progress-vs-cost", projectReportController.GetProgressVsCost)
}
