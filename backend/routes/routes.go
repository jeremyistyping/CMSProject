package routes

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/handlers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Environment detection helper
func getEnvironment() string {
	env := strings.ToLower(os.Getenv("ENV"))
	if env == "" {
		env = strings.ToLower(os.Getenv("GO_ENV"))
	}
	if env == "" {
		env = strings.ToLower(os.Getenv("ENVIRONMENT"))
	}
	if env == "" {
		env = "development"
	}
	return env
}

// Check if development features should be enabled
func isDevelopmentMode() bool {
	env := getEnvironment()
	return env == "development" || env == "dev" || env == "local"
}

// Check if debug routes should be enabled
func shouldEnableDebugRoutes() bool {
	return os.Getenv("ENABLE_DEBUG_ROUTES") == "true" && isDevelopmentMode()
}

// loadSwaggerJSON tries multiple locations to find a Swagger/OpenAPI JSON
func loadSwaggerJSON() ([]byte, string, error) {
	candidates := []string{
		"./docs/swagger.json",
		"./backend/docs/swagger.json",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			b, err := os.ReadFile(p)
			if err != nil {
				return nil, p, err
			}
			return b, p, nil
		}
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		try1 := filepath.Join(dir, "docs", "swagger.json")
		if _, err := os.Stat(try1); err == nil {
			b, err := os.ReadFile(try1)
			if err != nil {
				return nil, try1, err
			}
			return b, try1, nil
		}
	}
	return nil, "", fmt.Errorf("swagger spec not found")
}

// minimalSwaggerJSON returns a minimal OpenAPI document
func minimalSwaggerJSON() []byte {
	return []byte(`{
  "openapi": "3.0.3",
  "info": {
    "title": "Unipro Project Management API",
    "version": "1.0.0",
    "description": "API for Project Management System"
  },
  "servers": [{ "url": "/api/v1" }],
  "paths": {}
}`)
}

// swaggerCSPMiddleware relaxes CSP for Swagger UI
func swaggerCSPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Del("Content-Security-Policy")
		c.Header("Content-Security-Policy", "default-src 'self' https: data: blob:; script-src 'self' 'unsafe-inline' 'unsafe-eval' https:; style-src 'self' 'unsafe-inline' https:; img-src 'self' data: https:; font-src 'self' data: https:; connect-src 'self' data: blob: https:")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Next()
	}
}

