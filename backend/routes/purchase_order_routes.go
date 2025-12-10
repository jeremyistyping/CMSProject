package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupPurchaseOrderRoutes(router *gin.RouterGroup, db *gorm.DB, jwtManager *middleware.JWTManager) {
	// Initialize repositories
	poRepo := repositories.NewPurchaseOrderRepository(db)
	prRepo := repositories.NewPurchaseRequestRepository(db)

	// Initialize services
	poService := services.NewPurchaseOrderService(poRepo, prRepo)
	poPDFService := services.NewPOPDFService(poService)

	// Initialize controller
	poController := controllers.NewPurchaseOrderController(poService, poPDFService)

	// Routes
	po := router.Group("/purchase-orders")
	po.Use(jwtManager.AuthRequired())
	{
		// Create PO from approved PR
		po.POST("/from-pr", poController.CreateFromPR)

		// Get all POs
		po.GET("", poController.GetAll)

		// Get PO by ID
		po.GET("/:id", poController.GetByID)

		// Get POs by PR ID
		po.GET("/by-pr/:pr_id", poController.GetByPRID)

		// Send PO to vendor
		po.POST("/:id/send", poController.SendPO)

		// Delete PO
		po.DELETE("/:id", poController.Delete)

		// PDF Downloads
		po.GET("/:id/pdf", poController.DownloadPOPDF)
		po.GET("/:id/goods-receipt-pdf", poController.DownloadGRPDF)

		// Goods Receipt
		po.POST("/goods-receipt", poController.CreateGoodsReceipt)
		po.GET("/:id/goods-receipts", poController.GetGoodsReceiptsByPOID)
	}
}
