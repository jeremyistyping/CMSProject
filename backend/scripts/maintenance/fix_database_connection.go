package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"database/sql"
	_ "github.com/lib/pq"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func main() {
	fmt.Printf("%s=== DATABASE CONNECTION FIX UTILITY ===%s\n\n", ColorCyan, ColorReset)
	
	// Read current configuration
	currentConfig := readEnvFile()
	currentDatabaseURL := currentConfig["DATABASE_URL"]
	
	if currentDatabaseURL == "" {
		currentDatabaseURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}
	
	fmt.Printf("%sCurrent Database Configuration:%s\n", ColorYellow, ColorReset)
	fmt.Printf("  %s%s%s\n\n", ColorWhite, currentDatabaseURL, ColorReset)
	
	// Test current connection
	fmt.Printf("%sTesting current database connection...%s\n", ColorBlue, ColorReset)
	err := testDatabaseConnection(currentDatabaseURL)
	
	if err == nil {
		fmt.Printf("%s✅ Database connection is working fine!%s\n", ColorGreen, ColorReset)
		fmt.Printf("The reset script should work now. Try running:\n")
		fmt.Printf("  %sgo run reset_transaction_data_gorm.go%s\n", ColorCyan, ColorReset)
		return
	}
	
	fmt.Printf("%s❌ Database connection failed: %v%s\n\n", ColorRed, err, ColorReset)
	
	// Show solutions
	showSolutions(currentDatabaseURL, err)
}

func readEnvFile() map[string]string {
	envFile := "../../.env"
	config := make(map[string]string)
	
	file, err := os.Open(envFile)
	if err != nil {
		return config
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			config[parts[0]] = parts[1]
		}
	}
	
	return config
}

func testDatabaseConnection(databaseURL string) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()
	
	// Set connection timeout
	db.SetConnMaxLifetime(5 * time.Second)
	
	// Test the connection
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}
	
	return nil
}

func showSolutions(currentURL string, connectionError error) {
	fmt.Printf("%s=== SOLUTION OPTIONS ===%s\n\n", ColorPurple, ColorReset)
	
	errorStr := strings.ToLower(connectionError.Error())
	
	if strings.Contains(errorStr, "password authentication failed") {
		fmt.Printf("%s🔐 ISSUE: Password Authentication Failed%s\n\n", ColorRed, ColorReset)
		showPasswordSolutions()
	} else if strings.Contains(errorStr, "connection refused") || strings.Contains(errorStr, "no such host") {
		fmt.Printf("%s🚫 ISSUE: PostgreSQL Server Not Running or Not Found%s\n\n", ColorRed, ColorReset)
		showServerSolutions()
	} else if strings.Contains(errorStr, "does not exist") {
		fmt.Printf("%s🗄️ ISSUE: Database Does Not Exist%s\n\n", ColorRed, ColorReset)
		showDatabaseSolutions()
	} else {
		fmt.Printf("%s❓ ISSUE: General Connection Problem%s\n\n", ColorRed, ColorReset)
		showGeneralSolutions()
	}
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	askForAction(currentURL)
}

func showPasswordSolutions() {
	fmt.Printf("%sOption 1: Reset PostgreSQL Password%s\n", ColorGreen, ColorReset)
	fmt.Println("  1. Open Command Prompt as Administrator")
	fmt.Println("  2. Stop PostgreSQL service:")
	fmt.Println("     net stop postgresql-x64-15  (or your version)")
	fmt.Println("  3. Start PostgreSQL in single user mode")
	fmt.Println("  4. Reset password for 'postgres' user")
	fmt.Println("")
	
	fmt.Printf("%sOption 2: Use Different Password%s\n", ColorGreen, ColorReset)
	fmt.Println("  1. Find your actual PostgreSQL password")
	fmt.Println("  2. Update the .env file with correct password")
	fmt.Println("  3. Test connection again")
	fmt.Println("")
	
	fmt.Printf("%sOption 3: Create New User%s\n", ColorGreen, ColorReset)
	fmt.Println("  1. Login with existing admin account")
	fmt.Println("  2. Create new user for this application")
	fmt.Println("  3. Grant necessary permissions")
	fmt.Println("")
}

func showServerSolutions() {
	fmt.Printf("%sOption 1: Install PostgreSQL%s\n", ColorGreen, ColorReset)
	fmt.Println("  1. Download PostgreSQL from https://www.postgresql.org/download/")
	fmt.Println("  2. Install with default settings")
	fmt.Println("  3. Remember the password you set during installation")
	fmt.Println("")
	
	fmt.Printf("%sOption 2: Start PostgreSQL Service%s\n", ColorGreen, ColorReset)
	fmt.Println("  1. Open Services (services.msc)")
	fmt.Println("  2. Find PostgreSQL service")
	fmt.Println("  3. Start the service")
	fmt.Println("")
	
	if runtime.GOOS == "windows" {
		fmt.Printf("%sOption 3: Use Windows Commands%s\n", ColorGreen, ColorReset)
		fmt.Println("  Run as Administrator:")
		fmt.Println("  net start postgresql-x64-15")
		fmt.Println("")
	}
}

