package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type COAController struct {
	service services.COAService
}

func NewCOAController(service services.COAService) *COAController {
	return &COAController{service: service}
}

// GetAll retrieves all COA accounts
// GET /api/v1/master-data/coa
func (c *COAController) GetAll(ctx *gin.Context) {
	filter := make(map[string]interface{})

	if coaType := ctx.Query("type"); coaType != "" {
		filter["type"] = coaType
	}
	if category := ctx.Query("category"); category != "" {
		filter["category"] = category
	}
	if budgetCategory := ctx.Query("budget_category"); budgetCategory != "" {
		filter["budget_category"] = budgetCategory
	}
	if workPackage := ctx.Query("work_package"); workPackage != "" {
		filter["work_package"] = workPackage
	}
	if isActive := ctx.Query("is_active"); isActive != "" {
		filter["is_active"] = isActive == "true"
	}
	if search := ctx.Query("search"); search != "" {
		filter["search"] = search
	}

	accounts, err := c.service.GetAll(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": accounts, "total": len(accounts)})
}


// GetByID retrieves a COA account by ID
// GET /api/v1/master-data/coa/:id
func (c *COAController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid COA ID"})
		return
	}

	account, err := c.service.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "COA account not found"})
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// GetTree retrieves COA tree structure
// GET /api/v1/master-data/coa/tree
func (c *COAController) GetTree(ctx *gin.Context) {
	tree, err := c.service.GetTree()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": tree})
}

// GetByType retrieves COA accounts by type
// GET /api/v1/master-data/coa/type/:type
func (c *COAController) GetByType(ctx *gin.Context) {
	coaType := ctx.Param("type")
	accounts, err := c.service.GetByType(coaType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": accounts, "total": len(accounts)})
}

// GetByCategory retrieves COA accounts by category
// GET /api/v1/master-data/coa/category/:category
func (c *COAController) GetByCategory(ctx *gin.Context) {
	category := ctx.Param("category")
	accounts, err := c.service.GetByCategory(category)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": accounts, "total": len(accounts)})
}

// GetByBudgetCategory retrieves COA accounts by budget category
// GET /api/v1/master-data/coa/budget-category/:budgetCategory
func (c *COAController) GetByBudgetCategory(ctx *gin.Context) {
	budgetCategory := ctx.Param("budgetCategory")
	accounts, err := c.service.GetByBudgetCategory(budgetCategory)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": accounts, "total": len(accounts)})
}

// Create creates a new COA account
// POST /api/v1/master-data/coa
func (c *COAController) Create(ctx *gin.Context) {
	var dto models.CreateCOADTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := c.service.Create(&dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, account)
}

// Update updates an existing COA account
// PUT /api/v1/master-data/coa/:id
func (c *COAController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid COA ID"})
		return
	}

	var dto models.UpdateCOADTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := c.service.Update(uint(id), &dto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// Delete deletes a COA account
// DELETE /api/v1/master-data/coa/:id
func (c *COAController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid COA ID"})
		return
	}

	if err := c.service.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "COA account deleted successfully"})
}
