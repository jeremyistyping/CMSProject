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

// GetPRCBSMappings retrieves CBS mappings for a purchase request
// GET /api/v1/purchase-requests/:id/cbs-mappings
func (c *CBSController) GetPRCBSMappings(ctx *gin.Context) {
	prID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase request ID"})
		return
	}

	mappings, err := c.service.GetPRCBSMappings(uint(prID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, mappings)
}

type VerifyPRRequest struct {
	Mappings []models.PRCBSMapping `json:"mappings"`
	Notes    string                `json:"notes"`
}

// VerifyPurchaseRequest verifies a PR and saves CBS mappings
// POST /api/v1/purchase-requests/:id/verify
func (c *CBSController) VerifyPurchaseRequest(ctx *gin.Context) {
	prID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase request ID"})
		return
	}

	var req VerifyPRRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := c.service.VerifyPurchaseRequest(uint(prID), uint(userID.(float64)), req.Mappings, req.Notes); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Purchase request verified successfully"})
}
