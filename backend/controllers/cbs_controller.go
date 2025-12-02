package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBSController struct {
	service services.CBSService
}

func NewCBSController(service services.CBSService) *CBSController {
	return &CBSController{service: service}
}

// GetProjectCBSTree retrieves the CBS tree for a project
// GET /api/v1/projects/:id/cbs
func (c *CBSController) GetProjectCBSTree(ctx *gin.Context) {
	projectID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	nodes, err := c.service.GetProjectCBSTree(uint(projectID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, nodes)
}

// CreateCBSNode creates a new CBS node
// POST /api/v1/cbs-nodes
func (c *CBSController) CreateCBSNode(ctx *gin.Context) {
	var node models.CBSNode
	if err := ctx.ShouldBindJSON(&node); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.CreateCBSNode(&node); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, node)
}

// UpdateCBSNode updates an existing CBS node
// PUT /api/v1/cbs-nodes/:id
func (c *CBSController) UpdateCBSNode(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	var node models.CBSNode
	if err := ctx.ShouldBindJSON(&node); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.UpdateCBSNode(uint(id), &node); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "CBS node updated successfully"})
}

// DeleteCBSNode deletes a CBS node
// DELETE /api/v1/cbs-nodes/:id
func (c *CBSController) DeleteCBSNode(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	if err := c.service.DeleteCBSNode(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "CBS node deleted successfully"})
}

// GetNodeSummary retrieves cost summary for a node
// GET /api/v1/cbs-nodes/:id/summary
func (c *CBSController) GetNodeSummary(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	summary, err := c.service.GetNodeCostSummary(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}
