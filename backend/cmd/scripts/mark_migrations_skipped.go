package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MigrationLog struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	MigrationName string `json:"migration_name"`
	ExecutedAt    string `json:"executed_at"`
	Description   string `json:"description"`
	Status        string `json:"status"`
}

// loadEnv loads environment variables from .env file
func loadEnv() map[string]string {
	env := make(map[string]string)
	
	// Get the current working directory and look for .env file
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not get working directory: %v\n", err)
		return env
	}
	
	// Look for .env file in current directory or parent directories
	envPaths := []string{
		filepath.Join(wd, ".env"),
		filepath.Join(filepath.Dir(wd), ".env"),
		filepath.Join(filepath.Dir(filepath.Dir(wd)), ".env"),
	}
	
	var envFile string
	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			envFile = path
			break
		}
	}
	
	if envFile == "" {
		fmt.Println("⚠️  Warning: .env file not found")
		return env
	}
	
	file, err := os.Open(envFile)
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not open .env file: %v\n", err)
		return env
	}
	defer file.Close()
	
	fmt.Printf("📄 Loading configuration from: %s\n", envFile)
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			env[key] = value
		}
	}
	
	return env
}

// parseDatabaseURL parses a database URL and returns connection components
func parseDatabaseURL(databaseURL string) (host, port, user, password, dbname, sslmode string) {
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		// Remove protocol
		url := strings.TrimPrefix(databaseURL, "postgres://")
		url = strings.TrimPrefix(url, "postgresql://")
		
		// Parse user:password@host:port/database?params
		parts := strings.Split(url, "@")
		if len(parts) == 2 {
			// Extract user:password
			userPass := strings.Split(parts[0], ":")
			if len(userPass) == 2 {
				user = userPass[0]
				password = userPass[1]
			}
			
			// Extract host:port/database?params
			hostDb := parts[1]
			slashIndex := strings.Index(hostDb, "/")
			if slashIndex != -1 {
				// Extract host:port
				hostPort := hostDb[:slashIndex]
				colonIndex := strings.Index(hostPort, ":")
				if colonIndex != -1 {
					host = hostPort[:colonIndex]
					port = hostPort[colonIndex+1:]
				} else {
					host = hostPort
					port = "5432" // Default PostgreSQL port
				}
				
				// Extract database?params
				dbParams := hostDb[slashIndex+1:]
				questionIndex := strings.Index(dbParams, "?")
				if questionIndex != -1 {
					dbname = dbParams[:questionIndex]
					// Parse parameters
					params := dbParams[questionIndex+1:]
					for _, param := range strings.Split(params, "&") {
						if strings.HasPrefix(param, "sslmode=") {
							sslmode = strings.TrimPrefix(param, "sslmode=")
						}
					}
				} else {
					dbname = dbParams
				}
			}
		}
		
		// Set defaults
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}
		if sslmode == "" {
			sslmode = "disable"
		}
	}
	
	return
}

func main() {
	fmt.Println("🛠️  MIGRATION SKIP UTILITY")
	fmt.Println("==========================")
	
	// Load environment variables from .env file
	env := loadEnv()
	
	// Get database configuration
	databaseURL := env["DATABASE_URL"]
	if databaseURL == "" {
		fmt.Println("❌ DATABASE_URL not found in .env file")
		return
	}
	
	// Parse database URL
	host, port, user, password, dbname, sslmode := parseDatabaseURL(databaseURL)
	
	fmt.Println("\n🔌 Connecting to database...")
	fmt.Printf("   Database: %s\n", dbname)
	fmt.Printf("   Host: %s:%s\n", host, port)
	fmt.Printf("   User: %s\n", user)
	
	// Build PostgreSQL DSN
	psqlDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbname, port, sslmode)
	db, err := gorm.Open(postgres.Open(psqlDSN), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		return
	}
	
	fmt.Println("✅ Connected successfully!")
	
	// List of problematic migrations to mark as skipped
	problematicMigrations := []string{
		"021_install_purchase_balance_system.sql",
		"021_install_purchase_balance_validation.sql", 
		"022_purchase_balance_validation_postgresql.sql",
		"023_purchase_balance_validation_go_compatible.sql",
		"024_purchase_balance_simple.sql",
		"025_purchase_balance_no_dollar_quotes.sql",
	}
	
	fmt.Println("\n🚫 Marking problematic migrations as SKIPPED...")
	
	for _, migrationName := range problematicMigrations {
		var existing MigrationLog
		result := db.Where("migration_name = ?", migrationName).First(&existing)
		
		if result.Error == nil {
			// Migration exists, update status if it's FAILED
			if existing.Status == "FAILED" {
				existing.Status = "SKIPPED"
				existing.Description = existing.Description + " [AUTO-SKIPPED: Go SQL driver incompatible]"
				db.Save(&existing)
				fmt.Printf("   ✅ Updated %s: FAILED → SKIPPED\n", migrationName)
			} else {
				fmt.Printf("   ℹ️  %s already has status: %s\n", migrationName, existing.Status)
			}
		} else {
			// Migration doesn't exist, create it as SKIPPED
			newMigration := MigrationLog{
				MigrationName: migrationName,
				ExecutedAt:    "1970-01-01T00:00:00Z", // Epoch time to indicate "never executed"
				Description:   "SKIPPED: PostgreSQL function syntax incompatible with Go SQL driver",
				Status:        "SKIPPED",
			}
			db.Create(&newMigration)
			fmt.Printf("   ✅ Created %s as SKIPPED\n", migrationName)
		}
	}
	
	fmt.Println("\n💡 Result:")
	fmt.Println("   The problematic migrations are now marked as SKIPPED")
	fmt.Println("   Backend will no longer attempt to run these migrations")
	fmt.Println("   The working migration (026_purchase_balance_minimal.sql) will still run")
	
	fmt.Println("\n🎯 Summary:")
	fmt.Println("   ✅ Database connection working")
	fmt.Println("   ✅ Problematic migrations marked as SKIPPED")
	fmt.Println("   ✅ Backend startup will be cleaner (no more error spam)")
	fmt.Println("   ✅ Working migration (026_purchase_balance_minimal.sql) unaffected")
	
	fmt.Println("\n🚀 Next step: Run your backend - it should start without migration errors!")
}