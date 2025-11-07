package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupCashBankIntegratedRoutes sets up integrated routes for CashBank and SSOT Journal system
func SetupCashBankIntegratedRoutes(
	router *gin.RouterGroup, 
	db *gorm.DB, 
	jwtManager *middleware.JWTManager,
) {
	// Initialize dependencies
	permissionMiddleware := middleware.NewPermissionMiddleware(db)
	
	// Initialize repositories
	accountRepo := repositories.NewAccountRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	
	// Initialize existing services (reuse pattern from main routes)
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)
	unifiedJournalService := services.NewUnifiedJournalService(db)
	
	// Initialize integrated service
	integratedService := services.NewCashBankIntegratedService(
		db,
		cashBankService,
		unifiedJournalService,
		accountRepo,
	)
	
	// Initialize integrated controller
	integratedController := controllers.NewCashBankIntegratedController(integratedService)
	
	// Setup integrated routes with proper middleware
	integrated := router.Group("/cashbank/integrated")
	integrated.Use(jwtManager.AuthRequired()) // Require authentication
	{
		// Enhanced account details with SSOT integration
		integrated.GET("/accounts/:id", 
			permissionMiddleware.CanView("cash_bank"), 
			integratedController.GetIntegratedAccountDetails)
		
		// Integrated summary for all cash/bank accounts
		integrated.GET("/summary", 
			permissionMiddleware.CanView("cash_bank"), 
			integratedController.GetIntegratedSummary)
		
		// Balance reconciliation between CashBank and SSOT
		integrated.GET("/accounts/:id/reconciliation", 
			permissionMiddleware.CanView("cash_bank"), 
			integratedController.GetAccountReconciliation)
		
		// Journal entries for specific cash/bank account
		integrated.GET("/accounts/:id/journal-entries", 
			permissionMiddleware.CanView("cash_bank"), 
			integratedController.GetAccountJournalEntries)
		
		// Transaction history with journal references
		integrated.GET("/accounts/:id/transactions", 
			permissionMiddleware.CanView("cash_bank"), 
			integratedController.GetAccountTransactionHistory)

		// Reconcile all balances (admin action)
		integrated.POST("/reconcile", 
			permissionMiddleware.CanEdit("cash_bank"),
			integratedController.Reconcile)
	}
}

// RegisterCashBankIntegratedRoutes registers the integrated routes in main routes setup
// This function should be called from the main routes.go file
func RegisterCashBankIntegratedRoutes(r *gin.Engine, db *gorm.DB) {
	// Initialize JWT manager
	jwtManager := middleware.NewJWTManager(db)
	
	// API v1 routes
	v1 := r.Group("/api/v1")
	
	// Setup integrated routes
	SetupCashBankIntegratedRoutes(v1, db, jwtManager)
}