func SetupRoutes(r *gin.Engine, db *gorm.DB, startupService *services.StartupService) {
	// ===== CONTROLLERS =====
	authController := controllers.NewAuthController(db)
	userController := controllers.NewUserController(db)
	permissionController := controllers.NewPermissionController(db)
	debugController := controllers.NewDebugController()
	monitoringController := controllers.NewMonitoringController()

	// Initialize startup handler
	startupHandler := handlers.NewStartupHandler(startupService)

	// Notification repositories, services and handlers
	notificationRepo := repositories.NewNotificationRepository(db)
	notificationService := services.NewNotificationService(db, notificationRepo)
	notificationHandler := handlers.NewNotificationHandler(notificationService, nil)
	dashboardController := controllers.NewDashboardController(db, nil)

	// Approval service for purchase requests
	approvalService := services.NewApprovalService(db)

	// Purchase Request
	cbsRepo := repositories.NewCBSRepository(db)
	prRepo := repositories.NewPurchaseRequestRepository(db)
	prService := services.NewPurchaseRequestService(prRepo, cbsRepo, db, approvalService)
	prController := controllers.NewPurchaseRequestController(prService)

	// CBS (Cost Breakdown Structure)
	cbsService := services.NewCBSService(cbsRepo)
	cbsController := controllers.NewCBSController(cbsService)

	// Employee dashboard and approval handlers
	employeeDashboardService := services.NewEmployeeDashboardService(db)
	employeeApprovalHandler := handlers.NewEmployeeApprovalHandler(approvalService, employeeDashboardService)
	purchaseApprovalHandler := handlers.NewPurchaseApprovalHandler(nil, approvalService)

	// Initialize security middleware
	middleware.InitAuditLogger(db)
	middleware.InitTokenMonitor(db)

	// Activity Logger
	activityLoggerService := services.NewActivityLoggerService(db, "./logs")
	middleware.InitActivityLogger(activityLoggerService)
	log.Println("✅ Activity Logger initialized successfully")

	activityLogController := controllers.NewActivityLogController()
	securityController := controllers.NewSecurityController(db)

	// Initialize JWT Manager
	jwtManager := middleware.NewJWTManager(db)

	// Session Cleanup
	sessionCleanupService := services.NewSessionCleanupService(db)
	sessionController := controllers.NewSessionController(sessionCleanupService)

	// Debug Auth Controller
	debugAuthController := controllers.NewDebugAuthController()

	// Permission Middleware
	permMiddleware := middleware.NewPermissionMiddleware(db)

	// Enhanced Security Middleware
	enhancedSecurity := middleware.NewEnhancedSecurityMiddleware(db)

	// Apply global security middleware
	r.Use(enhancedSecurity.SecurityHeaders())
	r.Use(enhancedSecurity.RequestMonitoring())
	r.Use(middleware.APIUsageMiddleware())
	r.Use(middleware.ActivityLoggerMiddleware())

	if middleware.GlobalAuditLogger != nil {
		r.Use(middleware.GlobalAuditLogger.AuditMiddleware())
	}

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		auth.Use(middleware.AuthRateLimit())
		auth.Use(enhancedSecurity.SecurityHeaders())
		{
			auth.POST("/login", authController.Login)
			if isDevelopmentMode() || os.Getenv("ALLOW_REGISTRATION") == "true" {
				auth.POST("/register", authController.Register)
			}
			auth.POST("/refresh", authController.RefreshToken)
			auth.GET("/validate-token", jwtManager.AuthRequired(), authController.ValidateToken)
			auth.GET("/session-info", jwtManager.AuthRequired(), authController.GetSessionInfo)
		}

		// Debug auth routes (development only)
		if isDevelopmentMode() {
			debugAuth := v1.Group("/debug-auth")
			{
				debugAuth.GET("/headers", debugAuthController.DebugAuthHeader)
				debugAuth.GET("/session", jwtManager.AuthRequired(), debugAuthController.DebugSessionValidation)
			}
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(jwtManager.AuthRequired())
		{
			// Profile routes
			protected.GET("/profile", authController.Profile)

			// User management routes (admin only)
			users := protected.Group("/users")
			{
				users.GET("", middleware.RoleRequired("admin"), userController.GetUsers)
				users.GET("/:id", middleware.RoleRequired("admin"), userController.GetUser)
				users.POST("", middleware.RoleRequired("admin"), userController.CreateUser)
				users.PUT("/:id", middleware.RoleRequired("admin"), userController.UpdateUser)
				users.DELETE("/:id", middleware.RoleRequired("admin"), userController.DeleteUser)
			}

			// Permission management routes
			permissions := protected.Group("/permissions")
			{
				permissions.GET("/users", middleware.RoleRequired("admin"), permissionController.GetAllUsersPermissions)
				permissions.GET("/users/:userId", middleware.RoleRequired("admin"), permissionController.GetUserPermissions)
				permissions.PUT("/users/:userId", middleware.RoleRequired("admin"), permissionController.UpdateUserPermissions)
				permissions.POST("/users/:userId/reset", middleware.RoleRequired("admin"), permissionController.ResetToDefaultPermissions)
				permissions.GET("/me", permissionController.GetMyPermissions)
				permissions.GET("/check", permissionController.CheckUserPermission)
			}

			// Dashboard routes
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/analytics", permMiddleware.CanView("reports"), dashboardController.GetAnalytics)
				dashboard.GET("/employee", dashboardController.GetEmployeeDashboardData)
				dashboard.GET("/employee/workflows", dashboardController.GetEmployeeApprovalWorkflows)
				dashboard.GET("/employee/purchase-requests", dashboardController.GetEmployeePurchaseRequests)
				dashboard.GET("/employee/approval-notifications", dashboardController.GetEmployeeApprovalNotifications)
				dashboard.GET("/employee/purchase-approval-status", dashboardController.GetEmployeePurchaseApprovalStatus)
				dashboard.GET("/employee/notifications-summary", dashboardController.GetEmployeeNotificationsSummary)
				dashboard.PATCH("/employee/notifications/:id/read", dashboardController.MarkNotificationAsRead)
			}

			// Notification routes
			notifs := protected.Group("/notifications")
			{
				notifs.GET("", notificationHandler.GetNotifications)
				notifs.GET("/unread-count", notificationHandler.GetUnreadCount)
				notifs.PUT("/:id/read", notificationHandler.MarkNotificationAsRead)
				notifs.PUT("/read-all", notificationHandler.MarkAllNotificationsAsRead)
				notifs.GET("/type/:type", notificationHandler.GetNotificationsByType)
				notifs.GET("/approvals", notificationHandler.GetApprovalNotifications)
			}

			// Purchase Requests
			pr := protected.Group("/purchase-requests")
			{
				pr.POST("", permMiddleware.CanCreate("purchases"), prController.Create)
				pr.GET("", permMiddleware.CanView("purchases"), prController.GetAll)
				pr.GET("/:id", permMiddleware.CanView("purchases"), prController.GetByID)
				pr.PUT("/:id", permMiddleware.CanEdit("purchases"), prController.Update)
				pr.DELETE("/:id", permMiddleware.CanDelete("purchases"), prController.Delete)
				pr.PATCH("/:id/status", permMiddleware.CanApprove("purchases"), prController.UpdateStatus)
				pr.GET("/:id/material-impact", permMiddleware.CanView("purchases"), prController.GetMaterialImpact)
			}

			// Approval workflows routes
			workflows := protected.Group("/approval-workflows")
			{
				workflows.GET("", purchaseApprovalHandler.GetApprovalWorkflows)
				workflows.POST("", middleware.RoleRequired("admin"), purchaseApprovalHandler.CreateApprovalWorkflow)
			}

			// Employee approval routes
			employeeApprovals := protected.Group("/employee/approvals")
			{
				employeeApprovals.GET("/requests", employeeApprovalHandler.GetMyApprovalRequests)
				employeeApprovals.GET("/pending", employeeApprovalHandler.GetPendingApprovalsForMe)
				employeeApprovals.POST("/:id/process", employeeApprovalHandler.ProcessApproval)
				employeeApprovals.GET("/:id/history", employeeApprovalHandler.GetApprovalHistory)
				employeeApprovals.GET("/my-requests", employeeApprovalHandler.GetMySubmittedRequests)
				employeeApprovals.GET("/workflows", employeeApprovalHandler.GetApprovalWorkflowsForEmployee)
				employeeApprovals.GET("/statistics", employeeApprovalHandler.GetApprovalStatistics)
			}

			// Setup Settings routes
			SetupSettingsRoutes(protected, db)

			// Monitoring routes (admin only)
			monitoring := protected.Group("/monitoring")
			monitoring.Use(middleware.RoleRequired("admin"))
			{
				monitoring.GET("/status", monitoringController.GetSystemSecurityStatus)
				monitoring.GET("/rate-limits", monitoringController.GetRateLimitStatus)
				monitoring.GET("/security-alerts", monitoringController.GetSecurityAlerts)
				monitoring.GET("/audit-logs", monitoringController.GetAuditLogs)
				monitoring.GET("/token-stats", monitoringController.GetTokenStats)
				monitoring.GET("/refresh-events", monitoringController.GetRecentRefreshEvents)
				monitoring.GET("/sessions/stats", sessionController.GetSessionStats)
				monitoring.POST("/sessions/cleanup", sessionController.ForceCleanup)
				monitoring.GET("/sessions/active", jwtManager.AuthRequired(), sessionController.GetActiveSessions)
				monitoring.GET("/users/:user_id/security-summary", monitoringController.GetUserSecuritySummary)
				monitoring.GET("/startup-status", startupHandler.GetStartupStatus)
			}

			// API Usage monitoring
			apiUsageController := controllers.NewAPIUsageController()
			monitoring.GET("/api-usage/stats", apiUsageController.GetAPIUsageStats)
			monitoring.GET("/api-usage/top", apiUsageController.GetTopEndpoints)
			monitoring.GET("/api-usage/unused", apiUsageController.GetUnusedEndpoints)
			monitoring.GET("/api-usage/analytics", apiUsageController.GetUsageAnalytics)

			// Activity logs
			activityLogs := protected.Group("/activity-logs")
			activityLogs.Use(middleware.RoleRequired("admin"))
			{
				activityLogs.GET("", activityLogController.GetActivityLogs)
				activityLogs.GET("/user/:userId", activityLogController.GetUserActivityLogs)
				activityLogs.GET("/stats", activityLogController.GetActivityStats)
			}

			// Security routes
			security := protected.Group("/security")
			security.Use(middleware.RoleRequired("admin"))
			{
				security.GET("/dashboard", securityController.GetSecurityDashboard)
				security.GET("/threats", securityController.GetThreatAnalysis)
			}

			// Cost Control routes
			costControl := protected.Group("/cost-control")
			{
				// CBS routes
				cbs := costControl.Group("/cbs")
				{
					cbs.GET("", permMiddleware.CanView("projects"), cbsController.GetAll)
					cbs.GET("/:id", permMiddleware.CanView("projects"), cbsController.GetByID)
					cbs.POST("", permMiddleware.CanCreate("projects"), cbsController.Create)
					cbs.PUT("/:id", permMiddleware.CanEdit("projects"), cbsController.Update)
					cbs.DELETE("/:id", permMiddleware.CanDelete("projects"), cbsController.Delete)
					cbs.GET("/project/:projectId", permMiddleware.CanView("projects"), cbsController.GetByProject)
					cbs.GET("/project/:projectId/tree", permMiddleware.CanView("projects"), cbsController.GetTreeByProject)
				}

				// Material Tracking routes
				materialTrackingService := services.NewMaterialTrackingService(db)
				materialTrackingController := controllers.NewMaterialTrackingController(materialTrackingService)
				materialGroup := costControl.Group("/material-tracking")
				{
					materialGroup.GET("/:projectId/summary", materialTrackingController.GetSummary)
					materialGroup.GET("/:projectId/items", materialTrackingController.GetItems)
					materialGroup.GET("/:projectId/movements", materialTrackingController.GetMovements)
					materialGroup.POST("/:projectId/record-usage", materialTrackingController.RecordUsage)
				}

				// Project Reports routes
				SetupProjectReportRoutes(protected, db, jwtManager)
			}

			// Project routes
			SetupProjectRoutes(protected, db, permMiddleware)
		}
	}

	r.Static("/templates", "./templates")

	// Global favicon handler
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	// Health check endpoint
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger documentation
	if isDevelopmentMode() || os.Getenv("ENABLE_SWAGGER") == "true" {
		swaggerBytes, pathUsed, err := loadSwaggerJSON()
		if err != nil {
			log.Printf("⚠️  Swagger spec not found: %v", err)
			swaggerBytes = minimalSwaggerJSON()
			pathUsed = "embedded-minimal"
		}
		log.Printf("📄 Serving Swagger spec from: %s", pathUsed)
		r.GET("/openapi/doc.json", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/json", swaggerBytes)
		})
		sg := r.Group("/")
		sg.Use(swaggerCSPMiddleware())
		sg.GET("/swagger", func(c *gin.Context) { c.Redirect(http.StatusFound, "/swagger/index.html") })
		sg.GET("/swagger/index.html", func(c *gin.Context) {
			html := config.GetEnhancedSwaggerHTML("/openapi/enhanced-doc.json")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		})
		sg.GET("/docs", func(c *gin.Context) { c.Redirect(http.StatusFound, "/docs/index.html") })
		sg.GET("/docs/index.html", func(c *gin.Context) {
			html := config.GetEnhancedSwaggerHTML("/openapi/enhanced-doc.json")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		})
	}

	// Debug routes for development only
	if gin.Mode() == gin.DebugMode {
		debug := v1.Group("/debug")
		{
			debugWithAuth := debug.Group("/auth")
			debugWithAuth.Use(jwtManager.AuthRequired())
			{
				debugWithAuth.GET("/context", debugController.TestJWTContext)
				debugWithAuth.GET("/role", debugController.TestRolePermission)
			}
		}
	}
}
