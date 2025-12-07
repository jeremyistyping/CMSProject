package handlers

import (
	"net/http"

	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
)

type PurchaseApprovalHandler struct {
	approvalService *services.ApprovalService
}

func NewPurchaseApprovalHandler(_ interface{}, approvalService *services.ApprovalService) *PurchaseApprovalHandler {
	return &PurchaseApprovalHandler{
		approvalService: approvalService,
	}
}

// GetApprovalWorkflows returns all approval workflows
func (h *PurchaseApprovalHandler) GetApprovalWorkflows(c *gin.Context) {
	module := c.Query("module")

	workflows, err := h.approvalService.GetWorkflows(module)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    workflows,
	})
}

// CreateApprovalWorkflow creates a new approval workflow
func (h *PurchaseApprovalHandler) CreateApprovalWorkflow(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Workflow creation is handled via database seeding",
	})
}