func showDatabaseSolutions() {
	fmt.Printf("%sOption 1: Create Database via psql%s\n", ColorGreen, ColorReset)
	fmt.Println("  1. Open command prompt")
	fmt.Println("  2. Run: psql -U postgres")
	fmt.Println("  3. Enter: CREATE DATABASE sistem_akuntansi;")
	fmt.Println("  4. Enter: \\q to exit")
	fmt.Println("")
	
	fmt.Printf("%sOption 2: Create Database via pgAdmin%s\n", ColorGreen, ColorReset)
	fmt.Println("  1. Open pgAdmin")
	fmt.Println("  2. Connect to PostgreSQL server")
	fmt.Println("  3. Right-click Databases -> Create -> Database")
	fmt.Println("  4. Name: sistem_akuntansi")
	fmt.Println("")
}

func showGeneralSolutions() {
	fmt.Printf("%sGeneral Troubleshooting Steps:%s\n", ColorGreen, ColorReset)
	fmt.Println("  1. Check if PostgreSQL is installed")
	fmt.Println("  2. Verify PostgreSQL service is running")
	fmt.Println("  3. Check firewall settings")
	fmt.Println("  4. Verify connection parameters (host, port, database)")
	fmt.Println("  5. Check PostgreSQL logs for detailed errors")
	fmt.Println("")
}

func askForAction(currentURL string) {
	fmt.Printf("%sWhat would you like to do?%s\n", ColorYellow, ColorReset)
	fmt.Println("1. 🔧 Try to fix PostgreSQL password automatically")
	fmt.Println("2. ✏️  Update .env file with new database URL")
	fmt.Println("3. 🧪 Test a different database URL")
	fmt.Println("4. 💾 Create the database 'sistem_akuntansi'")
	fmt.Println("5. 🔍 Show detailed PostgreSQL status")
	fmt.Println("6. ❌ Exit")
	fmt.Print("\nEnter your choice [1-6]: ")
	
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	switch choice {
	case "1":
		tryFixPassword()
	case "2":
		updateEnvFile()
	case "3":
		testCustomURL()
	case "4":
		createDatabase()
	case "5":
		showPostgreSQLStatus()
	case "6":
		fmt.Printf("%sGoodbye!%s\n", ColorCyan, ColorReset)
		return
	default:
		fmt.Printf("%sInvalid choice. Please run the program again.%s\n", ColorRed, ColorReset)
	}
}

func tryFixPassword() {
	fmt.Printf("%s🔧 Attempting to fix PostgreSQL password...%s\n", ColorYellow, ColorReset)
	
	fmt.Printf("%sThis will attempt to reset the 'postgres' user password to 'postgres'%s\n", ColorRed, ColorReset)
	fmt.Print("Continue? [y/N]: ")
	
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	
	if confirm != "y" && confirm != "yes" {
		fmt.Println("Operation cancelled.")
		return
	}
	
	// Try different approaches based on OS
	if runtime.GOOS == "windows" {
		tryWindowsPasswordReset()
	} else {
		tryUnixPasswordReset()
	}
}

func tryWindowsPasswordReset() {
	fmt.Printf("%sStep 1: Stopping PostgreSQL service...%s\n", ColorBlue, ColorReset)
	
	// Try to stop PostgreSQL service
	services := []string{"postgresql-x64-15", "postgresql-x64-14", "postgresql-x64-13", "postgresql"}
	for _, service := range services {
		cmd := exec.Command("net", "stop", service)
		if err := cmd.Run(); err == nil {
			fmt.Printf("%s✅ Stopped service: %s%s\n", ColorGreen, service, ColorReset)
			break
		}
	}
	
	fmt.Printf("%s⚠️  Manual steps required:%s\n", ColorYellow, ColorReset)
	fmt.Println("1. Open Command Prompt as Administrator")
	fmt.Println("2. Navigate to PostgreSQL bin directory (usually C:\\Program Files\\PostgreSQL\\15\\bin)")
	fmt.Println("3. Run: pg_ctl -D \"C:\\Program Files\\PostgreSQL\\15\\data\" -l logfile start")
	fmt.Println("4. Run: psql -U postgres")
	fmt.Println("5. Execute: ALTER USER postgres PASSWORD 'postgres';")
	fmt.Println("6. Restart PostgreSQL service")
	
	fmt.Print("\nPress Enter after completing these steps...")
	bufio.NewReader(os.Stdin).ReadString('\n')
	
	// Test the connection again
	testDatabaseConnection("postgres://postgres:postgres@localhost/sistema_akuntansi?sslmode=disable")
}

