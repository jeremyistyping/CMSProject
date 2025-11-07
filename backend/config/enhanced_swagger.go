package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

// EnhancedSwaggerConfig provides comprehensive Swagger configuration with authentication support
type EnhancedSwaggerConfig struct {
	Host                string
	BasePath            string
	Schemes             []string
	Fixes               []string
	Errors              []string
	AuthenticationReady bool
	SecurityDefinitions map[string]interface{}
}

// publicSwaggerAllowedExact lists exact path entries we want to keep visible in public docs
var publicSwaggerAllowedExact = map[string]struct{}{
	// Auth & health
	"/api/v1/health":      {},
	"/api/v1/auth/login": {},

	// Sales
	"/api/v1/sales":                 {},
	"/api/v1/sales/{id}":            {},
	"/api/v1/sales/{id}/confirm":    {},
	"/api/v1/sales/{id}/invoice":    {},
	"/api/v1/sales/{id}/cancel":     {},
	"/api/v1/sales/validate-stock":  {},
	"/api/v1/sales/{id}/invoice/pdf": {}, // may not exist in spec but safe to allow
	"/api/v1/sales/{id}/receipt/pdf": {},
	"/api/v1/sales/report/pdf":      {},
	"/api/v1/sales/report/csv":      {},

	// Purchases
	"/api/v1/purchases":                        {},
	"/api/v1/purchases/{id}":                   {},
	"/api/v1/purchases/{id}/approve":           {},
	"/api/v1/purchases/{id}/reject":            {},
	"/api/v1/purchases/{id}/submit-approval":   {},
	"/api/v1/purchases/{id}/receive":           {},
	"/api/v1/purchases/{id}/integrated-payment": {},
	"/api/v1/purchases/{id}/payments":          {},
	"/api/v1/purchases/{id}/for-payment":       {},
	"/api/v1/purchases/{id}/cancel":            {},

	// Payments (SSOT preferred if present; also include safe exports and details)
	"/api/v1/payments/ssot/receivable":       {},
	"/api/v1/payments/ssot/payable":          {},
	"/api/v1/payments/ssot/{id}":             {},
	"/api/v1/payments/ssot/{id}/reverse":     {},
	"/api/v1/payments/ssot/preview-journal":  {},
	"/api/v1/payments/ssot/{id}/balance-updates": {},
	"/api/v1/payments/summary":                {},
	"/api/v1/payments/report/pdf":             {},
	"/api/v1/payments/export/excel":           {},
	"/api/v1/payments/unpaid-bills/{vendor_id}":    {},
	"/api/v1/payments/unpaid-invoices/{customer_id}": {},
	"/api/v1/payments/{id}":                         {},
	"/api/v1/payments/{id}/pdf":                    {},
	"/api/v1/payments/{id}/cancel":                 {},
	"/api/v1/payments/{id}/with-journal":           {},
	"/api/v1/payments/{id}/account-updates":        {},
	"/api/v1/payments/preview-journal":             {},

	// Cash & Bank, Accounts, Products, Contacts, Users (supporting lists and CRUD)
	"/api/v1/cash-bank/accounts":       {},
	"/api/v1/cash-bank/accounts/{id}":  {},
	"/api/v1/cash-bank/accounts/{id}/transactions": {},
	"/api/v1/accounts":                 {},
	"/api/v1/accounts/{id}":            {},
	"/api/v1/products":                 {},
	"/api/v1/products/{id}":            {},
	"/api/v1/contacts":                 {},
	"/api/v1/contacts/{id}":            {},
	"/api/v1/users":                    {},
	"/api/v1/users/{id}":               {},

	// Assets management (Fixed Assets and Depreciation)
	"/api/v1/assets":                              {},
	"/api/v1/assets/{id}":                         {},
	"/api/v1/assets/{id}/capitalize":              {},
	"/api/v1/assets/{id}/depreciation-schedule":   {},
	"/api/v1/assets/categories":                   {},
	"/api/v1/assets/summary":                      {},

	// Settings and configuration
	"/api/v1/settings":                    {},
	"/api/v1/settings/company":            {},
	"/api/v1/settings/system":             {},
	"/api/v1/settings/reset":              {},
	"/api/v1/settings/validation-rules":   {},
	"/api/v1/settings/history":            {},

	// Tax Account Settings (based on settings_routes.go analysis)
	"/api/v1/tax-accounts":                {},  // GetAllSettings (admin)
	"/api/v1/tax-accounts/current":        {},  // GetCurrentSettings
	"/api/v1/tax-accounts/{id}":           {},  // UpdateSettings
	"/api/v1/tax-accounts/{id}/activate":  {},  // ActivateSettings
	"/api/v1/tax-accounts/accounts":       {},  // GetAvailableAccounts
	"/api/v1/tax-accounts/validate":       {},  // ValidateAccountConfiguration & ValidateAccountSelection
	"/api/v1/tax-accounts/refresh-cache":  {},  // RefreshCache (admin)
	"/api/v1/tax-accounts/suggestions":    {},  // GetAccountSuggestions
	"/api/v1/tax-accounts/status":         {},  // GetStatus
}

