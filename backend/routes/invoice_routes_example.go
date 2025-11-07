package routes

import (
	"github.com/gin-gonic/gin"
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/middleware"
	"gorm.io/gorm"
)

// SetupInvoiceRoutes registers all invoice-related routes
// This is an example of how to integrate invoice routes with settings
func SetupInvoiceRoutes(protected *gin.RouterGroup, db *gorm.DB) {
	// Initialize repositories
	contactRepo := repositories.NewContactRepository(db)
	productRepo := repositories.NewProductRepository(db)
	
	// Initialize services
	invoiceService := services.NewInvoiceServiceFull(db, contactRepo, productRepo)
	quoteService := services.NewQuoteServiceFull(db, contactRepo, productRepo)
	
	// Initialize controllers
	invoiceController := controllers.NewInvoiceController(invoiceService)
	quoteController := controllers.NewQuoteController(quoteService, invoiceService)

	// Initialize Permission Middleware
	permMiddleware := middleware.NewPermissionMiddleware(db)
	
	// Invoice routes (guarded by Sales module permissions)
	invoices := protected.Group("/invoices")
	{
		invoices.GET("", permMiddleware.CanView("sales"), invoiceController.GetInvoices)
		invoices.GET("/:id", permMiddleware.CanView("sales"), invoiceController.GetInvoice)
		invoices.POST("", permMiddleware.CanCreate("sales"), invoiceController.CreateInvoice)
		invoices.PUT("/:id", permMiddleware.CanEdit("sales"), invoiceController.UpdateInvoice)
		invoices.DELETE("/:id", permMiddleware.CanDelete("sales"), invoiceController.DeleteInvoice)
		
		// Utility endpoints
		invoices.POST("/generate-code", permMiddleware.CanCreate("sales"), invoiceController.GenerateInvoiceCode)
		invoices.POST("/format-currency", permMiddleware.CanView("sales"), invoiceController.FormatCurrency)
	}
	
	// Quote routes (guarded by Sales module permissions)
	quotes := protected.Group("/quotes")
	{
		quotes.GET("", permMiddleware.CanView("sales"), quoteController.GetQuotes)
		quotes.GET("/:id", permMiddleware.CanView("sales"), quoteController.GetQuote)
		quotes.POST("", permMiddleware.CanCreate("sales"), quoteController.CreateQuote)
		quotes.PUT("/:id", permMiddleware.CanEdit("sales"), quoteController.UpdateQuote)
		quotes.DELETE("/:id", permMiddleware.CanDelete("sales"), quoteController.DeleteQuote)
		
		// Utility endpoints
		quotes.POST("/generate-code", permMiddleware.CanCreate("sales"), quoteController.GenerateQuoteCode)
		quotes.POST("/format-currency", permMiddleware.CanView("sales"), quoteController.FormatCurrency)
		quotes.POST("/:id/convert-to-invoice", permMiddleware.CanCreate("sales"), quoteController.ConvertToInvoice)
	}
}
