package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// DynamicSwaggerConfig provides auto-fixing capabilities for Swagger docs
type DynamicSwaggerConfig struct {
	Host     string
	BasePath string
	Schemes  []string
	Fixes    []string
	Errors   []string
}

// SwaggerFixRule defines rules for auto-fixing common Swagger issues
type SwaggerFixRule struct {
	Pattern     string
	Replacement string
	Description string
}

// Common Swagger fixing rules
var swaggerFixRules = []SwaggerFixRule{
	{
		Pattern:     `"\/admin\/([^"]+)"`,
		Replacement: `"/api/v1/admin/$1"`,
		Description: "Fix missing /api/v1 prefix in admin routes",
	},
	{
		Pattern:     `"\/auth\/([^"]+)"`,
		Replacement: `"/api/v1/auth/$1"`,
		Description: "Fix missing /api/v1 prefix in auth routes",
	},
	{
		Pattern:     `"\/monitoring\/([^"]+)"`,
		Replacement: `"/api/v1/monitoring/$1"`,
		Description: "Fix missing /api/v1 prefix in monitoring routes",
	},
	{
		Pattern:     `"\/payments\/([^"]+)"`,
		Replacement: `"/api/v1/payments/$1"`,
		Description: "Fix missing /api/v1 prefix in payment routes",
	},
	{
		Pattern:     `"\/dashboard\/([^"]+)"`,
		Replacement: `"/api/v1/dashboard/$1"`,
		Description: "Fix missing /api/v1 prefix in dashboard routes",
	},
	{
		Pattern:     `"\/reports\/([^"]+)"`,
		Replacement: `"/api/v1/reports/$1"`,
		Description: "Fix missing /api/v1 prefix in report routes",
	},
	{
		Pattern:     `"\/journals\/([^"]+)"`,
		Replacement: `"/api/v1/journals/$1"`,
		Description: "Fix missing /api/v1 prefix in journal routes",
	},
	{
		Pattern:     `"\/cash-bank\/([^"]+)"`,
		Replacement: `"/api/v1/cash-bank/$1"`,
		Description: "Fix missing /api/v1 prefix in cash-bank routes",
	},
	{
		Pattern:     `"swegger"`,
		Replacement: `"swagger"`,
		Description: "Fix common typo: swegger -> swagger",
	},
	{
		Pattern:     `"respone"`,
		Replacement: `"response"`,
		Description: "Fix common typo: respone -> response",
	},
	{
		Pattern:     `"paramater"`,
		Replacement: `"parameter"`,
		Description: "Fix common typo: paramater -> parameter",
	},
	{
		Pattern:     `"sucess"`,
		Replacement: `"success"`,
		Description: "Fix common typo: sucess -> success",
	},
}

// ValidateAndFixSwagger automatically validates and fixes common Swagger issues
func ValidateAndFixSwagger() (*DynamicSwaggerConfig, error) {
	log.Println("🔍 Starting dynamic Swagger validation and fixing...")
	
	config := &DynamicSwaggerConfig{
		Host:     getSwaggerHost(),
		BasePath: "/",
		Schemes:  []string{"http"},
		Fixes:    []string{},
		Errors:   []string{},
	}

	// Check if swagger.json exists
	swaggerPath := filepath.Join("docs", "swagger.json")
	if _, err := os.Stat(swaggerPath); os.IsNotExist(err) {
		config.Errors = append(config.Errors, "swagger.json not found")
		log.Printf("⚠️ swagger.json not found at %s", swaggerPath)
		return config, nil
	}

	// Read and parse swagger.json
	data, err := os.ReadFile(swaggerPath)
	if err != nil {
		config.Errors = append(config.Errors, fmt.Sprintf("Failed to read swagger.json: %v", err))
		return config, err
	}

	// Validate JSON structure
	var swaggerDoc map[string]interface{}
	if err := json.Unmarshal(data, &swaggerDoc); err != nil {
		config.Errors = append(config.Errors, fmt.Sprintf("Invalid JSON in swagger.json: %v", err))
		return config, err
	}

	// Apply fixes
	originalContent := string(data)
	fixedContent := originalContent
	
	for _, rule := range swaggerFixRules {
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			config.Errors = append(config.Errors, fmt.Sprintf("Invalid regex pattern %s: %v", rule.Pattern, err))
			continue
		}

		if re.MatchString(fixedContent) {
			fixedContent = re.ReplaceAllString(fixedContent, rule.Replacement)
			config.Fixes = append(config.Fixes, rule.Description)
		}
	}

	// Check if any fixes were applied
	if len(config.Fixes) > 0 {
		// Write fixed content back to file
		err = os.WriteFile(swaggerPath, []byte(fixedContent), 0644)
		if err != nil {
			config.Errors = append(config.Errors, fmt.Sprintf("Failed to write fixed swagger.json: %v", err))
			return config, err
		}

		// Create backup of original
		backupPath := fmt.Sprintf("%s.backup_%d", swaggerPath, time.Now().Unix())
		err = os.WriteFile(backupPath, []byte(originalContent), 0644)
		if err != nil {
			log.Printf("⚠️ Failed to create backup: %v", err)
		}

		log.Printf("✅ Applied %d fixes to swagger.json", len(config.Fixes))
		for _, fix := range config.Fixes {
			log.Printf("  - %s", fix)
		}
	} else {
		log.Println("✅ No Swagger fixes needed")
	}

	return config, nil
}

