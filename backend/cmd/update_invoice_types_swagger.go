package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Path to swagger.json
	swaggerPath := filepath.Join("docs", "swagger.json")

	// Read existing swagger.json
	data, err := os.ReadFile(swaggerPath)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", swaggerPath, err)
	}

	// Backup current file
	backup := fmt.Sprintf("%s.backup_%d", swaggerPath, time.Now().Unix())
	if err := os.WriteFile(backup, data, 0644); err != nil {
		log.Printf("⚠️ Failed to create backup: %v", err)
	} else {
		log.Printf("✅ Backup created: %s", backup)
	}

	// Parse JSON
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		log.Fatalf("Invalid JSON in swagger.json: %v", err)
	}

	// Ensure paths exists
	paths, _ := doc["paths"].(map[string]interface{})
	if paths == nil {
		paths = map[string]interface{}{}
		doc["paths"] = paths
	}

	// Helper to set an operation on a path
	setOp := func(path, method string, op map[string]interface{}) {
		// Ensure path map exists
		pm, _ := paths[path].(map[string]interface{})
		if pm == nil {
			pm = map[string]interface{}{}
			paths[path] = pm
		}
		// Only add if not exist to avoid overwriting
		if _, exists := pm[method]; exists {
			log.Printf("ℹ️  Skipping %s %s (already exists)", method, path)
			return
		}
		pm[method] = op
		log.Printf("➕ Added %s %s", method, path)
	}

	// Common objects
	bearerSec := []map[string]interface{}{{"BearerAuth": []string{}}}
	okResp := map[string]interface{}{
		"description": "Success",
	}
	createdResp := map[string]interface{}{
		"description": "Created",
	}
	badReq := map[string]interface{}{"description": "Bad Request"}
	notFound := map[string]interface{}{"description": "Not Found"}
	serverErr := map[string]interface{}{"description": "Internal Server Error"}

	// Define operations for Invoice Types
	// GET /api/v1/invoice-types
	setOp("/api/v1/invoice-types", "get", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "List invoice types",
		"description": "Get all invoice types, with optional active_only query to filter only active types",
		"parameters": []map[string]interface{}{
			{"name": "active_only", "in": "query", "type": "boolean", "required": false, "description": "Filter only active types"},
		},
		"responses": map[string]interface{}{
			"200": okResp, "500": serverErr,
		},
		"security": bearerSec,
	})

	// POST /api/v1/invoice-types
	setOp("/api/v1/invoice-types", "post", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "Create invoice type",
		"description": "Create a new invoice type",
		"parameters": []map[string]interface{}{
			{"name": "request", "in": "body", "required": true, "schema": map[string]interface{}{"$ref": "#/definitions/models.InvoiceTypeCreateRequest"}},
		},
		"responses": map[string]interface{}{
			"201": createdResp, "400": badReq, "401": badReq, "500": serverErr,
		},
		"security": bearerSec,
	})

	// GET /api/v1/invoice-types/{id}
	setOp("/api/v1/invoice-types/{id}", "get", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "Get invoice type",
		"description": "Get a single invoice type by ID",
		"parameters": []map[string]interface{}{
			{"name": "id", "in": "path", "type": "integer", "required": true},
		},
		"responses": map[string]interface{}{
			"200": okResp, "400": badReq, "404": notFound,
		},
		"security": bearerSec,
	})

	// PUT /api/v1/invoice-types/{id}
	setOp("/api/v1/invoice-types/{id}", "put", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "Update invoice type",
		"description": "Update an existing invoice type",
		"parameters": []map[string]interface{}{
			{"name": "id", "in": "path", "type": "integer", "required": true},
			{"name": "request", "in": "body", "required": true, "schema": map[string]interface{}{"$ref": "#/definitions/models.InvoiceTypeUpdateRequest"}},
		},
		"responses": map[string]interface{}{
			"200": okResp, "400": badReq, "500": serverErr,
		},
		"security": bearerSec,
	})

	// DELETE /api/v1/invoice-types/{id}
	setOp("/api/v1/invoice-types/{id}", "delete", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "Delete invoice type",
		"description": "Delete an invoice type (only if not used)",
		"parameters": []map[string]interface{}{
			{"name": "id", "in": "path", "type": "integer", "required": true},
		},
		"responses": map[string]interface{}{
			"200": okResp, "400": badReq, "500": serverErr,
		},
		"security": bearerSec,
	})

	// POST /api/v1/invoice-types/{id}/toggle
	setOp("/api/v1/invoice-types/{id}/toggle", "post", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "Toggle active status",
		"description": "Toggle the active status of an invoice type",
		"parameters": []map[string]interface{}{
			{"name": "id", "in": "path", "type": "integer", "required": true},
		},
		"responses": map[string]interface{}{
			"200": okResp, "400": badReq, "500": serverErr,
		},
		"security": bearerSec,
	})

	// GET /api/v1/invoice-types/active
	setOp("/api/v1/invoice-types/active", "get", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "List active invoice types",
		"description": "Get only active invoice types (for dropdowns)",
		"responses": map[string]interface{}{
			"200": okResp, "500": serverErr,
		},
		"security": bearerSec,
	})

	// POST /api/v1/invoice-types/preview-number
	setOp("/api/v1/invoice-types/preview-number", "post", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "Preview next invoice number",
		"description": "Preview what the next invoice number would be for a given type and date",
		"parameters": []map[string]interface{}{
			{"name": "request", "in": "body", "required": true, "schema": map[string]interface{}{"$ref": "#/definitions/models.InvoiceNumberRequest"}},
		},
		"responses": map[string]interface{}{
			"200": okResp, "400": badReq, "500": serverErr,
		},
		"security": bearerSec,
	})

	// GET /api/v1/invoice-types/{id}/preview
	setOp("/api/v1/invoice-types/{id}/preview", "get", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "Preview next invoice number by ID",
		"description": "Preview the next invoice number using path param and optional date query (YYYY-MM-DD)",
		"parameters": []map[string]interface{}{
			{"name": "id", "in": "path", "type": "integer", "required": true},
			{"name": "date", "in": "query", "type": "string", "required": false},
		},
		"responses": map[string]interface{}{
			"200": okResp, "400": badReq, "500": serverErr,
		},
		"security": bearerSec,
	})

	// GET /api/v1/invoice-types/{id}/counter-history
	setOp("/api/v1/invoice-types/{id}/counter-history", "get", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "Get counter history",
		"description": "Get counter history for a specific invoice type",
		"parameters": []map[string]interface{}{
			{"name": "id", "in": "path", "type": "integer", "required": true},
		},
		"responses": map[string]interface{}{
			"200": okResp, "400": badReq, "500": serverErr,
		},
		"security": bearerSec,
	})

	// POST /api/v1/invoice-types/{id}/reset-counter
	setOp("/api/v1/invoice-types/{id}/reset-counter", "post", map[string]interface{}{
		"tags":        []string{"Invoice Types"},
		"summary":     "Reset invoice counter",
		"description": "Reset counter for a specific invoice type and year",
		"parameters": []map[string]interface{}{
			{"name": "id", "in": "path", "type": "integer", "required": true},
			{"name": "request", "in": "body", "required": true, "schema": map[string]interface{}{
				"type": "object",
				"required": []string{"year", "counter"},
				"properties": map[string]interface{}{
					"year":    map[string]interface{}{"type": "integer", "example": 2025},
					"counter": map[string]interface{}{"type": "integer", "example": 150},
				},
			}},
		},
		"responses": map[string]interface{}{
			"200": okResp, "400": badReq, "500": serverErr,
		},
		"security": bearerSec,
	})

	// Ensure global tags include "Invoice Types"
	if tags, ok := doc["tags"].([]interface{}); ok {
		found := false
		for _, t := range tags {
			if m, ok := t.(map[string]interface{}); ok {
				if name, _ := m["name"].(string); name == "Invoice Types" {
					found = true
					break
				}
			}
		}
		if !found {
			doc["tags"] = append(tags, map[string]interface{}{
				"name":        "Invoice Types",
				"description": "Invoice types management and numbering - Manajemen tipe invoice dan penomoran",
			})
		}
	} else {
		doc["tags"] = []interface{}{
			map[string]interface{}{
				"name":        "Invoice Types",
				"description": "Invoice types management and numbering - Manajemen tipe invoice dan penomoran",
			},
		}
	}

	// Save updated swagger.json
	updated, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal updated swagger.json: %v", err)
	}
	if err := os.WriteFile(swaggerPath, updated, 0644); err != nil {
		log.Fatalf("Failed to write %s: %v", swaggerPath, err)
	}

	log.Println("🎉 Invoice Types endpoints merged into Swagger successfully!")
}