// publicSwaggerAllowedPrefixes allows quick inclusion by prefix (kept narrow to avoid pulling in unrelated modules)
var publicSwaggerAllowedPrefixes = []string{
	"/api/v1/sales",
	"/api/v1/purchases",
	"/api/v1/payments/ssot",
	"/api/v1/cash-bank",
	"/api/v1/products",
	"/api/v1/contacts",
	"/api/v1/users",
	"/api/v1/accounts",
	"/api/v1/assets",
	"/api/v1/settings",
	"/api/v1/tax-accounts",
	"/api/v1/invoice-types",
	// Include all SSOT Reports endpoints used by the Reports SSOT page
	"/api/v1/ssot-reports",
}

// filterSwaggerSpecForPublic prunes the swagger spec to only include allowed paths according to the whitelist
func filterSwaggerSpecForPublic(spec map[string]interface{}) map[string]interface{} {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok || len(paths) == 0 {
		return spec
	}

	filtered := make(map[string]interface{})

	// helper to test if path is allowed by exact or prefix
	isAllowed := func(p string) bool {
		if _, ok := publicSwaggerAllowedExact[p]; ok {
			return true
		}
		for _, pref := range publicSwaggerAllowedPrefixes {
			if len(p) >= len(pref) && p[:len(pref)] == pref {
				return true
			}
		}
		return false
	}

	for p, v := range paths {
		if isAllowed(p) {
			filtered[p] = v
		}
	}

	// Replace paths with filtered ones
	spec["paths"] = filtered

	// Optionally annotate description to indicate this is a pruned public view
	if info, ok := spec["info"].(map[string]interface{}); ok {
		desc := "Public API documentation limited to Sales, Purchases, Payments and essential supporting resources."
		if existing, ok := info["description"].(string); ok && existing != "" {
			info["description"] = existing + "\n\nNOTE: Non-essential/internal endpoints are hidden in this public view."
		} else {
			info["description"] = desc
		}
	}

	return spec
}