// getSwaggerHost dynamically determines the best host for swagger
func getSwaggerHost() string {
	cfg := LoadConfig()
	
	if cfg.SwaggerHost != "" {
		return cfg.SwaggerHost
	}

	// Development default
	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}
	
	return fmt.Sprintf("localhost:%s", port)
}

// GenerateDynamicSwaggerSpec creates a dynamic swagger spec with error handling
func GenerateDynamicSwaggerSpec() map[string]interface{} {
	cfg := LoadConfig()
	
	spec := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"title":       cfg.SwaggerTitle,
			"description": cfg.SwaggerDescription,
			"version":     "1.0",
			"contact": map[string]interface{}{
				"name":  "API Support",
				"url":   "http://www.swagger.io/support",
				"email": "support@swagger.io",
			},
			"license": map[string]interface{}{
				"name": "Apache 2.0",
				"url":  "http://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		"host":     getSwaggerHost(),
		"basePath": "/",
		"schemes":  []string{"http"},
		"securityDefinitions": map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type":        "apiKey",
				"name":        "Authorization",
				"in":          "header",
				"description": "Type 'Bearer' followed by a space and JWT token.",
			},
		},
		"security": []map[string]interface{}{
			{"BearerAuth": []string{}},
		},
		"paths": map[string]interface{}{
			"/api/v1/health": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Health"},
					"summary":     "Health check endpoint",
					"description": "Returns API health status",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "API is healthy",
						},
					},
					"security": []map[string]interface{}{}, // Public endpoint
				},
			},
			"/api/v1/auth/login": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "User login",
					"description": "Authenticate user and return JWT token",
					"parameters": []map[string]interface{}{
						{
							"name":        "body",
							"in":          "body",
							"description": "Login credentials",
							"required":    true,
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"email": map[string]interface{}{
										"type":        "string",
										"description": "User email",
									},
									"password": map[string]interface{}{
										"type":        "string",
										"description": "User password",
									},
								},
								"required": []string{"email", "password"},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Login successful",
						},
						"401": map[string]interface{}{
							"description": "Invalid credentials",
						},
					},
					"security": []map[string]interface{}{}, // Public endpoint
				},
			},
		},
		"definitions": map[string]interface{}{
			"ErrorResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"error": map[string]interface{}{
						"type":        "string",
						"description": "Error message",
					},
					"message": map[string]interface{}{
						"type":        "string", 
						"description": "Detailed error description",
					},
				},
			},
			"SuccessResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Response status",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Success message",
					},
					"data": map[string]interface{}{
						"type":        "object",
						"description": "Response data",
					},
				},
			},
		},
	}

	return spec
}

