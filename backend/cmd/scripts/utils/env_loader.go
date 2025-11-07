package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadEnvFromFile reads .env file and sets environment variables
func LoadEnvFromFile(envFile string) error {
	file, err := os.Open(envFile)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", envFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			   (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}
		
		// Set environment variable (don't override existing ones)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	
	return scanner.Err()
}

// LoadEnvWithFallback tries to load .env file, with fallback paths
func LoadEnvWithFallback() error {
	envPaths := []string{
		".env",
		"../.env",
		"../../.env",
	}
	
	var lastErr error
	for _, path := range envPaths {
		if err := LoadEnvFromFile(path); err == nil {
			fmt.Printf("✅ Loaded environment from: %s\n", path)
			return nil
		} else {
			lastErr = err
		}
	}
	
	return fmt.Errorf("failed to load .env from any path: %v", lastErr)
}

// GetDatabaseURL returns the DATABASE_URL from environment
func GetDatabaseURL() (string, error) {
	// Try to load .env if DATABASE_URL is not set
	if os.Getenv("DATABASE_URL") == "" {
		if err := LoadEnvWithFallback(); err != nil {
			return "", fmt.Errorf("DATABASE_URL not found in environment and failed to load .env: %v", err)
		}
	}
	
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return "", fmt.Errorf("DATABASE_URL environment variable is not set")
	}
	
	return databaseURL, nil
}

// PrintEnvInfo prints current environment information
func PrintEnvInfo() {
	fmt.Printf("🔧 ENVIRONMENT INFO:\n")
	fmt.Printf("   DATABASE_URL: %s\n", maskSensitiveURL(os.Getenv("DATABASE_URL")))
	fmt.Printf("   SERVER_PORT: %s\n", getEnvWithDefault("SERVER_PORT", "8080"))
	fmt.Printf("   ENVIRONMENT: %s\n", getEnvWithDefault("ENVIRONMENT", "development"))
	fmt.Printf("   SKIP_BALANCE_RESET: %s\n", getEnvWithDefault("SKIP_BALANCE_RESET", "false"))
	fmt.Printf("\n")
}

// maskSensitiveURL masks password in database URL for logging
func maskSensitiveURL(url string) string {
	if url == "" {
		return "NOT_SET"
	}
	
	// Simple masking: postgres://user:***@host/db
	parts := strings.Split(url, "@")
	if len(parts) != 2 {
		return "***MASKED***"
	}
	
	userParts := strings.Split(parts[0], ":")
	if len(userParts) >= 3 {
		return fmt.Sprintf("%s:%s:***@%s", userParts[0], userParts[1], parts[1])
	}
	
	return "***MASKED***"
}

// getEnvWithDefault returns environment variable value or default
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}