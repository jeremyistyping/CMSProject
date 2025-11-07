package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// UpdateSwaggerDocs dynamically updates the generated Swagger documentation
// This allows us to override compile-time annotations with runtime configuration
func UpdateSwaggerDocs() {
	swaggerConfig := GetSwaggerConfig()
	
	// Path to the generated swagger.json
	docsPath := filepath.Join("docs", "swagger.json")
	
	// Check if swagger.json exists
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		log.Printf("Swagger docs not found at %s, skipping dynamic update", docsPath)
		return
	}
	
	// Read the existing swagger.json
	data, err := os.ReadFile(docsPath)
	if err != nil {
		log.Printf("Failed to read swagger.json: %v", err)
		return
	}
	
	// Parse the JSON
	var swaggerDoc map[string]interface{}
	if err := json.Unmarshal(data, &swaggerDoc); err != nil {
		log.Printf("Failed to parse swagger.json: %v", err)
		return
	}
	
	// Remove deprecated/unused sections (PSAK)
	filterDeprecatedPSAK(swaggerDoc)
	// Hide CashBank Integration endpoints from public docs
	filterCashBankIntegration(swaggerDoc)
	// Normalize paths to match actual backend + frontend usage
	normalizePaths(swaggerDoc)
	
	// Update the dynamic fields
	swaggerDoc["host"] = swaggerConfig.Host
	swaggerDoc["schemes"] = []string{swaggerConfig.Scheme}
	
	// Update info section if it exists
	if info, ok := swaggerDoc["info"].(map[string]interface{}); ok {
		info["title"] = swaggerConfig.Title
		info["description"] = swaggerConfig.Description
	}
	
	// Force basePath to root so paths are absolute in UI
	swaggerDoc["basePath"] = "/"

	// Apply global security (BearerAuth) and exempt public endpoints
	applyGlobalSecurity(swaggerDoc)
	
	// Marshal back to JSON
	updatedData, err := json.MarshalIndent(swaggerDoc, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal updated swagger.json: %v", err)
		return
	}
	
	// Write back to file
	if err := os.WriteFile(docsPath, updatedData, 0644); err != nil {
		log.Printf("Failed to write updated swagger.json: %v", err)
		return
	}
	
	// Also update the go docs if they exist (host/schemes only)
	updateSwaggerGoDocs(swaggerConfig)
	
	log.Printf("✅ Swagger docs updated dynamically: %s", swaggerConfig.GetSwaggerURL())
}

// applyGlobalSecurity sets default BearerAuth across all operations, and clears it for public endpoints.
func applyGlobalSecurity(swaggerDoc map[string]interface{}) {
	// Define global security requirement
	globalSec := []interface{}{map[string]interface{}{"BearerAuth": []interface{}{}}}
	swaggerDoc["security"] = globalSec

	paths, ok := swaggerDoc["paths"].(map[string]interface{})
	if !ok {
		return
	}

	// Public endpoints (no auth required)
	publicPrefixes := []string{
		"/auth/login", 
		"/api/v1/auth/login",
		"/auth/register",
		"/api/v1/auth/register",
		"/auth/refresh",
		"/api/v1/auth/refresh",
		"/api/v1/health",
		"/health",
	}

	// Helper to mark operation-level security as empty (override global)
	clearOpSecurity := func(op map[string]interface{}) {
		op["security"] = []interface{}{}
	}

	for p, v := range paths {
		isPublic := false
		for _, pref := range publicPrefixes {
			if strings.EqualFold(p, pref) {
				isPublic = true
				break
			}
		}
		if !isPublic {
			continue
		}
		// For each method under the public path, clear security
		if m, ok := v.(map[string]interface{}); ok {
			for _, method := range []string{"get", "post", "put", "delete", "patch", "options", "head"} {
				if opRaw, ok := m[method]; ok {
					if op, ok := opRaw.(map[string]interface{}); ok {
						clearOpSecurity(op)
					}
				}
			}
		}
	}
}

