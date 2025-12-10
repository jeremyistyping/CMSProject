package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupMasterDataRoutes sets up all master data routes
func SetupMasterDataRoutes(router *gin.RouterGroup, db *gorm.DB, permMiddleware *middleware.PermissionMiddleware) {
	// Initialize repositories
	coaRepo := repositories.NewCOARepository(db)
	materialRepo := repositories.NewMaterialRepository(db)
	vendorRepo := repositories.NewVendorRepository(db)

	// Initialize services
	coaService := services.NewCOAService(coaRepo)
	materialService := services.NewMaterialService(materialRepo)
	vendorService := services.NewVendorService(vendorRepo)

	// Initialize controllers
	coaController := controllers.NewCOAController(coaService)
	materialController := controllers.NewMaterialController(materialService)
	vendorController := controllers.NewVendorController(vendorService)

	// Master Data routes group
	masterData := router.Group("/master-data")
	{
		// COA Routes
		coa := masterData.Group("/coa")
		{
			coa.GET("", permMiddleware.CanView("master_data"), coaController.GetAll)
			coa.GET("/tree", permMiddleware.CanView("master_data"), coaController.GetTree)
			coa.GET("/type/:type", permMiddleware.CanView("master_data"), coaController.GetByType)
			coa.GET("/category/:category", permMiddleware.CanView("master_data"), coaController.GetByCategory)
			coa.GET("/budget-category/:budgetCategory", permMiddleware.CanView("master_data"), coaController.GetByBudgetCategory)
			coa.GET("/:id", permMiddleware.CanView("master_data"), coaController.GetByID)
			coa.POST("", permMiddleware.CanCreate("master_data"), coaController.Create)
			coa.PUT("/:id", permMiddleware.CanEdit("master_data"), coaController.Update)
			coa.DELETE("/:id", permMiddleware.CanDelete("master_data"), coaController.Delete)
		}

		// Material Routes
		materials := masterData.Group("/materials")
		{
			materials.GET("", permMiddleware.CanView("master_data"), materialController.GetAll)
			materials.GET("/summary", permMiddleware.CanView("master_data"), materialController.GetSummary)
			materials.GET("/:id", permMiddleware.CanView("master_data"), materialController.GetByID)
			materials.GET("/:id/vendors", permMiddleware.CanView("master_data"), vendorController.GetMaterialVendors)
			materials.POST("", permMiddleware.CanCreate("master_data"), materialController.Create)
			materials.PUT("/:id", permMiddleware.CanEdit("master_data"), materialController.Update)
			materials.DELETE("/:id", permMiddleware.CanDelete("master_data"), materialController.Delete)
		}

		// Material Category Routes
		materialCategories := masterData.Group("/material-categories")
		{
			materialCategories.GET("", permMiddleware.CanView("master_data"), materialController.GetAllCategories)
			materialCategories.GET("/tree", permMiddleware.CanView("master_data"), materialController.GetCategoryTree)
			materialCategories.POST("", permMiddleware.CanCreate("master_data"), materialController.CreateCategory)
			materialCategories.PUT("/:id", permMiddleware.CanEdit("master_data"), materialController.UpdateCategory)
			materialCategories.DELETE("/:id", permMiddleware.CanDelete("master_data"), materialController.DeleteCategory)
		}

		// Unit of Measure Routes
		uom := masterData.Group("/uom")
		{
			uom.GET("", permMiddleware.CanView("master_data"), materialController.GetAllUoM)
		}

		// Vendor Routes
		vendors := masterData.Group("/vendors")
		{
			vendors.GET("", permMiddleware.CanView("master_data"), vendorController.GetAll)
			vendors.GET("/summary", permMiddleware.CanView("master_data"), vendorController.GetSummary)
			vendors.GET("/:id", permMiddleware.CanView("master_data"), vendorController.GetByID)
			vendors.GET("/:id/materials", permMiddleware.CanView("master_data"), vendorController.GetVendorMaterials)
			vendors.POST("", permMiddleware.CanCreate("master_data"), vendorController.Create)
			vendors.PUT("/:id", permMiddleware.CanEdit("master_data"), vendorController.Update)
			vendors.DELETE("/:id", permMiddleware.CanDelete("master_data"), vendorController.Delete)
			vendors.POST("/:id/materials", permMiddleware.CanCreate("master_data"), vendorController.AddVendorMaterial)
			vendors.DELETE("/:id/materials/:materialId", permMiddleware.CanDelete("master_data"), vendorController.RemoveVendorMaterial)
		}

		// Vendor Category Routes
		vendorCategories := masterData.Group("/vendor-categories")
		{
			vendorCategories.GET("", permMiddleware.CanView("master_data"), vendorController.GetAllCategories)
			vendorCategories.POST("", permMiddleware.CanCreate("master_data"), vendorController.CreateCategory)
			vendorCategories.PUT("/:id", permMiddleware.CanEdit("master_data"), vendorController.UpdateCategory)
			vendorCategories.DELETE("/:id", permMiddleware.CanDelete("master_data"), vendorController.DeleteCategory)
		}
	}
}

// SeedMasterData seeds default master data
func SeedMasterData(db *gorm.DB) error {
	coaRepo := repositories.NewCOARepository(db)
	materialRepo := repositories.NewMaterialRepository(db)
	vendorRepo := repositories.NewVendorRepository(db)

	coaService := services.NewCOAService(coaRepo)
	materialService := services.NewMaterialService(materialRepo)
	vendorService := services.NewVendorService(vendorRepo)

	// Seed COA
	if err := coaService.SeedDefaultCOA(); err != nil {
		return err
	}

	// Seed Material Categories and UoM
	if err := materialService.SeedDefaultCategories(); err != nil {
		return err
	}
	if err := materialService.SeedDefaultUoM(); err != nil {
		return err
	}

	// Seed Vendor Categories
	if err := vendorService.SeedDefaultCategories(); err != nil {
		return err
	}

	return nil
}