// AuthenticationHelperJS provides JavaScript for Swagger UI authentication
func getAuthenticationHelperJS() string {
	return `
// Enhanced Authentication Helper for Swagger UI (Fixed Version)
class SwaggerAuthHelper {
    constructor() {
        this.token = localStorage.getItem('swagger_auth_token') || '';
        this.isSwaggerReady = false;
        this.pendingAuth = false;
        
        // Wait for DOM to be ready
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => this.setupUI());
        } else {
            this.setupUI();
        }
    }

    setupUI() {
        // Add login interface
        const authContainer = document.createElement('div');
        authContainer.id = 'swagger-auth-helper';
        
        const authHTML = '<div style="position: fixed; top: 10px; right: 10px; z-index: 9999; ' +
            'background: white; padding: 15px; border-radius: 8px; ' +
            'box-shadow: 0 2px 10px rgba(0,0,0,0.1); border: 1px solid #ddd;">' +
            '<div style="font-weight: bold; margin-bottom: 10px; color: #3b4151;">' +
            '🔐 API Authentication (Fixed)</div>' +
            '<div id="auth-status" style="margin-bottom: 10px; font-size: 12px;">' +
            'Status: <span id="status-text">Not authenticated</span></div>' +
            '<div style="margin-bottom: 10px;">' +
            '<input type="email" id="login-email" placeholder="Email (admin@company.com)" ' +
            'style="width: 200px; padding: 5px; margin-bottom: 5px; border: 1px solid #ddd; border-radius: 4px;">' +
            '<input type="password" id="login-password" placeholder="Password (admin123)" ' +
            'style="width: 200px; padding: 5px; margin-bottom: 5px; border: 1px solid #ddd; border-radius: 4px;">' +
            '</div><div style="display: flex; gap: 5px;">' +
            '<button id="login-btn" onclick="authHelper.login()" ' +
            'style="padding: 5px 10px; background: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer;">' +
            'Login</button>' +
            '<button id="logout-btn" onclick="authHelper.logout()" ' +
            'style="padding: 5px 10px; background: #f44336; color: white; border: none; border-radius: 4px; cursor: pointer;">' +
            'Logout</button>' +
            '<button onclick="authHelper.toggleUI()" ' +
            'style="padding: 5px 10px; background: #008CBA; color: white; border: none; border-radius: 4px; cursor: pointer;">' +
            'Toggle</button>' +
            '<button onclick="authHelper.debugAuth()" ' +
            'style="padding: 5px 10px; background: #ff9800; color: white; border: none; border-radius: 4px; cursor: pointer;">' +
            'Debug</button></div>' +
            '<div style="margin-top: 10px; font-size: 11px; color: #666;">' +
            'Fixed: Headers will be properly set</div></div>';
        
        authContainer.innerHTML = authHTML;
        document.body.appendChild(authContainer);
        
        // Set default credentials with delay
        setTimeout(() => {
            const emailInput = document.getElementById('login-email');
            const passwordInput = document.getElementById('login-password');
            if (emailInput && passwordInput) {
                emailInput.value = 'admin@company.com';
                passwordInput.value = 'admin123';
            }
        }, 100);
        
        this.updateStatus();
        
        // Watch for Swagger UI initialization
        this.waitForSwaggerUI();
    }

    // Wait for Swagger UI to be ready
    waitForSwaggerUI() {
        const checkSwaggerUI = () => {
            if (window.ui) {
                this.isSwaggerReady = true;
                console.log('✅ Swagger UI is ready');
                
                // Apply saved token if exists
                if (this.token && !this.pendingAuth) {
                    this.applyTokenToSwagger();
                }
            } else {
                setTimeout(checkSwaggerUI, 100);
            }
        };
        checkSwaggerUI();
    }
    
    // Apply token to Swagger UI with proper error handling
    applyTokenToSwagger() {
        if (!this.isSwaggerReady || !window.ui) {
            console.log('⏳ Swagger UI not ready yet, queuing token application');
            this.pendingAuth = true;
            return;
        }
        
        try {
            // Clear any existing authorization first
            window.ui.preauthorizeApiKey('BearerAuth', '');
            
            // Apply new token
            window.ui.preauthorizeApiKey('BearerAuth', 'Bearer ' + this.token);
            console.log('🔐 Successfully applied token to Swagger UI:', this.token.substring(0, 20) + '...');
            
            this.pendingAuth = false;
        } catch (error) {
            console.error('❌ Error applying token to Swagger UI:', error);
        }
    }
    
    // Debug authentication status
    debugAuth() {
        const debugInfo = {
            hasToken: !!this.token,
            tokenPreview: this.token ? this.token.substring(0, 50) + '...' : 'None',
            swaggerReady: this.isSwaggerReady,
            windowUI: !!window.ui,
            localStorage: {
                swaggerToken: localStorage.getItem('swagger_auth_token'),
                allKeys: Object.keys(localStorage)
            },
            pendingAuth: this.pendingAuth
        };
        
        console.group('🔍 Swagger Auth Debug Info');
        console.log('Token Status:', debugInfo.hasToken ? '✅ Present' : '❌ Missing');
        console.log('Token Preview:', debugInfo.tokenPreview);
        console.log('Swagger UI Ready:', debugInfo.swaggerReady ? '✅ Ready' : '❌ Not Ready');
        console.log('Window.ui Available:', debugInfo.windowUI ? '✅ Available' : '❌ Missing');
        console.log('Pending Auth:', debugInfo.pendingAuth ? '⏳ Pending' : '✅ Clear');
        console.log('LocalStorage Keys:', debugInfo.localStorage.allKeys);
        console.groupEnd();
        
        // Show alert with summary
        alert('🔍 Debug Summary:\n\nToken: ' + (debugInfo.hasToken ? 'Present' : 'Missing') + '\nSwagger UI: ' + (debugInfo.swaggerReady ? 'Ready' : 'Not Ready') + '\nCheck console for details');
    }

    async login() {
        const email = document.getElementById('login-email').value;
        const password = document.getElementById('login-password').value;
        
        if (!email || !password) {
            alert('Please enter both email and password');
            return;
        }
        
        console.log('🔄 Starting login process...');

        try {
            const response = await fetch('/api/v1/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password })
            });

            if (response.ok) {
                const data = await response.json();
                this.token = data.access_token || data.token;
                
                if (!this.token) {
                    throw new Error('No token received from server');
                }
                
                // Save token
                localStorage.setItem('swagger_auth_token', this.token);
                console.log('💾 Token saved to localStorage');
                
                // Update UI status
                this.updateStatus();
                
                // Apply to Swagger UI
                this.applyTokenToSwagger();
                
                // Test the token immediately
                this.testToken();
                
                alert('✅ Login successful! Token applied and tested.');
            } else {
                const error = await response.json();
                console.error('❌ Login failed:', error);
                alert('❌ Login failed: ' + (error.error || 'Invalid credentials'));
            }
        } catch (err) {
            console.error('❌ Login error:', err);
            alert('❌ Login error: ' + err.message);
        }
    }
    
    // Test token with a simple API call
    async testToken() {
        if (!this.token) return;
        
        try {
            console.log('🧪 Testing token with /api/v1/profile...');
            const response = await fetch('/api/v1/profile', {
                method: 'GET',
                headers: {
                    'Authorization': 'Bearer ' + this.token,
                    'Content-Type': 'application/json'
                }
            });
            
            if (response.ok) {
                const data = await response.json();
                console.log('✅ Token test successful:', data);
            } else {
                const error = await response.json();
                console.error('❌ Token test failed:', response.status, error);
                
                if (response.status === 401) {
                    alert('⚠️ Token test failed - AUTH_HEADER_MISSING. Check browser console for details.');
                }
            }
        } catch (err) {
            console.error('❌ Token test error:', err);
        }
    }

    logout() {
        this.token = '';
        localStorage.removeItem('swagger_auth_token');
        this.updateStatus();
        
				// Clear Swagger UI authorization
				if (window.ui) {
					window.ui.preauthorizeApiKey('BearerAuth', '');
					console.log('🔓 Cleared Swagger UI authorization');
				}
        
        alert('🔓 Logged out successfully');
    }

    updateStatus() {
        const statusText = document.getElementById('status-text');
        if (this.token) {
            statusText.textContent = '✅ Authenticated';
            statusText.style.color = 'green';
        } else {
            statusText.textContent = '❌ Not authenticated';
            statusText.style.color = 'red';
        }
    }

    toggleUI() {
        const container = document.getElementById('swagger-auth-helper');
        container.style.display = container.style.display === 'none' ? 'block' : 'none';
    }

    // Intercept all requests to add Bearer token
    getToken() {
        return this.token;
    }
}

// Initialize auth helper
window.authHelper = new SwaggerAuthHelper();
`
}