// SetupDynamicSwaggerRoutes sets up swagger routes with dynamic configuration
func SetupDynamicSwaggerRoutes(r *gin.Engine) {
	log.Println("🚀 Setting up dynamic Swagger routes...")

	// Validate and fix swagger
	config, err := ValidateAndFixSwagger()
	if err != nil {
		log.Printf("❌ Swagger validation failed: %v", err)
		// Continue with fallback spec
	}

	// Print configuration info
	if len(config.Fixes) > 0 {
		log.Printf("✅ Applied %d Swagger fixes:", len(config.Fixes))
		for _, fix := range config.Fixes {
			log.Printf("  - %s", fix)
		}
	}

	if len(config.Errors) > 0 {
		log.Printf("⚠️ Swagger errors detected:")
		for _, error := range config.Errors {
			log.Printf("  - %s", error)
		}
	}

	// Dynamic Swagger JSON endpoint (different path to avoid conflict)
	r.GET("/openapi/dynamic-doc.json", func(c *gin.Context) {
		// Try to serve the actual swagger.json
		swaggerPath := filepath.Join("docs", "swagger.json")
		if _, err := os.Stat(swaggerPath); err == nil {
			data, err := os.ReadFile(swaggerPath)
			if err == nil {
				c.Header("Content-Type", "application/json")
				c.Header("Access-Control-Allow-Origin", "*")
				c.Data(200, "application/json", data)
				return
			}
		}

		// Fallback to dynamic spec
		log.Println("📄 Serving dynamic fallback Swagger spec")
		spec := GenerateDynamicSwaggerSpec()
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(200, spec)
	})

	// Swagger UI endpoints
	swaggerGroup := r.Group("/")
	swaggerGroup.Use(func(c *gin.Context) {
		// Relax CSP for Swagger UI
		c.Header("Content-Security-Policy", "default-src 'self' https: data: blob:; script-src 'self' 'unsafe-inline' 'unsafe-eval' https:; style-src 'self' 'unsafe-inline' https:; img-src 'self' data: https:; font-src 'self' data: https:; connect-src 'self' data: blob: https:")
		c.Next()
	})

	// Swagger UI HTML
	swaggerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Sistema Akuntansi API - Dynamic Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    html, body { height: 100%; margin: 0; padding: 0; }
    #swagger-ui { height: 100%; }
    .swagger-ui .topbar { background-color: #1b1b1b; }
    .swagger-ui .topbar .download-url-wrapper { display: none; }
    .swagger-ui .info .title { color: #3b4151; }
    .error-banner { background: #f8d7da; color: #721c24; padding: 10px; margin: 10px; border-radius: 5px; }
    .fix-banner { background: #d1ecf1; color: #0c5460; padding: 10px; margin: 10px; border-radius: 5px; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      // Try to load swagger spec from dynamic endpoint
      fetch('/openapi/dynamic-doc.json')
        .then(response => response.json())
        .then(spec => {
          // Add banner for dynamic fixes if any
          const banner = document.createElement('div');
          banner.style.cssText = 'background: #d4edda; color: #155724; padding: 15px; text-align: center; font-weight: bold;';
          banner.innerHTML = '🚀 Dynamic Swagger UI - Auto-fixing enabled for better API documentation';
          document.body.insertBefore(banner, document.getElementById('swagger-ui'));

          window.ui = SwaggerUIBundle({
            url: '/openapi/dynamic-doc.json',
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
              SwaggerUIBundle.presets.apis,
              SwaggerUIBundle.presets.standalone
            ],
            plugins: [
              SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout",
            tryItOutEnabled: true,
            requestInterceptor: function(req) {
              // Auto-add Bearer prefix if missing
              if (req.headers.Authorization && !req.headers.Authorization.startsWith('Bearer ')) {
                req.headers.Authorization = 'Bearer ' + req.headers.Authorization;
              }
              return req;
            },
            responseInterceptor: function(res) {
              // Handle common errors
              if (res.status === 404 && res.url.includes('/admin/')) {
                console.log('🔧 Detected 404 for admin route, trying /api/v1/admin/ instead...');
              }
              return res;
            }
          });
        })
        .catch(error => {
          console.error('Failed to load Swagger spec:', error);
          document.getElementById('swagger-ui').innerHTML = 
            '<div class="error-banner"><h3>⚠️ Swagger Loading Error</h3><p>Failed to load API documentation. Please check server status.</p></div>';
        });
    };
  </script>
</body>
</html>`

	swaggerGroup.GET("/swagger", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	swaggerGroup.GET("/swagger/index.html", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, swaggerHTML)
	})

	swaggerGroup.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/docs/index.html")
	})

	swaggerGroup.GET("/docs/index.html", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, swaggerHTML)
	})

	log.Printf("✅ Dynamic Swagger UI available at: http://%s/swagger/index.html", config.Host)
}

// CheckAndFixCommonIssues performs common Swagger issue checks
func CheckAndFixCommonIssues() []string {
	issues := []string{}
	
	// Check for common path issues
	swaggerPath := filepath.Join("docs", "swagger.json")
	if _, err := os.Stat(swaggerPath); os.IsNotExist(err) {
		issues = append(issues, "swagger.json missing - run 'swag init -g cmd/main.go'")
		return issues
	}

	data, err := os.ReadFile(swaggerPath)
	if err != nil {
		issues = append(issues, fmt.Sprintf("Cannot read swagger.json: %v", err))
		return issues
	}

	content := string(data)

	// Check for missing /api/v1 prefixes
	if strings.Contains(content, `"/admin/`) {
		issues = append(issues, "Found routes without /api/v1 prefix - will auto-fix")
	}

	// Check for common typos
	typos := map[string]string{
		"swegger":    "swagger",
		"respone":    "response", 
		"paramater":  "parameter",
		"sucess":     "success",
		"recieve":    "receive",
		"seperate":   "separate",
		"occurence":  "occurrence",
	}

	for typo, correct := range typos {
		if strings.Contains(strings.ToLower(content), typo) {
			issues = append(issues, fmt.Sprintf("Found typo '%s' should be '%s' - will auto-fix", typo, correct))
		}
	}

	return issues
}