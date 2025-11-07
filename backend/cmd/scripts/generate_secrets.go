package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func main() {
	fmt.Println("==============================================")
	fmt.Println("       JWT Secret Generator for Production    ")
	fmt.Println("==============================================")
	
	// Generate secrets
	jwtAccessSecret, err := generateSecureSecret(64)
	if err != nil {
		fmt.Printf("Error generating JWT access secret: %v\n", err)
		return
	}
	
	jwtRefreshSecret, err := generateSecureSecret(64)
	if err != nil {
		fmt.Printf("Error generating JWT refresh secret: %v\n", err)
		return
	}
	
	csrfSecret, err := generateSecureSecret(32)
	if err != nil {
		fmt.Printf("Error generating CSRF secret: %v\n", err)
		return
	}
	
	// Generate session secret
	sessionSecret, err := generateSecureSecret(32)
	if err != nil {
		fmt.Printf("Error generating session secret: %v\n", err)
		return
	}
	
	// Create .env.production content
	envContent := fmt.Sprintf(`# ============================================
# PRODUCTION ENVIRONMENT CONFIGURATION
# Generated on: %s
# ============================================

# Database Configuration
DATABASE_URL=postgres://username:password@localhost/sistem_akuntansi?sslmode=require

# JWT Configuration - CRITICAL SECURITY
JWT_ACCESS_SECRET=%s
JWT_REFRESH_SECRET=%s
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
JWT_ISSUER=accounting-system-prod
JWT_AUDIENCE=accounting-app

# Session Configuration
SESSION_SECRET=%s
SESSION_MAX_AGE=28800 # 8 hours in seconds
SESSION_IDLE_TIMEOUT=1800 # 30 minutes in seconds

# CSRF Protection
CSRF_SECRET=%s
ENABLE_CSRF=true

# Security Configuration
ENVIRONMENT=production
ENABLE_HTTPS=true
ENABLE_RATE_LIMIT=true
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION=15m
ENABLE_SECURITY_HEADERS=true
ENABLE_MONITORING=true

# Server Configuration
SERVER_PORT=8080
ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
TRUSTED_PROXIES=

# Redis Cache (Optional but recommended)
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Monitoring & Alerting
ALERT_EMAIL=admin@yourdomain.com
ALERT_WEBHOOK_URL=
LOG_LEVEL=info
LOG_FILE=/var/log/accounting_system.log

# Rate Limiting Configuration
RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE=10
RATE_LIMIT_API_REQUESTS_PER_MINUTE=100

# Security Headers
HSTS_MAX_AGE=31536000
CSP_POLICY=default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';
X_FRAME_OPTIONS=DENY

# Token Configuration
ACCESS_TOKEN_COOKIE_NAME=access_token
REFRESH_TOKEN_COOKIE_NAME=refresh_token
COOKIE_DOMAIN=.yourdomain.com
COOKIE_SECURE=true
COOKIE_HTTP_ONLY=true
COOKIE_SAME_SITE=strict
`, 
		strings.Replace(fmt.Sprintf("%v", os.Args), " ", " ", -1),
		jwtAccessSecret,
		jwtRefreshSecret,
		sessionSecret,
		csrfSecret,
	)
	
	// Save to file
	dir, _ := os.Getwd()
	envPath := filepath.Join(dir, "..", ".env.production")
	
	err = os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		fmt.Printf("Error writing .env.production file: %v\n", err)
		return
	}
	
	// Also create a backup
	backupPath := filepath.Join(dir, "..", ".env.production.backup")
	err = os.WriteFile(backupPath, []byte(envContent), 0600)
	if err != nil {
		fmt.Printf("Warning: Could not create backup file: %v\n", err)
	}
	
	// Print summary
	fmt.Println("\n✅ Secrets generated successfully!")
	fmt.Println("================================================")
	fmt.Printf("📁 Files created:\n")
	fmt.Printf("   - .env.production\n")
	fmt.Printf("   - .env.production.backup\n")
	fmt.Println("\n🔐 Generated Secrets (Store these securely!):")
	fmt.Println("------------------------------------------------")
	fmt.Printf("JWT_ACCESS_SECRET  : %s...\n", jwtAccessSecret[:20])
	fmt.Printf("JWT_REFRESH_SECRET : %s...\n", jwtRefreshSecret[:20])
	fmt.Printf("SESSION_SECRET     : %s...\n", sessionSecret[:20])
	fmt.Printf("CSRF_SECRET        : %s...\n", csrfSecret[:20])
	fmt.Println("\n⚠️  IMPORTANT SECURITY NOTES:")
	fmt.Println("------------------------------------------------")
	fmt.Println("1. NEVER commit .env.production to version control")
	fmt.Println("2. Set appropriate file permissions (600)")
	fmt.Println("3. Rotate secrets regularly (every 30-90 days)")
	fmt.Println("4. Keep backup in secure location")
	fmt.Println("5. Update DATABASE_URL with actual credentials")
	fmt.Println("6. Update ALLOWED_ORIGINS with your domain")
	fmt.Println("================================================")
}