func tryUnixPasswordReset() {
	fmt.Println("Unix/Linux password reset steps:")
	fmt.Println("1. sudo -u postgres psql")
	fmt.Println("2. ALTER USER postgres PASSWORD 'postgres';")
	fmt.Println("3. \\q")
}

func updateEnvFile() {
	fmt.Printf("%s✏️  Update .env file with new database configuration%s\n", ColorYellow, ColorReset)
	
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Print("Enter database host [localhost]: ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}
	
	fmt.Print("Enter database port [5432]: ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)
	if port == "" {
		port = "5432"
	}
	
	fmt.Print("Enter database username [postgres]: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "postgres"
	}
	
	fmt.Print("Enter database password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	
	fmt.Print("Enter database name [sistem_akuntansi]: ")
	dbname, _ := reader.ReadString('\n')
	dbname = strings.TrimSpace(dbname)
	if dbname == "" {
		dbname = "sistem_akuntansi"
	}
	
	newURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, dbname)
	
	// Test the new URL first
	fmt.Printf("%s🧪 Testing new configuration...%s\n", ColorBlue, ColorReset)
	if err := testDatabaseConnection(newURL); err != nil {
		fmt.Printf("%s❌ Connection test failed: %v%s\n", ColorRed, err, ColorReset)
		fmt.Print("Do you still want to save this configuration? [y/N]: ")
		confirm, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
			fmt.Println("Configuration not saved.")
			return
		}
	} else {
		fmt.Printf("%s✅ Connection successful!%s\n", ColorGreen, ColorReset)
	}
	
	// Update the .env file
	if err := updateEnvDatabase(newURL); err != nil {
		fmt.Printf("%s❌ Failed to update .env file: %v%s\n", ColorRed, err, ColorReset)
		return
	}
	
	fmt.Printf("%s✅ .env file updated successfully!%s\n", ColorGreen, ColorReset)
	fmt.Printf("%sNew DATABASE_URL: %s%s\n", ColorCyan, newURL, ColorReset)
}

func testCustomURL() {
	fmt.Printf("%s🧪 Test a custom database URL%s\n", ColorYellow, ColorReset)
	
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter database URL: ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)
	
	if url == "" {
		fmt.Println("No URL provided.")
		return
	}
	
	fmt.Printf("%sTesting connection...%s\n", ColorBlue, ColorReset)
	if err := testDatabaseConnection(url); err != nil {
		fmt.Printf("%s❌ Connection failed: %v%s\n", ColorRed, err, ColorReset)
	} else {
		fmt.Printf("%s✅ Connection successful!%s\n", ColorGreen, ColorReset)
		
		fmt.Print("Do you want to save this URL to .env file? [y/N]: ")
		confirm, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(confirm)) == "y" {
			updateEnvDatabase(url)
		}
	}
}

func createDatabase() {
	fmt.Printf("%s💾 Create database 'sistem_akuntansi'%s\n", ColorYellow, ColorReset)
	
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter admin username [postgres]: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "postgres"
	}
	
	fmt.Print("Enter admin password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	
	adminURL := fmt.Sprintf("postgres://%s:%s@localhost/postgres?sslmode=disable", username, password)
	
	// Test admin connection first
	db, err := sql.Open("postgres", adminURL)
	if err != nil {
		fmt.Printf("%s❌ Failed to connect as admin: %v%s\n", ColorRed, err, ColorReset)
		return
	}
	defer db.Close()
	
	if err = db.Ping(); err != nil {
		fmt.Printf("%s❌ Failed to ping admin database: %v%s\n", ColorRed, err, ColorReset)
		return
	}
	
	// Create the database
	_, err = db.Exec("CREATE DATABASE sistem_akuntansi")
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Printf("%s✅ Database 'sistem_akuntansi' already exists%s\n", ColorGreen, ColorReset)
		} else {
			fmt.Printf("%s❌ Failed to create database: %v%s\n", ColorRed, err, ColorReset)
			return
		}
	} else {
		fmt.Printf("%s✅ Database 'sistem_akuntansi' created successfully!%s\n", ColorGreen, ColorReset)
	}
	
	// Test the new database connection
	newURL := fmt.Sprintf("postgres://%s:%s@localhost/sistem_akuntansi?sslmode=disable", username, password)
	if err := testDatabaseConnection(newURL); err != nil {
		fmt.Printf("%s❌ Failed to connect to new database: %v%s\n", ColorRed, err, ColorReset)
	} else {
		fmt.Printf("%s✅ New database connection successful!%s\n", ColorGreen, ColorReset)
		
		fmt.Print("Do you want to update .env with this configuration? [y/N]: ")
		confirm, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(confirm)) == "y" {
			updateEnvDatabase(newURL)
		}
	}
}

