package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VendorController struct {
	service services.VendorService
}

func NewVendorController(service services.VendorService) *VendorController {
	return &VendorController{service: service}
}

// GetAll retrieves all vendors
// GET /api/v1/master-data/vendors
func (c *VendorController) GetAll(ctx *gin.Context) {
	filter := make(map[string]interface{})

	if categoryID := ctx.Query("category_id"); categoryID != "" {
		if id, err := strconv.ParseUint(categoryID, 10, 32); err == nil {
			filter["category_id"] = uint(id)
		}
	}
	if isActive := ctx.Query("is_active"); isActive != "" {
		filter["is_active"] = isActive == "true"
	}
	if search := ctx.Query("search"); search != "" {
		filter["search"] = search
	}
	if city := ctx.Query("city"); city != "" {
		filter["city"] = city
	}

	vendors, err := c.service.GetAll(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": vendors, "total": len(vendors)})
}


// GetByID retrieves a vendor by ID
// GET /api/v1/master-data/vendors/:id
func (c *VendorController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vendor ID"})
		return
	}

	vendor, err := c.service.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
		return
	}

	ctx.JSON(http.StatusOK, vendor)
}

// Create creates a new vendor
// POST /api/v1/master-data/vendors
func (c *VendorController) Create(ctx *gin.Context) {
	var dto models.CreateVendorDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Handle different types for userID
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case float64:
		uid = uint(v)
	case int:
		uid = uint(v)
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	vendor, err := c.service.Create(&dto, uid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, vendor)
}

// Update updates an existing vendor
// PUT /api/v1/master-data/vendors/:id
func (c *VendorController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vendor ID"})
		return
	}

	var dto models.UpdateVendorDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vendor, err := c.service.Update(uint(id), &dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, vendor)
}

// Delete deletes a vendor
// DELETE /api/v1/master-data/vendors/:id
func (c *VendorController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vendor ID"})
		return
	}

	if err := c.service.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Vendor deleted successfully"})
}

// GetSummary retrieves vendor summary statistics
// GET /api/v1/master-data/vendors/summary
func (c *VendorController) GetSummary(ctx *gin.Context) {
	summary, err := c.service.GetSummary()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

// GetAllCategories retrieves all vendor categories
// GET /api/v1/master-data/vendor-categories
func (c *VendorController) GetAllCategories(ctx *gin.Context) {
	categories, err := c.service.GetAllCategories()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": categories, "total": len(categories)})
}

// CreateCategory creates a new vendor category
// POST /api/v1/master-data/vendor-categories
func (c *VendorController) CreateCategory(ctx *gin.Context) {
	var dto models.CreateVendorCategoryDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := c.service.CreateCategory(&dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, category)
}

// UpdateCategory updates an existing vendor category
// PUT /api/v1/master-data/vendor-categories/:id
func (c *VendorController) UpdateCategory(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var dto models.CreateVendorCategoryDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := c.service.UpdateCategory(uint(id), &dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, category)
}

// DeleteCategory deletes a vendor category
// DELETE /api/v1/master-data/vendor-categories/:id
func (c *VendorController) DeleteCategory(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	if err := c.service.DeleteCategory(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// GetVendorMaterials retrieves materials supplied by a vendor
// GET /api/v1/master-data/vendors/:id/materials
func (c *VendorController) GetVendorMaterials(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vendor ID"})
		return
	}

	materials, err := c.service.GetVendorMaterials(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": materials, "total": len(materials)})
}

// GetMaterialVendors retrieves vendors that supply a material
// GET /api/v1/master-data/materials/:id/vendors
func (c *VendorController) GetMaterialVendors(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	vendors, err := c.service.GetMaterialVendors(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": vendors, "total": len(vendors)})
}

type AddVendorMaterialRequest struct {
	MaterialID  uint    `json:"material_id" binding:"required"`
	UnitPrice   float64 `json:"unit_price"`
	LeadTime    int     `json:"lead_time_days"`
	IsPreferred bool    `json:"is_preferred"`
}

// AddVendorMaterial adds a material to a vendor's supply list
// POST /api/v1/master-data/vendors/:id/materials
func (c *VendorController) AddVendorMaterial(ctx *gin.Context) {
	vendorID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vendor ID"})
		return
	}

	var req AddVendorMaterialRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.AddVendorMaterial(uint(vendorID), req.MaterialID, req.UnitPrice, req.LeadTime, req.IsPreferred); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Material added to vendor successfully"})
}

// RemoveVendorMaterial removes a material from a vendor's supply list
// DELETE /api/v1/master-data/vendors/:id/materials/:materialId
func (c *VendorController) RemoveVendorMaterial(ctx *gin.Context) {
	vendorID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vendor ID"})
		return
	}

	materialID, err := strconv.ParseUint(ctx.Param("materialId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	if err := c.service.RemoveVendorMaterial(uint(vendorID), uint(materialID)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Material removed from vendor successfully"})
}
