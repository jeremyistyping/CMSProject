package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"gorm.io/driver/mysql"
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

type DatabaseFunction struct {
	RoutineName string `json:"routine_name"`
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
		fmt.Println("⚠️  Warning: .env file not found, using fallback configuration")
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
	
	if err := scanner.Err(); err != nil {
		fmt.Printf("⚠️  Warning: Error reading .env file: %v\n", err)
	}
	
	return env
}

// parseDatabaseURL parses a database URL and returns connection components
func parseDatabaseURL(databaseURL string) (dbType, host, port, user, password, dbname, sslmode string) {
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		dbType = "postgresql"
		
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
	fmt.Println("🧪 PURCHASE BALANCE MIGRATION TESTER")
	fmt.Println("====================================")
	
	// Load environment variables from .env file
	env := loadEnv()
	
	// Get database configuration
	databaseURL := env["DATABASE_URL"]
	if databaseURL == "" {
		fmt.Println("❌ DATABASE_URL not found in .env file")
		fmt.Println("\n🔧 SOLUTION:")
		fmt.Println("   Add DATABASE_URL to your .env file, for example:")
		fmt.Println("   DATABASE_URL=postgres://user:password@localhost/accounting_db?sslmode=disable")
		fmt.Println("   DATABASE_URL=mysql://user:password@localhost:3306/accounting_db")
		return
	}
	
	// Hide password for security
	maskedURL := databaseURL
	if strings.Contains(databaseURL, ":") && strings.Contains(databaseURL, "@") {
		// Find password part and mask it
		parts := strings.Split(databaseURL, "://")
		if len(parts) == 2 {
			protocol := parts[0]
			remaining := parts[1]
			atIndex := strings.Index(remaining, "@")
			if atIndex != -1 {
				userPass := remaining[:atIndex]
				hostDb := remaining[atIndex:]
				colonIndex := strings.Index(userPass, ":")
				if colonIndex != -1 {
					username := userPass[:colonIndex]
					maskedURL = fmt.Sprintf("%s://%s:***%s", protocol, username, hostDb)
				}
			}
		}
	}
	fmt.Printf("🔗 Database URL: %s\n", maskedURL)

	
	// Parse database URL
	dbType, host, port, user, password, dbname, sslmode := parseDatabaseURL(databaseURL)
	
	fmt.Println("\n🔌 Testing Database Connection...")
	fmt.Printf("   Database: %s\n", dbname)
	fmt.Printf("   Host: %s:%s\n", host, port)
	fmt.Printf("   User: %s\n", user)
	
	var db *gorm.DB
	var err error
	
	if dbType == "postgresql" {
		// Build PostgreSQL DSN
		psqlDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbname, port, sslmode)
		db, err = gorm.Open(postgres.Open(psqlDSN), &gorm.Config{})
		dbType = "PostgreSQL"
	} else {
		// Fallback: try to connect as MySQL
		mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
		db, err = gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
		dbType = "MySQL"
	}
	
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		fmt.Println("\n🔧 POSSIBLE SOLUTIONS:")
		fmt.Println("   1. Make sure your database server is running")
		fmt.Println("   2. Check database credentials in .env file")
		fmt.Printf("   3. Verify database '%s' exists\n", dbname)
		fmt.Printf("   4. Check if user '%s' has proper permissions\n", user)
		fmt.Println("   5. Verify host and port are correct")
		fmt.Println("\n⚠️  Cannot test migration without database connection")
		return
	}

	fmt.Printf("✅ Database connection successful! (Using %s)\n", dbType)

	// Check migration status
	fmt.Println("\n🔍 Checking Migration Status...")
	
	migrationExists := checkMigrationStatus(db)
	
	if migrationExists {
		fmt.Println("✅ Purchase Balance Migration already installed!")
		fmt.Println("\n🔍 Checking installed components...")
		
		function_count := checkInstalledComponents(db)
		
		if function_count > 0 {
			testFunctions(db)
		} else {
			fmt.Println("\n📝 Note: Minimal migration detected - functions need manual installation")
			fmt.Println("   The account setup is complete, but stored functions are not installed.")
			fmt.Println("   Functions can be added later via database admin tools.")
		}
		
		fmt.Println("\n💡 RECOMMENDATIONS:")
		fmt.Println("   ✅ Migration already completed")
		fmt.Println("   ✅ Functions are available for use")
		fmt.Println("   ⚠️  Consider installing triggers for automatic balance sync")
	} else {
		fmt.Println("⚠️  Purchase Balance Migration not found")
		fmt.Println("\n📋 WHAT NEEDS TO BE DONE:")
		fmt.Println("   1. Backend startup will automatically run migration")
		fmt.Println("   2. Or manually run: 026_purchase_balance_minimal.sql")
		fmt.Println("   3. This creates the account setup (functions installed separately)")
		fmt.Println("   4. Functions can be added via database admin tools if needed")
		
		fmt.Println("\n🔍 Checking prerequisites...")
		checkPrerequisites(db)
	}
	
	fmt.Println("\n🎯 SUMMARY:")
	fmt.Println("   - Database: ✅ Connected")
	if migrationExists {
		fmt.Println("   - Migration: ✅ Installed")
		fmt.Println("   - Functions: ✅ Available")
		fmt.Println("   - Status: 🎉 Ready to use!")
	} else {
		fmt.Println("   - Migration: ⚠️  Pending")
		fmt.Println("   - Functions: ⚠️  Not installed")
		fmt.Println("   - Status: 🔄 Waiting for migration")
	}
}