func showPostgreSQLStatus() {
	fmt.Printf("%s🔍 PostgreSQL Status Information%s\n", ColorYellow, ColorReset)
	fmt.Println(strings.Repeat("=", 50))
	
	// Check if PostgreSQL is installed
	fmt.Printf("%sChecking PostgreSQL installation...%s\n", ColorBlue, ColorReset)
	
	// Try to find psql
	cmd := exec.Command("psql", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s❌ psql not found in PATH%s\n", ColorRed, ColorReset)
		fmt.Println("PostgreSQL may not be installed or not in PATH")
	} else {
		fmt.Printf("%s✅ PostgreSQL found: %s%s", ColorGreen, string(output), ColorReset)
	}
	
	// Check service status (Windows)
	if runtime.GOOS == "windows" {
		fmt.Printf("%sChecking PostgreSQL service status...%s\n", ColorBlue, ColorReset)
		checkWindowsService()
	}
	
	// Try to connect with common passwords
	fmt.Printf("%sTrying common password combinations...%s\n", ColorBlue, ColorReset)
	commonPasswords := []string{"postgres", "password", "admin", "123456", ""}
	
	for _, pwd := range commonPasswords {
		testURL := fmt.Sprintf("postgres://postgres:%s@localhost/postgres?sslmode=disable", pwd)
		if err := testDatabaseConnection(testURL); err == nil {
			fmt.Printf("%s✅ Connection successful with password: '%s'%s\n", ColorGreen, pwd, ColorReset)
			
			// Check if sistem_akuntansi database exists
			checkDatabaseExists(testURL)
			return
		}
	}
	
	fmt.Printf("%s❌ No successful connections with common passwords%s\n", ColorRed, ColorReset)
}

func checkWindowsService() {
	services := []string{"postgresql-x64-15", "postgresql-x64-14", "postgresql-x64-13", "postgresql"}
	
	for _, service := range services {
		cmd := exec.Command("sc", "query", service)
		output, err := cmd.CombinedOutput()
		if err == nil {
			fmt.Printf("%sService %s:%s\n", ColorCyan, service, ColorReset)
			if strings.Contains(string(output), "RUNNING") {
				fmt.Printf("%s  ✅ Status: RUNNING%s\n", ColorGreen, ColorReset)
			} else if strings.Contains(string(output), "STOPPED") {
				fmt.Printf("%s  ❌ Status: STOPPED%s\n", ColorRed, ColorReset)
				fmt.Printf("    Try: %snet start %s%s\n", ColorCyan, service, ColorReset)
			} else {
				fmt.Printf("  Status: %s\n", strings.TrimSpace(string(output)))
			}
			return
		}
	}
	
	fmt.Printf("%s❌ No PostgreSQL service found%s\n", ColorRed, ColorReset)
}

func checkDatabaseExists(adminURL string) {
	db, err := sql.Open("postgres", adminURL)
	if err != nil {
		return
	}
	defer db.Close()
	
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = 'sistem_akuntansi')").Scan(&exists)
	if err != nil {
		fmt.Printf("%s⚠️  Could not check if sistem_akuntansi database exists%s\n", ColorYellow, ColorReset)
		return
	}
	
	if exists {
		fmt.Printf("%s✅ Database 'sistem_akuntansi' exists%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("%s❌ Database 'sistem_akuntansi' does not exist%s\n", ColorRed, ColorReset)
		fmt.Printf("    Create it with: %sCREATE DATABASE sistem_akuntansi;%s\n", ColorCyan, ColorReset)
	}
}

func updateEnvDatabase(newURL string) error {
	envFile := "../../.env"
	
	// Read existing .env file
	content, err := os.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %v", err)
	}
	
	lines := strings.Split(string(content), "\n")
	var newLines []string
	updated := false
	
	for _, line := range lines {
		line = strings.TrimRight(line, "\r") // Handle Windows line endings
		if strings.HasPrefix(line, "DATABASE_URL=") {
			newLines = append(newLines, "DATABASE_URL="+newURL)
			updated = true
		} else {
			newLines = append(newLines, line)
		}
	}
	
	if !updated {
		// Add DATABASE_URL if it doesn't exist
		newLines = append(newLines, "DATABASE_URL="+newURL)
	}
	
	// Write back to file
	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(envFile, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .env file: %v", err)
	}
	
	return nil
}