// normalizePaths updates path keys to align with actual routes used by frontend/backend
func normalizePaths(swaggerDoc map[string]interface{}) {
	paths, ok := swaggerDoc["paths"].(map[string]interface{})
	if !ok {
		return
	}

	changes := make(map[string]interface{})
	removed := 0
	for p, v := range paths {
		newPath := p

		// Fix common prefix issues by adding /api/v1 and correcting segment names
		if strings.HasPrefix(newPath, "/api/cashbank") {
			newPath = strings.Replace(newPath, "/api/cashbank", "/api/v1/cash-bank", 1)
		}
		if strings.HasPrefix(newPath, "/api/monitoring") {
			newPath = strings.Replace(newPath, "/api/monitoring", "/api/v1/monitoring", 1)
		}
		if strings.HasPrefix(newPath, "/api/payments") {
			newPath = strings.Replace(newPath, "/api/payments", "/api/v1/payments", 1)
		}
		if strings.HasPrefix(newPath, "/api/purchases") {
			newPath = strings.Replace(newPath, "/api/purchases", "/api/v1/purchases", 1)
		}
		if strings.HasPrefix(newPath, "/api/admin") {
			newPath = strings.Replace(newPath, "/api/admin", "/api/v1/admin", 1)
		}

		// Normalize root-level routes expected to live under /api/v1
		if strings.HasPrefix(newPath, "/auth") ||
			strings.HasPrefix(newPath, "/profile") ||
			strings.HasPrefix(newPath, "/dashboard/") ||
			strings.HasPrefix(newPath, "/journal-drilldown") ||
			strings.HasPrefix(newPath, "/monitoring/") {
			newPath = "/api/v1" + newPath
		}

		// If nothing changed, continue
		if newPath == p {
			continue
		}

		// Plan change: move value to new key
		changes[newPath] = v
		delete(paths, p)
		removed++
	}

	// Apply changes
	for np, val := range changes {
		paths[np] = val
	}

	if removed > 0 {
		log.Printf("🔧 Normalized %d Swagger path(s) to align with /api/v1 and naming conventions", removed)
	}
}

// filterDeprecatedPSAK removes PSAK endpoints and tags from swaggerDoc in-place
func filterDeprecatedPSAK(swaggerDoc map[string]interface{}) {
	const psakPrefix = "/api/v1/reports/psak"

	// Remove PSAK paths
	if paths, ok := swaggerDoc["paths"].(map[string]interface{}); ok {
		removed := 0
		for p := range paths {
			if strings.HasPrefix(p, psakPrefix) {
				delete(paths, p)
				removed++
			}
		}
		if removed > 0 {
			log.Printf("🧹 Removed %d PSAK path(s) from Swagger", removed)
		}
	}

	// Remove PSAK tag from top-level tags array
	if tags, ok := swaggerDoc["tags"].([]interface{}); ok {
		filtered := make([]interface{}, 0, len(tags))
		for _, t := range tags {
			m, _ := t.(map[string]interface{})
			if m != nil {
				if name, _ := m["name"].(string); strings.EqualFold(name, "PSAK Reports") {
					continue // skip PSAK tag
				}
			}
			filtered = append(filtered, t)
		}
		swaggerDoc["tags"] = filtered
	}
}

// filterCashBankIntegration removes CashBank Integration endpoints and tag from swaggerDoc in-place
func filterCashBankIntegration(swaggerDoc map[string]interface{}) {
	// We hide all endpoints related to CashBank Integration from public docs
	integratedPrefixes := []string{
		"/api/cashbank/integrated",
		"/api/v1/cash-bank/integrated",
		"/api/v1/cash-bank/ssot",
	}

	// Remove paths for CashBank Integration
	if paths, ok := swaggerDoc["paths"].(map[string]interface{}); ok {
		removed := 0
		for p := range paths {
			for _, pref := range integratedPrefixes {
				if strings.HasPrefix(p, pref) || strings.Contains(p, pref) {
					delete(paths, p)
					removed++
					break
				}
			}
		}
		if removed > 0 {
			log.Printf("🧹 Removed %d CashBank Integration path(s) from Swagger", removed)
		}
	}

	// Remove the "CashBank Integration" tag from top-level tags array
	if tags, ok := swaggerDoc["tags"].([]interface{}); ok {
		filtered := make([]interface{}, 0, len(tags))
		for _, t := range tags {
			if m, ok := t.(map[string]interface{}); ok {
				if name, _ := m["name"].(string); strings.EqualFold(name, "CashBank Integration") {
					continue // skip this tag to hide from UI
				}
			}
			filtered = append(filtered, t)
		}
		swaggerDoc["tags"] = filtered
	}
}

