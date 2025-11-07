package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ProductUnitController struct {
	DB *gorm.DB
}

func NewProductUnitController(db *gorm.DB) *ProductUnitController {
	return &ProductUnitController{DB: db}
}

// GetProductUnits retrieves all product units
func (puc *ProductUnitController) GetProductUnits(c *gin.Context) {
	var units []models.ProductUnit
	
	query := puc.DB.Where("is_active = ?", true)
	
	// Add search functionality
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ? OR symbol ILIKE ?", 
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	
	// Filter by type
	if unitType := c.Query("type"); unitType != "" {
		query = query.Where("type = ?", unitType)
	}
	
	if err := query.Order("name ASC").Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product units"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product units retrieved successfully",
		"data":    units,
	})
}

// GetProductUnit retrieves a single product unit by ID
func (puc *ProductUnitController) GetProductUnit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var unit models.ProductUnit
	if err := puc.DB.First(&unit, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product unit not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product unit"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product unit retrieved successfully",
		"data":    unit,
	})
}

// CreateProductUnit creates a new product unit
func (puc *ProductUnitController) CreateProductUnit(c *gin.Context) {
	var unit models.ProductUnit
	if err := c.ShouldBindJSON(&unit); err != nil {
		fmt.Printf("JSON binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Basic validation
	if unit.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is required"})
		return
	}
	
	if unit.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	// Check if unit code already exists (including soft deleted records)
	var existingUnit models.ProductUnit
	findResult := puc.DB.Unscoped().Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("code = ?", unit.Code).First(&existingUnit)
	
	// Handle the three possible outcomes:
	switch {
	case findResult.Error == nil:
		// Record found - check if it's soft deleted or active
		if existingUnit.DeletedAt.Valid {
			// Found a soft deleted record - restore it with new data
			existingUnit.Name = unit.Name
			existingUnit.Symbol = unit.Symbol
			existingUnit.Type = unit.Type
			existingUnit.Description = unit.Description
			existingUnit.IsActive = unit.IsActive
			existingUnit.DeletedAt = gorm.DeletedAt{} // Clear the deleted_at
			
			if err := puc.DB.Unscoped().Save(&existingUnit).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore product unit"})
				return
			}
			
			c.JSON(http.StatusCreated, gin.H{
				"message": "Product unit created successfully (restored from deleted)",
				"data":    existingUnit,
			})
			return
		} else {
			// Active record with same code exists
			c.JSON(http.StatusConflict, gin.H{
				"error": fmt.Sprintf("Unit code '%s' already exists", unit.Code),
				"existing_unit": map[string]interface{}{
					"id": existingUnit.ID,
					"name": existingUnit.Name,
					"code": existingUnit.Code,
				},
			})
			return
		}
		
	case findResult.Error == gorm.ErrRecordNotFound:
		// No existing record found - proceed to create new unit
		// This is the normal case for new units
		break
		
	default:
		// Some other database error occurred
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check existing product unit",
			"details": findResult.Error.Error(),
		})
		return
	}

	if err := puc.DB.Create(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product unit"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product unit created successfully",
		"data":    unit,
	})
}

// UpdateProductUnit updates an existing product unit
func (puc *ProductUnitController) UpdateProductUnit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var unit models.ProductUnit
	if err := puc.DB.First(&unit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product unit not found"})
		return
	}

	var updateData models.ProductUnit
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if new code conflicts with existing units
	if updateData.Code != unit.Code {
		var existingUnit models.ProductUnit
		if err := puc.DB.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("code = ? AND id != ?", updateData.Code, id).First(&existingUnit).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Unit code already exists"})
			return
		}
	}

	if err := puc.DB.Model(&unit).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product unit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product unit updated successfully",
		"data":    unit,
	})
}

// DeleteProductUnit soft deletes a product unit
func (puc *ProductUnitController) DeleteProductUnit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var unit models.ProductUnit
	if err := puc.DB.First(&unit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product unit not found"})
		return
	}

	// Check if unit is being used by products
	var productCount int64
	if err := puc.DB.Model(&models.Product{}).Where("unit = ?", unit.Code).Count(&productCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check unit usage"})
		return
	}

	if productCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete unit that is being used by products",
			"product_count": productCount,
		})
		return
	}

	// Soft delete
	if err := puc.DB.Delete(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product unit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product unit deleted successfully",
	})
}
