package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/services"
)

// RegisterContactHistoryRoutes registers routes for contact history reports
func RegisterContactHistoryRoutes(router *gin.Engine, db *gorm.DB, pdfService services.PDFServiceInterface) {
	// Initialize controller
	contactHistoryController := controllers.NewContactHistoryController(db, pdfService)
	
	// API v1 routes
	api := router.Group("/api/v1")
	{
		reports := api.Group("/reports")
		{
			// Customer History
			reports.GET("/customer-history", contactHistoryController.GetCustomerHistory)
			
			// Vendor History
			reports.GET("/vendor-history", contactHistoryController.GetVendorHistory)
		}
	}
}