// updateSwaggerGoDocs updates the generated docs.go file with dynamic values
func updateSwaggerGoDocs(config *SwaggerConfig) {
	docsGoPath := filepath.Join("docs", "docs.go")
	
	if _, err := os.Stat(docsGoPath); os.IsNotExist(err) {
		return // docs.go doesn't exist, skip
	}
	
	// Read the docs.go file
	data, err := os.ReadFile(docsGoPath)
	if err != nil {
		log.Printf("Failed to read docs.go: %v", err)
		return
	}
	
	content := string(data)
	
	// Replace the host, schemes, and other dynamic values
	content = replaceInSwaggerDoc(content, `\"host\":`, fmt.Sprintf(`\"host\": \"%s\"`, config.Host))
	content = replaceInSwaggerDoc(content, `\"schemes\":`, fmt.Sprintf(`\"schemes\": [\"%s\"]`, config.Scheme))
	
	// Write back
	if err := os.WriteFile(docsGoPath, []byte(content), 0644); err != nil {
		log.Printf("Failed to write updated docs.go: %v", err)
	}
}

// replaceInSwaggerDoc replaces a field in the swagger doc string
func replaceInSwaggerDoc(content, field, replacement string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, field) {
			// Find the start and end of the field value
			start := strings.Index(line, field)
			if start != -1 {
				// Find the comma or closing brace
				end := strings.Index(line[start:], ",")
				if end == -1 {
					end = strings.Index(line[start:], "}")
				}
				if end != -1 {
					end += start
					lines[i] = line[:start] + replacement + "," + line[end+1:]
				}
			}
			break
		}
	}
	return strings.Join(lines, "\n")
}

// PrintSwaggerInfo prints helpful Swagger configuration information
func PrintSwaggerInfo() {
	cfg := LoadConfig()
	swaggerConfig := GetSwaggerConfig()
	
	fmt.Printf("🚀 Swagger Configuration:\n")
	fmt.Printf("   Environment: %s\n", cfg.Environment)
	fmt.Printf("   Swagger URL: %s\n", swaggerConfig.GetSwaggerURL())
	fmt.Printf("   API Base URL: %s\n", swaggerConfig.GetAPIBaseURL())
	fmt.Printf("   Host: %s\n", swaggerConfig.Host)
	fmt.Printf("   Scheme: %s\n", swaggerConfig.Scheme)
	
	// Print CORS origins
	origins := GetAllowedOrigins(cfg)
	fmt.Printf("   CORS Origins: %v\n", origins)
	
	// Print environment variable hints
	if cfg.Environment == "production" {
		fmt.Printf("\n💡 Production Environment Variables:\n")
		fmt.Printf("   SWAGGER_HOST: Set your production domain (e.g., api.yourdomain.com)\n")
		fmt.Printf("   SWAGGER_SCHEME: https (recommended for production)\n")
		fmt.Printf("   ALLOWED_ORIGINS: Your frontend URL(s)\n")
		fmt.Printf("   DOMAIN or APP_URL: Your main domain\n")
		fmt.Printf("   ENABLE_HTTPS: true (recommended for production)\n")
	} else {
		fmt.Printf("\n💡 Development Mode - Dynamic Configuration Active\n")
		fmt.Printf("   To override: Set SWAGGER_HOST, SWAGGER_SCHEME, ALLOWED_ORIGINS\n")
	}
	fmt.Printf("\n")
}
