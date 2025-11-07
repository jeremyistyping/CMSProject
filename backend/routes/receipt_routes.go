package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupReceiptRoutes sets up all receipt-related routes
func SetupReceiptRoutes(protected *gin.RouterGroup, db *gorm.DB, jwtManager *middleware.JWTManager) {
	// Initialize Permission Middleware
	permMiddleware := middleware.NewPermissionMiddleware(db)
	
	// Initialize Receipt Controller
	receiptController := controllers.NewReceiptController(db)
	
	// Receipt routes
	receipts := protected.Group("/receipts")
	receipts.Use(jwtManager.AuthRequired())
	{
		// Basic CRUD operations
		receipts.GET("", permMiddleware.CanView("payments"), receiptController.GetReceipts)
		receipts.GET("/:id", permMiddleware.CanView("payments"), receiptController.GetReceipt)
		receipts.POST("", permMiddleware.CanCreate("payments"), receiptController.CreateReceipt)
		receipts.PUT("/:id", permMiddleware.CanEdit("payments"), receiptController.UpdateReceipt)
		receipts.DELETE("/:id", permMiddleware.CanDelete("payments"), receiptController.DeleteReceipt)
		
		// Receipt operations
		receipts.POST("/:id/print", permMiddleware.CanView("payments"), receiptController.PrintReceipt)
		receipts.GET("/:id/pdf", permMiddleware.CanExport("payments"), receiptController.GenerateReceiptPDF)
		
		// Bulk operations
		receipts.GET("/export", permMiddleware.CanExport("payments"), receiptController.ExportReceipts)
		receipts.POST("/bulk-print", permMiddleware.CanView("payments"), receiptController.BulkPrintReceipts)
		
		// Search and filter
		receipts.GET("/search", permMiddleware.CanView("payments"), receiptController.SearchReceipts)
		receipts.GET("/by-payment/:payment_id", permMiddleware.CanView("payments"), receiptController.GetReceiptsByPayment)
		receipts.GET("/by-date-range", permMiddleware.CanView("payments"), receiptController.GetReceiptsByDateRange)
	}
}