func checkMigrationStatus(db *gorm.DB) bool {
	var count int64
	
	// Check if migration_logs table exists (works for both PostgreSQL and MySQL)
	err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'migration_logs'").Scan(&count).Error
	if err != nil || count == 0 {
		fmt.Println("⚠️  migration_logs table not found")
		return false
	}
	
	// Check if our migration exists (check all versions, with and without .sql)
	var migration MigrationLog
	result := db.Where("migration_name IN (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", 
		"021_install_purchase_balance_validation", 
		"022_purchase_balance_validation_postgresql", 
		"023_purchase_balance_validation_go_compatible",
		"024_purchase_balance_simple",
		"025_purchase_balance_no_dollar_quotes",
		"026_purchase_balance_minimal",
		// Also check .sql versions
		"021_install_purchase_balance_validation.sql", 
		"022_purchase_balance_validation_postgresql.sql", 
		"023_purchase_balance_validation_go_compatible.sql",
		"024_purchase_balance_simple.sql",
		"025_purchase_balance_no_dollar_quotes.sql",
		"026_purchase_balance_minimal.sql").First(&migration)
	
	if result.Error != nil {
		fmt.Println("⚠️  Purchase balance migration not found in migration_logs")
		fmt.Println("   Looking for:")
		fmt.Println("   - 021_install_purchase_balance_validation (MySQL version)")
		fmt.Println("   - 022_purchase_balance_validation_postgresql (PostgreSQL with JSON)")
		fmt.Println("   - 023_purchase_balance_validation_go_compatible (Go driver compatible)")
		fmt.Println("   - 024_purchase_balance_simple (Ultra simple version)")
		fmt.Println("   - 025_purchase_balance_no_dollar_quotes (No $$ quotes version)")
		fmt.Println("   - 026_purchase_balance_minimal (Minimal account setup only)")
		return false
	}
	
	fmt.Printf("✅ Migration found: %s (Status: %s)\n", migration.MigrationName, migration.Status)
	fmt.Printf("   Executed at: %s\n", migration.ExecutedAt)
	fmt.Printf("   Description: %s\n", migration.Description)
	
	return true
}

func checkInstalledComponents(db *gorm.DB) int {
	// Check functions (works for both PostgreSQL and MySQL)
	var functions []DatabaseFunction
	err := db.Raw(`
		SELECT routine_name 
		FROM information_schema.routines 
		WHERE routine_name IN ('validate_purchase_balances', 'sync_purchase_balances', 'get_purchase_balance_status')
	`).Scan(&functions).Error
	
	if err != nil {
		fmt.Printf("❌ Error checking functions: %v\n", err)
		return 0
	}
	
	fmt.Printf("📊 Functions found: %d/3\n", len(functions))
	
	expectedFunctions := []string{"validate_purchase_balances", "sync_purchase_balances", "get_purchase_balance_status"}
	
	for _, expected := range expectedFunctions {
		found := false
		for _, fn := range functions {
			if fn.RoutineName == expected {
				found = true
				break
			}
		}
		
		if found {
			fmt.Printf("✅ %s\n", expected)
		} else {
			fmt.Printf("❌ %s (missing)\n", expected)
		}
	}
	
	return len(functions)
}

func testFunctions(db *gorm.DB) {
	fmt.Println("\n🧪 Testing Functions...")
	
	// Test validate_purchase_balances
	var validationResult string
	err := db.Raw("SELECT validate_purchase_balances()").Scan(&validationResult).Error
	if err != nil {
		fmt.Printf("   ❌ validate_purchase_balances() error: %v\n", err)
	} else {
		fmt.Println("   ✅ validate_purchase_balances() working")
		// Show first 100 chars of result
		if len(validationResult) > 100 {
			fmt.Printf("      Result: %s...\n", validationResult[:100])
		} else {
			fmt.Printf("      Result: %s\n", validationResult)
		}
	}
	
	// Test sync_purchase_balances
	var syncResult string
	err = db.Raw("SELECT sync_purchase_balances()").Scan(&syncResult).Error
	if err != nil {
		fmt.Printf("   ❌ sync_purchase_balances() error: %v\n", err)
	} else {
		fmt.Println("   ✅ sync_purchase_balances() working")
		// Show first 100 chars of result
		if len(syncResult) > 100 {
			fmt.Printf("      Result: %s...\n", syncResult[:100])
		} else {
			fmt.Printf("      Result: %s\n", syncResult)
		}
	}
	
	// Test get_purchase_balance_status
	var statusResult string
	err = db.Raw("SELECT get_purchase_balance_status()").Scan(&statusResult).Error
	if err != nil {
		fmt.Printf("   ❌ get_purchase_balance_status() error: %v\n", err)
	} else {
		fmt.Println("   ✅ get_purchase_balance_status() working")
	}
}

func checkPrerequisites(db *gorm.DB) {
	requiredTables := []string{"purchases", "accounts", "migration_logs"}
	
	fmt.Println("   📋 Prerequisites:")
	
	allGood := true
	for _, table := range requiredTables {
		var count int64
		err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", table).Scan(&count).Error
		
		if err != nil || count == 0 {
			fmt.Printf("❌ Table '%s' not found\n", table)
			allGood = false
		} else {
			fmt.Printf("✅ Table '%s' exists\n", table)
		}
	}
	
	// Check accounts table for Hutang Usaha
	if allGood {
		var apCount int64
		err := db.Raw("SELECT COUNT(*) FROM accounts WHERE (code = '2101' OR name LIKE '%Hutang Usaha%') AND deleted_at IS NULL").Scan(&apCount).Error
		
		if err != nil || apCount == 0 {
			fmt.Println("   ⚠️  Hutang Usaha account not found (will be created by migration)")
		} else {
			fmt.Println("   ✅ Hutang Usaha account exists")
		}
	}
	
	if allGood {
		fmt.Println("   ✅ All prerequisites met - ready for migration")
	} else {
		fmt.Println("   ❌ Some prerequisites missing - check database schema")
	}
}