// Enhanced Swagger HTML with better authentication support
func getEnhancedSwaggerHTML(docURL string) string {
	return GetEnhancedSwaggerHTML(docURL)
}

// GetEnhancedSwaggerHTML returns Enhanced Swagger HTML (exported version)
func GetEnhancedSwaggerHTML(docURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Sistema Akuntansi API - Enhanced Dynamic Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    html, body { height: 100%%; margin: 0; padding: 0; font-family: Arial, sans-serif; }
    #swagger-ui { height: 100%%; }
    
    /* Enhanced styling */
    .swagger-ui .topbar { background-color: #1b1b1b; }
    .swagger-ui .topbar .download-url-wrapper { display: none; }
    .swagger-ui .info .title { color: #3b4151; font-size: 36px; }
    
    /* Hide all info description content to prevent Quick Start from showing */
    .swagger-ui .info .description,
    .swagger-ui .info .description *,
    .info .description,
    .info .description *,
    [style*="background-color: green"],
    [style*="background: green"],
    [style*="background-color: #e8f5e8"],
    [style*="background: #e8f5e8"] {
      display: none !important;
      visibility: hidden !important;
      height: 0 !important;
      overflow: hidden !important;
    }
    
    /* Custom banners */
    .dynamic-banner { 
      background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); 
      color: white; 
      padding: 15px; 
      text-align: center; 
      font-weight: bold; 
      font-size: 14px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }
    
    .auth-banner { 
      background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%); 
      color: white; 
      padding: 10px 15px; 
      text-align: center; 
      font-size: 12px;
      margin-top: -1px;
    }
    
    /* Info banner removed - Quick Start content eliminated */
    
    /* Loading overlay */
    .loading-overlay {
      position: fixed;
      top: 0;
      left: 0;
      width: 100%%;
      height: 100%%;
      background: rgba(255,255,255,0.9);
      display: flex;
      justify-content: center;
      align-items: center;
      z-index: 10000;
      font-size: 18px;
      color: #333;
    }
  </style>
</head>
<body>
  <!-- Loading overlay -->
  <div class="loading-overlay" id="loading">
    <div>🚀 Loading Enhanced Swagger UI...</div>
  </div>

  <!-- Fail-safe: hide overlay even if other scripts fail -->
  <script>
    document.addEventListener('DOMContentLoaded', function() {
      try {
        var el = document.getElementById('loading');
        if (el) el.style.display = 'none';
      } catch (e) { /* ignore */ }
    });
    // Report uncaught errors into the UI to avoid blank screen
    window.addEventListener('error', function(ev) {
      try {
        var root = document.getElementById('swagger-ui');
        if (root && !root.innerHTML) {
          root.innerHTML = '<div style="padding:16px;color:red;">⚠️ Swagger UI runtime error: ' + (ev.error && ev.error.message ? ev.error.message : ev.message) + '</div>';
        }
      } catch (e) { /* ignore */ }
    });
  </script>
  
  <!-- Dynamic banners -->
  <div class="dynamic-banner">
    🚀 Enhanced Dynamic Swagger UI - Auto-fixing enabled | Authentication Helper Active
  </div>
  <div class="auth-banner">
    🔐 Use the Authentication Helper (top-right) to login and test protected endpoints
  </div>
  
  <!-- Main Swagger UI -->
  <div id="swagger-ui"></div>
  
  <!-- Authentication info moved to banner above -->

  <!-- Swagger UI Bundle -->
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  
  <!-- Authentication Helper -->
  <script>
    %s
  </script>
  
  <!-- Main Swagger UI initialization -->
  <script>
    window.onload = function() {
      // Hide loading overlay after a short delay
      setTimeout(() => {
        document.getElementById('loading').style.display = 'none';
      }, 1000);
      
      // Fetch swagger spec and initialize UI
      fetch('%s')
        .then(response => response.json())
        .then(spec => {
          window.ui = SwaggerUIBundle({
            url: '%s',
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
              SwaggerUIBundle.presets.apis,
              SwaggerUIBundle.presets.standalone
            ],
            plugins: [
              SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "BaseLayout",
            tryItOutEnabled: true,
            
			// Enhanced request interceptor with conflict prevention
			requestInterceptor: function(req) {
				console.log('🔍 [FIXED] Request interceptor:', req.method, req.url);
				
				// Initialize headers if not exist
				if (!req.headers) {
					req.headers = {};
				}
				
				// Check if Authorization header already exists (from preauthorizeApiKey)
				const existingAuth = req.headers['Authorization'] || req.headers['authorization'];
				
				if (existingAuth) {
					console.log('🔐 [FIXED] Authorization header already present:', existingAuth.substring(0, 30) + '...');
				} else {
					// Get token from localStorage as fallback
					const token = localStorage.getItem('swagger_auth_token') || '';
					
					if (token) {
						// Add Bearer token to requests that don't have it
						req.headers['Authorization'] = 'Bearer ' + token;
						console.log('🔐 [FIXED] Added Bearer token to request:', token.substring(0, 20) + '...');
					} else {
						console.log('⚠️ [FIXED] No token available - request may fail for protected endpoints');
					}
				}
              
              // Ensure proper content type for POST/PUT requests
              if (req.method === 'POST' || req.method === 'PUT' || req.method === 'PATCH') {
                if (!req.headers['Content-Type']) {
                  req.headers['Content-Type'] = 'application/json';
                }
              }
              
              // Log final headers for debugging
              console.log('📨 [FIXED] Final request headers:', Object.keys(req.headers));
              
              return req;
            },
            
            // Enhanced response interceptor
            responseInterceptor: function(res) {
              console.log('📨 Response interceptor:', res.status, res.url);
              
              // Handle common errors
              if (res.status === 401) {
                console.log('🔒 Unauthorized - please check your authentication');
                if (res.url.includes('/admin/')) {
                  console.log('💡 This is an admin endpoint - make sure you are logged in as admin');
                }
              } else if (res.status === 404 && res.url.includes('/admin/')) {
                console.log('🔧 404 for admin route - checking if /api/v1 prefix is needed...');
              } else if (res.status === 500) {
                console.log('💥 Server error - check server logs for details');
              }
              
              return res;
            },
            
            // Enhanced auto-authorization with retry mechanism
            onComplete: function() {
              console.log('✅ [FIXED] Swagger UI loaded successfully');
              
              // Notify auth helper that Swagger is ready
              if (window.authHelper) {
                window.authHelper.isSwaggerReady = true;
                console.log('✅ [FIXED] Notified authHelper that Swagger is ready');
              }
              
				// Auto-authorize if token exists with retry mechanism
				const attemptAutoAuth = (retries = 3) => {
					const token = localStorage.getItem('swagger_auth_token');
					if (token) {
						try {
							// Clear existing auth first
							window.ui.preauthorizeApiKey('BearerAuth', '');
							
							// Set new authorization
							window.ui.preauthorizeApiKey('BearerAuth', 'Bearer ' + token);
							console.log('🔐 [FIXED] Auto-authorized with saved token:', token.substring(0, 20) + '...');
							
							// Verify authorization was set
							setTimeout(() => {
								const authObj = window.ui.auth();
								if (authObj && authObj.BearerAuth) {
									console.log('✅ [FIXED] Authorization verified in UI');
								} else {
									console.log('⚠️ [FIXED] Authorization not verified, retrying...');
									if (retries > 0) {
										setTimeout(() => attemptAutoAuth(retries - 1), 500);
									}
								}
							}, 100);
							
						} catch (error) {
							console.error('❌ [FIXED] Auto-authorization error:', error);
							if (retries > 0) {
								setTimeout(() => attemptAutoAuth(retries - 1), 500);
							}
						}
					} else {
						console.log('⚠️ [FIXED] No saved token found for auto-authorization');
					}
				};
				
				// Start auto-authorization with slight delay
				setTimeout(() => attemptAutoAuth(), 100);
				
				// DOM aggressive cleanup disabled to prevent blank screen after a few seconds.
				// If you want to hide specific sections, target them precisely and avoid wildcard removals.
            }
          });
        })
        .catch(error => {
          console.error('❌ Failed to load Swagger spec:', error);
          document.getElementById('swagger-ui').innerHTML = 
            '<div style="padding: 20px; text-align: center; color: red;">' +
            '<h3>⚠️ Swagger Loading Error</h3>' +
            '<p>Failed to load API documentation. Please check server status.</p>' +
            '<p>Error: ' + error.message + '</p>' +
            '</div>';
        });
    };
  </script>
</body>
</html>`, getAuthenticationHelperJS(), docURL, docURL)
}

// getBasicSwaggerHTML returns a minimal Swagger UI page without custom helpers
func getBasicSwaggerHTML(docURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI - Basic</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>html,body,#swagger{height:100%%;margin:0;padding:0}</style>
</head>
<body>
  <div id="swagger"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      try {
        window.ui = SwaggerUIBundle({ url: '%s', dom_id: '#swagger' });
      } catch (e) {
        document.getElementById('swagger').innerHTML = '<div style="padding:16px;color:red;">⚠️ Failed to initialize Swagger UI: ' + e.message + '</div>';
      }
    };
  </script>
</body>
</html>`, docURL)
}

