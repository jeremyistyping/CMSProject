package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProjectBudgetController handles HTTP endpoints for project-level budgets
// (backed by project_budgets table)
type ProjectBudgetController struct {
	service services.ProjectBudgetService
}

// NewProjectBudgetController creates a new controller instance
func NewProjectBudgetController(service services.ProjectBudgetService) *ProjectBudgetController {
	return &ProjectBudgetController{service: service}
}

// GetProjectBudgets godoc
// @Summary Get budgets for a project
// @Description Get list of COA budgets for a specific project
// @Tags project-budgets
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {array} models.ProjectBudget
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{id}/budgets [get]
func (pc *ProjectBudgetController) GetProjectBudgets(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectIDUint, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	budgets, err := pc.service.GetBudgetsByProject(uint(projectIDUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, budgets)
}

// UpsertProjectBudgets godoc
// @Summary Create or update budgets for a project
// @Description Upsert budget rows per (project_id, account_id)
// @Tags project-budgets
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param body body []models.ProjectBudgetUpsertRequest true "Budget items"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{id}/budgets [post]
func (pc *ProjectBudgetController) UpsertProjectBudgets(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectIDUint, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var items []models.ProjectBudgetUpsertRequest
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := pc.service.UpsertBudgets(uint(projectIDUint), items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project budgets updated successfully"})
}

// DeleteProjectBudget godoc
// @Summary Delete a single project budget row
// @Description Soft delete a budget row for a project
// @Tags project-budgets
// @Produce json
// @Param id path int true "Project ID"
// @Param budgetId path int true "Budget ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{id}/budgets/{budgetId} [delete]
func (pc *ProjectBudgetController) DeleteProjectBudget(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectIDUint, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	budgetIDStr := c.Param("budgetId")
	budgetIDUint, err := strconv.ParseUint(budgetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid budget ID"})
		return
	}

	if err := pc.service.DeleteBudget(uint(projectIDUint), uint(budgetIDUint)); err != nil {
		// Service already differentiates not found vs other errors via error message
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project budget deleted successfully"})
}

