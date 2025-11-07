package controllers

import (
	"net/http"
	"strconv"

	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type WarehouseLocationController struct {
	DB *gorm.DB
}

func NewWarehouseLocationController(db *gorm.DB) *WarehouseLocationController {
	return &WarehouseLocationController{
		DB: db,
	}
}

func (wc *WarehouseLocationController) GetWarehouseLocations(c *gin.Context) {
	var locations []models.WarehouseLocation
	
	query := wc.DB.Where("is_active = ?", true)
	
	// Add search functionality
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	if err := query.Find(&locations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch warehouse locations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warehouse locations retrieved successfully",
		"data":    locations,
	})
}

func (wc *WarehouseLocationController) GetWarehouseLocation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warehouse location ID"})
		return
	}

	var location models.WarehouseLocation
	if err := wc.DB.First(&location, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse location not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warehouse location retrieved successfully",
		"data":    location,
	})
}

func (wc *WarehouseLocationController) CreateWarehouseLocation(c *gin.Context) {
	var location models.WarehouseLocation
	if err := c.ShouldBindJSON(&location); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if location code already exists
	var existingLocation models.WarehouseLocation
	if err := wc.DB.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("code = ?", location.Code).First(&existingLocation).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Warehouse location code already exists"})
		return
	}

	if err := wc.DB.Create(&location).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create warehouse location"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Warehouse location created successfully",
		"data":    location,
	})
}

func (wc *WarehouseLocationController) UpdateWarehouseLocation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warehouse location ID"})
		return
	}

	var location models.WarehouseLocation
	if err := wc.DB.First(&location, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse location not found"})
		return
	}

	var updateData models.WarehouseLocation
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if new code conflicts with existing locations
	if updateData.Code != location.Code {
		var existingLocation models.WarehouseLocation
		if err := wc.DB.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("code = ? AND id != ?", updateData.Code, id).First(&existingLocation).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Warehouse location code already exists"})
			return
		}
	}

	if err := wc.DB.Model(&location).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update warehouse location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warehouse location updated successfully",
		"data":    location,
	})
}

func (wc *WarehouseLocationController) DeleteWarehouseLocation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warehouse location ID"})
		return
	}

	var location models.WarehouseLocation
	if err := wc.DB.First(&location, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse location not found"})
		return
	}

	// Check if location is being used by any products
	var productCount int64
	if err := wc.DB.Model(&models.Product{}).Where("warehouse_location_id = ?", id).Count(&productCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check location usage"})
		return
	}

	if productCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete warehouse location as it is being used by products"})
		return
	}

	// Soft delete
	if err := wc.DB.Delete(&location).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete warehouse location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warehouse location deleted successfully",
	})
}