// Enhanced Swagger fixing rules with more comprehensive patterns
var enhancedSwaggerFixRules = []SwaggerFixRule{
	// API prefix fixes
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
		Pattern:     `"\/products\/([^"]+)"`,
		Replacement: `"/api/v1/products/$1"`,
		Description: "Fix missing /api/v1 prefix in product routes",
	},
	{
		Pattern:     `"\/contacts\/([^"]+)"`,
		Replacement: `"/api/v1/contacts/$1"`,
		Description: "Fix missing /api/v1 prefix in contact routes",
	},
	{
		Pattern:     `"\/accounts\/([^"]+)"`,
		Replacement: `"/api/v1/accounts/$1"`,
		Description: "Fix missing /api/v1 prefix in account routes",
	},
	{
		Pattern:     `"\/sales\/([^"]+)"`,
		Replacement: `"/api/v1/sales/$1"`,
		Description: "Fix missing /api/v1 prefix in sales routes",
	},
	{
		Pattern:     `"\/purchases\/([^"]+)"`,
		Replacement: `"/api/v1/purchases/$1"`,
		Description: "Fix missing /api/v1 prefix in purchase routes",
	},
	{
		Pattern:     `"\/assets\/([^"]+)"`,
		Replacement: `"/api/v1/assets/$1"`,
		Description: "Fix missing /api/v1 prefix in asset routes",
	},
	{
		Pattern:     `"\/inventory\/([^"]+)"`,
		Replacement: `"/api/v1/inventory/$1"`,
		Description: "Fix missing /api/v1 prefix in inventory routes",
	},
	{
		Pattern:     `"\/users\/([^"]+)"`,
		Replacement: `"/api/v1/users/$1"`,
		Description: "Fix missing /api/v1 prefix in user routes",
	},
	{
		Pattern:     `"\/settings\/([^"]+)"`,
		Replacement: `"/api/v1/settings/$1"`,
		Description: "Fix missing /api/v1 prefix in settings routes",
	},
	// Common typo fixes
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
	{
		Pattern:     `"recieve"`,
		Replacement: `"receive"`,
		Description: "Fix common typo: recieve -> receive",
	},
	{
		Pattern:     `"seperate"`,
		Replacement: `"separate"`,
		Description: "Fix common typo: seperate -> separate",
	},
}

