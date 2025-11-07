package controllers

import (
	"net/http"
	"encoding/json"
	"io"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
)

// ContactController handles HTTP requests for Contacts
type ContactController struct {
	contactService services.ContactService
}

// NewContactController creates a new ContactController
func NewContactController(contactService services.ContactService) *ContactController {
	return &ContactController{contactService: contactService}
}

// GetContacts returns a list of contacts
func (cc *ContactController) GetContacts(c *gin.Context) {
	// Check if type query parameter is provided
	contactType := c.Query("type")
	
	var contacts []models.Contact
	var err error
	
	if contactType != "" {
		// If type is specified, filter by type
		contacts, err = cc.contactService.GetContactsByType(contactType)
	} else {
		// Otherwise, get all contacts
		contacts, err = cc.contactService.GetAllContacts()
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": contacts})
}

// GetContact returns a contact by ID
func (cc *ContactController) GetContact(c *gin.Context) {
	id := c.Param("id")
	contact, err := cc.contactService.GetContactByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contact)
}

// CreateContact creates a new contact
func (cc *ContactController) CreateContact(c *gin.Context) {
	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newContact, err := cc.contactService.CreateContact(contact)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newContact)
}

// UpdateContact updates an existing contact by ID
func (cc *ContactController) UpdateContact(c *gin.Context) {
	id := c.Param("id")
	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedContact, err := cc.contactService.UpdateContact(id, contact)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedContact)
}

// DeleteContact deletes a contact by ID
func (cc *ContactController) DeleteContact(c *gin.Context) {
	id := c.Param("id")
	if err := cc.contactService.DeleteContact(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// GetContactsByType returns contacts filtered by type
func (cc *ContactController) GetContactsByType(c *gin.Context) {
	contactType := c.Param("type")
	contacts, err := cc.contactService.GetContactsByType(contactType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": contacts})
}

// SearchContacts searches contacts by query
func (cc *ContactController) SearchContacts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}
	
	contacts, err := cc.contactService.SearchContacts(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": contacts})
}

// ImportContacts imports contacts from JSON
func (cc *ContactController) ImportContacts(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	
	var contacts []models.Contact
	if err := json.Unmarshal(body, &contacts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	
	if err := cc.contactService.ImportContacts(contacts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Contacts imported successfully", "count": len(contacts)})
}

// ExportContacts exports all contacts as JSON
func (cc *ContactController) ExportContacts(c *gin.Context) {
	contacts, err := cc.contactService.ExportContacts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=contacts.json")
	c.JSON(http.StatusOK, contacts)
}

// AddContactAddress adds a new address to a contact
func (cc *ContactController) AddContactAddress(c *gin.Context) {
	id := c.Param("id")
	var address models.ContactAddress
	if err := c.ShouldBindJSON(&address); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	newAddress, err := cc.contactService.AddContactAddress(id, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newAddress)
}

// UpdateContactAddress updates an existing contact address
func (cc *ContactController) UpdateContactAddress(c *gin.Context) {
	contactID := c.Param("id")
	addressID := c.Param("address_id")
	var address models.ContactAddress
	if err := c.ShouldBindJSON(&address); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	updatedAddress, err := cc.contactService.UpdateContactAddress(contactID, addressID, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedAddress)
}

// DeleteContactAddress deletes a contact address
func (cc *ContactController) DeleteContactAddress(c *gin.Context) {
	contactID := c.Param("id")
	addressID := c.Param("address_id")
	
	if err := cc.contactService.DeleteContactAddress(contactID, addressID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

