package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MaterialController struct {
	service services.MaterialService
}

func NewMaterialController(service services.MaterialService) *MaterialController {
	return &MaterialController{service: service}
}

// GetAll retrieves all materials
// GET /api/v1/master-data/materials
func (c *MaterialController) GetAll(ctx *gin.Context) {
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
	if lowStock := ctx.Query("low_stock"); lowStock == "true" {
		filter["low_stock"] = true
	}

	materials, err := c.service.GetAll(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": materials, "total": len(materials)})
}


// GetByID retrieves a material by ID
// GET /api/v1/master-data/materials/:id
func (c *MaterialController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	material, err := c.service.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
		return
	}

	ctx.JSON(http.StatusOK, material)
}

// Create creates a new material
// POST /api/v1/master-data/materials
func (c *MaterialController) Create(ctx *gin.Context) {
	var dto models.CreateMaterialDTO
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

	material, err := c.service.Create(&dto, uid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, material)
}

// Update updates an existing material
// PUT /api/v1/master-data/materials/:id
func (c *MaterialController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	var dto models.UpdateMaterialDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	material, err := c.service.Update(uint(id), &dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, material)
}

// Delete deletes a material
// DELETE /api/v1/master-data/materials/:id
func (c *MaterialController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	if err := c.service.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Material deleted successfully"})
}

// GetSummary retrieves material summary statistics
// GET /api/v1/master-data/materials/summary
func (c *MaterialController) GetSummary(ctx *gin.Context) {
	summary, err := c.service.GetSummary()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

// GetAllCategories retrieves all material categories
// GET /api/v1/master-data/material-categories
func (c *MaterialController) GetAllCategories(ctx *gin.Context) {
	categories, err := c.service.GetAllCategories()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": categories, "total": len(categories)})
}

// GetCategoryTree retrieves material category tree
// GET /api/v1/master-data/material-categories/tree
func (c *MaterialController) GetCategoryTree(ctx *gin.Context) {
	tree, err := c.service.GetCategoryTree()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": tree})
}

// CreateCategory creates a new material category
// POST /api/v1/master-data/material-categories
func (c *MaterialController) CreateCategory(ctx *gin.Context) {
	var dto models.CreateMaterialCategoryDTO
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

// UpdateCategory updates an existing material category
// PUT /api/v1/master-data/material-categories/:id
func (c *MaterialController) UpdateCategory(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var dto models.CreateMaterialCategoryDTO
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

// DeleteCategory deletes a material category
// DELETE /api/v1/master-data/material-categories/:id
func (c *MaterialController) DeleteCategory(ctx *gin.Context) {
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

// GetAllUoM retrieves all units of measure
// GET /api/v1/master-data/uom
func (c *MaterialController) GetAllUoM(ctx *gin.Context) {
	uoms, err := c.service.GetAllUoM()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": uoms, "total": len(uoms)})
}