// ValidateAndFixEnhancedSwagger performs comprehensive swagger validation and fixing
func ValidateAndFixEnhancedSwagger() (*EnhancedSwaggerConfig, error) {
	log.Println("🔍 Starting enhanced Swagger validation and fixing...")
	
	config := &EnhancedSwaggerConfig{
		Host:                getSwaggerHost(),
		BasePath:            "/api/v1",
		Schemes:             []string{"http"},
		Fixes:               []string{},
		Errors:              []string{},
		AuthenticationReady: true,
		SecurityDefinitions: map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type":        "apiKey",
				"name":        "Authorization",
				"in":          "header",
				"description": "Type 'Bearer' followed by a space and JWT token.",
			},
		},
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

	// Apply enhanced fixes
	originalContent := string(data)
	fixedContent := originalContent
	
	for _, rule := range enhancedSwaggerFixRules {
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

		log.Printf("✅ Applied %d enhanced fixes to swagger.json", len(config.Fixes))
		for _, fix := range config.Fixes {
			log.Printf("  - %s", fix)
		}
	} else {
		log.Println("✅ No Swagger fixes needed")
	}

	return config, nil
}

// GenerateEnhancedSwaggerSpec creates a comprehensive swagger spec with proper authentication
func GenerateEnhancedSwaggerSpec() map[string]interface{} {
	_ = LoadConfig() // Load config but don't use it for now
	
	spec := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"title":       "Sistema Akuntansi API - Enhanced",
			"description": "Comprehensive accounting system API with enhanced authentication support.",
			"version":     "1.0.0",
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
		"basePath": "/api/v1",
		"schemes":  []string{"http"},
		"securityDefinitions": map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type":        "apiKey",
				"name":        "Authorization",
				"in":          "header",
				"description": "JWT Bearer token. Use the Authentication Helper to login automatically.",
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
					"description": "Returns API health status (public endpoint)",
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
					"description": "Authenticate user and return JWT token. Default: admin@company.com / admin123",
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
										"description": "User email (try: admin@company.com)",
										"example":     "admin@company.com",
									},
									"password": map[string]interface{}{
										"type":        "string",
										"description": "User password (try: admin123)",
										"example":     "admin123",
									},
								},
								"required": []string{"email", "password"},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Login successful",
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"success": map[string]interface{}{
										"type": "boolean",
									},
									"message": map[string]interface{}{
										"type": "string",
									},
									"access_token": map[string]interface{}{
										"type": "string",
									},
									"token": map[string]interface{}{
										"type": "string",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Invalid credentials",
						},
					},
					"security": []map[string]interface{}{}, // Public endpoint
				},
			},
			"/api/v1/admin/check-cashbank-gl-links": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Admin", "Cash Bank"},
					"summary":     "Check cash bank GL account links status",
					"description": "Check which cash/bank accounts are missing GL account connections (Admin only)",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Cash bank GL links status",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized - Bearer token required",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin role required",
						},
					},
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
				},
			},
			"/api/v1/admin/balance-health/check": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Admin", "Balance Health"},
					"summary":     "Check balance health status",
					"description": "Check the health status of account balances (Admin only)",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Balance health status",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized - Bearer token required",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin role required",
						},
					},
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
				},
			},
			"/api/v1/admin/balance-health/auto-heal": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Admin", "Balance Health"},
					"summary":     "Auto-heal balance issues",
					"description": "Automatically heal detected balance issues (Admin only)",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Auto-heal completed successfully",
						},
						"206": map[string]interface{}{
							"description": "Auto-heal completed with warnings",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized - Bearer token required",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin role required",
						},
						"500": map[string]interface{}{
							"description": "Internal server error",
						},
					},
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
				},
			},
			"/api/v1/cash-bank/accounts": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Cash Bank"},
					"summary":     "Get all cash and bank accounts",
					"description": "Retrieve list of all cash and bank accounts (requires authentication)",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of cash and bank accounts",
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"data": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"id":         map[string]interface{}{"type": "integer"},
												"code":       map[string]interface{}{"type": "string"},
												"name":       map[string]interface{}{"type": "string"},
												"type":       map[string]interface{}{"type": "string"},
												"account_no": map[string]interface{}{"type": "string"},
												"currency":   map[string]interface{}{"type": "string"},
												"balance":    map[string]interface{}{"type": "number"},
											},
										},
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized - Bearer token required",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Insufficient permissions",
						},
					},
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
				},
			},
			"/api/v1/permissions/me": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Permissions"},
					"summary":     "Get current user permissions",
					"description": "Retrieve permissions for the authenticated user",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "User permissions",
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"user": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"id":       map[string]interface{}{"type": "integer"},
											"username": map[string]interface{}{"type": "string"},
											"email":    map[string]interface{}{"type": "string"},
											"role":     map[string]interface{}{"type": "string"},
										},
									},
									"permissions": map[string]interface{}{
										"type": "object",
										"additionalProperties": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"can_view":    map[string]interface{}{"type": "boolean"},
												"can_create":  map[string]interface{}{"type": "boolean"},
												"can_edit":    map[string]interface{}{"type": "boolean"},
												"can_delete":  map[string]interface{}{"type": "boolean"},
												"can_approve": map[string]interface{}{"type": "boolean"},
												"can_export":  map[string]interface{}{"type": "boolean"},
											},
										},
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized - Bearer token required",
						},
					},
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
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
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Error code",
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
					"success": map[string]interface{}{
						"type":        "boolean",
						"description": "Success status",
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

// SetupEnhancedSwaggerRoutes sets up enhanced swagger routes with comprehensive authentication support
func SetupEnhancedSwaggerRoutes(r *gin.Engine) {
	log.Println("🚀 Setting up enhanced Swagger routes with authentication support...")

	// Validate and fix swagger
	config, err := ValidateAndFixEnhancedSwagger()
	if err != nil {
		log.Printf("❌ Enhanced Swagger validation failed: %v", err)
		// Continue with fallback spec
	}

	// Print configuration info
	if len(config.Fixes) > 0 {
		log.Printf("✅ Applied %d enhanced Swagger fixes:", len(config.Fixes))
		for _, fix := range config.Fixes {
			log.Printf("  - %s", fix)
		}
	}

	if len(config.Errors) > 0 {
		log.Printf("⚠️ Enhanced Swagger errors detected:")
		for _, error := range config.Errors {
			log.Printf("  - %s", error)
		}
	}

	// Enhanced Swagger JSON endpoint
	r.GET("/openapi/enhanced-doc.json", func(c *gin.Context) {
		// Try to serve the actual swagger.json first, but prune it to public endpoints only
		swaggerPath := filepath.Join("docs", "swagger.json")
		if _, err := os.Stat(swaggerPath); err == nil {
			data, err := os.ReadFile(swaggerPath)
			if err == nil {
				var spec map[string]interface{}
				if err := json.Unmarshal(data, &spec); err == nil {
					// Filter spec to only include allowed public endpoints
					filtered := filterSwaggerSpecForPublic(spec)
					buf, merr := json.Marshal(filtered)
					if merr == nil {
						c.Header("Content-Type", "application/json")
						c.Header("Access-Control-Allow-Origin", "*")
						c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
						c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
						c.Data(200, "application/json", buf)
						return
					}
				}
			}
		}

		// Fallback to enhanced dynamic spec (also pruned for consistency)
		log.Println("📄 Serving enhanced dynamic fallback Swagger spec (public view)")
		spec := GenerateEnhancedSwaggerSpec()
		spec = filterSwaggerSpecForPublic(spec)
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.JSON(200, spec)
	})

	// Enhanced Swagger UI endpoints
	swaggerGroup := r.Group("/")
	swaggerGroup.Use(func(c *gin.Context) {
		// Enhanced CSP for Swagger UI with authentication support
		c.Header("Content-Security-Policy", "default-src 'self' https: data: blob:; script-src 'self' 'unsafe-inline' 'unsafe-eval' https:; style-src 'self' 'unsafe-inline' https:; img-src 'self' data: https:; font-src 'self' data: https:; connect-src 'self' data: blob: https:; frame-src 'self'")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Next()
	})

	// Enhanced Swagger UI routes
	swaggerGroup.GET("/enhanced-swagger", func(c *gin.Context) {
		c.Redirect(302, "/enhanced-swagger/index.html")
	})

	swaggerGroup.GET("/enhanced-swagger/index.html", func(c *gin.Context) {
		html := getEnhancedSwaggerHTML("/openapi/enhanced-doc.json")
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, html)
	})

	// Also serve on /swagger-enhanced for alternative access
	swaggerGroup.GET("/swagger-enhanced", func(c *gin.Context) {
		c.Redirect(302, "/enhanced-swagger/index.html")
	})

	// Basic minimal Swagger UI (fallback if enhanced UI causes issues)
	swaggerGroup.GET("/swagger/basic.html", func(c *gin.Context) {
		basic := getBasicSwaggerHTML("/openapi/enhanced-doc.json")
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, basic)
	})
	swaggerGroup.GET("/docs/basic.html", func(c *gin.Context) {
		basic := getBasicSwaggerHTML("/openapi/enhanced-doc.json")
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, basic)
	})
	
	// Serve diagnostic tool for AUTH_HEADER_MISSING debugging
	swaggerGroup.GET("/swagger-debug", func(c *gin.Context) {
		c.Redirect(302, "/swagger-debug.html")
	})
	
	swaggerGroup.GET("/swagger-debug.html", func(c *gin.Context) {
		// Read diagnostic tool HTML
		diagnosticHTML, err := os.ReadFile("swagger_debug.html")
		if err != nil {
			// Fallback if file doesn't exist
			c.String(404, "Diagnostic tool not found. Please ensure swagger_debug.html exists in the project root.")
			return
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, string(diagnosticHTML))
	})

	log.Printf("✅ Enhanced Dynamic Swagger UI available at: http://%s/enhanced-swagger/index.html", config.Host)
	log.Printf("🔐 Authentication Helper built-in - login with admin@company.com / admin123")
	log.Printf("🔍 Swagger Debug Tool available at: http://%s/swagger-debug.html", config.Host)
	log.Printf("🌟 Enhanced JSON endpoint: http://%s/openapi/enhanced-doc.json", config.Host)
}
