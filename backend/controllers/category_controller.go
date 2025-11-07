package controllers

import (
	"log"
	"net/http"
	"strconv"
	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type CategoryController struct {
	DB *gorm.DB
}

func NewCategoryController(db *gorm.DB) *CategoryController {
	return &CategoryController{DB: db}
}

// GetCategories retrieves all product categories with hierarchical structure
func (cc *CategoryController) GetCategories(c *gin.Context) {
	var categories []models.ProductCategory
	
	query := cc.DB.Where("is_active = ?", true)
	
	// Include parent/children relationships
	if c.Query("include_relations") == "true" {
		query = query.Preload("Parent").Preload("Children")
	}
	
	// Filter by parent ID to get categories at specific level
	if parentID := c.Query("parent_id"); parentID != "" {
		if parentID == "0" || parentID == "null" {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", parentID)
		}
	}

	if err := query.Order("name ASC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Categories retrieved successfully",
		"data":    categories,
	})
}

// GetCategory retrieves a single product category by ID
func (cc *CategoryController) GetCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var category models.ProductCategory
	query := cc.DB.Preload("Parent").Preload("Children").Preload("Products")
	
	if err := query.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category retrieved successfully",
		"data":    category,
	})
}

// CreateCategory creates a new product category
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	log.Printf("[DEBUG] CreateCategory called with method: %s", c.Request.Method)
	
	var category models.ProductCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		log.Printf("[ERROR] Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("[DEBUG] Parsed category: Code=%s, Name=%s, ParentID=%v", category.Code, category.Name, category.ParentID)

	// Check if category code already exists (including soft deleted)
	var existingCategory models.ProductCategory
	if err := cc.DB.Unscoped().Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("code = ?", category.Code).First(&existingCategory).Error; err == nil {
		// If found and soft deleted, undelete and update it
		if existingCategory.DeletedAt.Valid {
			log.Printf("[DEBUG] Found soft-deleted category with code %s, restoring it...", category.Code)
			existingCategory.DeletedAt = gorm.DeletedAt{}
			existingCategory.Name = category.Name
			existingCategory.Description = category.Description
			existingCategory.ParentID = category.ParentID
			existingCategory.IsActive = true
			existingCategory.DefaultExpenseAccountID = category.DefaultExpenseAccountID
			
			if err := cc.DB.Unscoped().Save(&existingCategory).Error; err != nil {
				log.Printf("[ERROR] Failed to restore category: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore category"})
				return
			}
			
			log.Printf("[SUCCESS] Category restored successfully with ID: %d", existingCategory.ID)
			c.JSON(http.StatusCreated, gin.H{
				"message": "Category restored successfully",
				"data":    existingCategory,
			})
			return
		}
		
		// If found and not deleted, it's a conflict
		c.JSON(http.StatusConflict, gin.H{"error": "Category code already exists"})
		return
	}

	// Validate parent category exists if parent_id is provided
	if category.ParentID != nil {
		var parentCategory models.ProductCategory
		if err := cc.DB.First(&parentCategory, *category.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent category not found"})
			return
		}
	}

	log.Printf("[DEBUG] Attempting to save category to database...")
	if err := cc.DB.Create(&category).Error; err != nil {
		log.Printf("[ERROR] Database error creating category: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}
	
	log.Printf("[SUCCESS] Category created successfully with ID: %d", category.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Category created successfully",
		"data":    category,
	})
}

// UpdateCategory updates an existing product category
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var category models.ProductCategory
	if err := cc.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var updateData models.ProductCategory
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if new code conflicts with existing categories
	if updateData.Code != category.Code {
		var existingCategory models.ProductCategory
		// Check both active and soft deleted records
		if err := cc.DB.Unscoped().Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("code = ? AND id != ? AND deleted_at IS NULL", updateData.Code, id).First(&existingCategory).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Category code already exists"})
			return
		}
	}

	// Validate parent category exists if parent_id is provided
	if updateData.ParentID != nil {
		// Prevent self-referencing
		if *updateData.ParentID == uint(id) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category cannot be its own parent"})
			return
		}
		
		var parentCategory models.ProductCategory
		if err := cc.DB.First(&parentCategory, *updateData.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent category not found"})
			return
		}
	}

	if err := cc.DB.Model(&category).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category updated successfully",
		"data":    category,
	})
}

// DeleteCategory soft deletes a product category
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var category models.ProductCategory
	if err := cc.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Check if category has products
	var productCount int64
	if err := cc.DB.Model(&models.Product{}).Where("category_id = ?", id).Count(&productCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check category usage"})
		return
	}

	if productCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete category with associated products",
			"product_count": productCount,
		})
		return
	}

	// Check if category has child categories
	var childCount int64
	if err := cc.DB.Model(&models.ProductCategory{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check child categories"})
		return
	}

	if childCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete category with child categories",
			"child_count": childCount,
		})
		return
	}

	// Soft delete
	if err := cc.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}

// GetCategoryTree retrieves the complete category hierarchy tree
func (cc *CategoryController) GetCategoryTree(c *gin.Context) {
	var rootCategories []models.ProductCategory
	
	// Get root categories (those without parent)
	if err := cc.DB.Where("parent_id IS NULL AND is_active = ?", true).
		Preload("Children", "is_active = ?", true).
		Order("name ASC").Find(&rootCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category tree"})
		return
	}

	// Recursively load children
	for i := range rootCategories {
		cc.loadCategoryChildren(&rootCategories[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category tree retrieved successfully",
		"data":    rootCategories,
	})
}

// GetCategoryProducts retrieves all products in a specific category (including subcategories)
func (cc *CategoryController) GetCategoryProducts(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	// Get category
	var category models.ProductCategory
	if err := cc.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Get all subcategory IDs
	categoryIDs := []uint{uint(id)}
	cc.collectSubcategoryIDs(uint(id), &categoryIDs)

	// Get products from this category and all subcategories
	var products []models.Product
	query := cc.DB.Where("category_id IN ? AND is_active = ?", categoryIDs, true)
	
	// Add search functionality
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Preload("Category").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category products retrieved successfully",
		"category": category,
		"products": products,
		"count":    len(products),
	})
}

// Helper methods

// loadCategoryChildren recursively loads all children for a category
func (cc *CategoryController) loadCategoryChildren(category *models.ProductCategory) {
	var children []models.ProductCategory
	if err := cc.DB.Where("parent_id = ? AND is_active = ?", category.ID, true).
		Order("name ASC").Find(&children).Error; err != nil {
		return
	}

	category.Children = children
	for i := range category.Children {
		cc.loadCategoryChildren(&category.Children[i])
	}
}

// collectSubcategoryIDs recursively collects all subcategory IDs
func (cc *CategoryController) collectSubcategoryIDs(parentID uint, categoryIDs *[]uint) {
	var children []models.ProductCategory
	if err := cc.DB.Where("parent_id = ? AND is_active = ?", parentID, true).Find(&children).Error; err != nil {
		return
	}

	for _, child := range children {
		*categoryIDs = append(*categoryIDs, child.ID)
		cc.collectSubcategoryIDs(child.ID, categoryIDs)
	}
}
