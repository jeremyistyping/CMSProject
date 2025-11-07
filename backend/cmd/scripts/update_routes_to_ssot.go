package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("🔄 Updating Routes to SSOT Only")
	fmt.Println("==============================")
	fmt.Println("This will:")
	fmt.Println("1. Remove old journal routes")
	fmt.Println("2. Keep only SSOT journal routes")  
	fmt.Println("3. Update route documentation")
	fmt.Println("")

	// Step 1: Update API routes
	fmt.Println("1. 📝 Updating API routes...")
	if err := updateAPIRoutes(); err != nil {
		log.Fatalf("Failed to update routes: %v", err)
	}

	// Step 2: Clean up old controllers
	fmt.Println("\n2. 🧹 Cleaning up old controllers...")
	if err := cleanupOldControllers(); err != nil {
		log.Printf("Warning: Failed to cleanup controllers: %v", err)
	}

	// Step 3: Update main.go imports
	fmt.Println("\n3. 🔧 Updating main.go imports...")
	if err := updateMainImports(); err != nil {
		log.Printf("Warning: Failed to update main imports: %v", err)
	}

	fmt.Println("\n🎉 Routes Updated Successfully!")
	fmt.Println("==============================")
	fmt.Println("✅ Old journal routes removed")
	fmt.Println("✅ SSOT routes active")
	fmt.Println("✅ Old controllers archived")
	
	fmt.Println("\n💡 Next Steps:")
	fmt.Println("• Test all SSOT endpoints")
	fmt.Println("• Update frontend to use new routes")
	fmt.Println("• Update API documentation")
}

func updateAPIRoutes() error {
	routesFile := "routes/routes.go"
	
	// Read current routes file
	content, err := ioutil.ReadFile(routesFile)
	if err != nil {
		return fmt.Errorf("failed to read routes file: %v", err)
	}

	originalContent := string(content)
	
	// Create backup
	backupFile := "routes/routes.go.backup"
	if err := ioutil.WriteFile(backupFile, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}
	fmt.Printf("   ✅ Backed up routes to %s\n", backupFile)

	// Remove old journal routes patterns
	updatedContent := removeOldJournalRoutes(originalContent)
	
	// Ensure SSOT routes are present
	updatedContent = ensureSSOTRoutes(updatedContent)

	// Write updated content
	if err := ioutil.WriteFile(routesFile, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated routes: %v", err)
	}

	fmt.Println("   ✅ API routes updated")
	return nil
}

func removeOldJournalRoutes(content string) string {
	// Remove old journal entries routes block
	updatedContent := content
	
	// Remove the entire journal-entries API group block
	journalEntriesPattern := `\s*// 📋 Journal Entry Management routes.*?\n\s*journalEntriesAPI := v1\.Group\("/journal-entries"\).*?\n\s*\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`
	re := regexp.MustCompile(journalEntriesPattern)
	updatedContent = re.ReplaceAllString(updatedContent, "")
	
	// Remove individual old journal routes if any remain
	patterns := []string{
		// Old journal entry routes
		`\s*journalEntriesAPI\.GET\([^\)]+\)[^\n]*`,
		`\s*journalEntriesAPI\.POST\([^\)]+\)[^\n]*`,
		`\s*journalEntriesAPI\.PUT\([^\)]+\)[^\n]*`,
		`\s*journalEntriesAPI\.DELETE\([^\)]+\)[^\n]*`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		updatedContent = re.ReplaceAllString(updatedContent, "")
	}

	return updatedContent
}

func ensureSSOTRoutes(content string) string {
	// Check if SSOT routes are already present
	if strings.Contains(content, "unifiedJournals") || strings.Contains(content, "unified-journals") || strings.Contains(content, "UnifiedJournalController") {
		fmt.Println("   ℹ️  SSOT unified journal routes already present")
		return content
	}

	// Add SSOT routes if not present
	ssotRoutes := `
	// SSOT Journal Routes
	journals := r.Group("/api/v1/journals")
	{
		journals.POST("/", ssot_journal_controller.CreateJournal)
		journals.GET("/", ssot_journal_controller.GetJournals)
		journals.GET("/:id", ssot_journal_controller.GetJournal)
		journals.PUT("/:id", ssot_journal_controller.UpdateJournal)
		journals.DELETE("/:id", ssot_journal_controller.DeleteJournal)
		journals.POST("/:id/post", ssot_journal_controller.PostJournal)
		journals.POST("/:id/reverse", ssot_journal_controller.ReverseJournal)
	}

	// Account balances route
	r.POST("/api/v1/account-balances/refresh", ssot_journal_controller.RefreshAccountBalances)
`

	// Find a good place to insert the routes (typically before the last })
	lines := strings.Split(content, "\n")
	var updatedLines []string

	inserted := false
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "}") && !inserted && i > 0 {
			// Insert before the last closing brace
			updatedLines = append([]string{ssotRoutes}, updatedLines...)
			inserted = true
		}
		updatedLines = append([]string{lines[i]}, updatedLines...)
	}

	if !inserted {
		// If couldn't find a good place, append at the end
		updatedLines = append(lines, ssotRoutes)
	}

	fmt.Println("   ✅ Added SSOT routes")
	return strings.Join(updatedLines, "\n")
}

func cleanupOldControllers() error {
	controllersDir := "./controllers"
	backupDir := "./controllers_backup"
	
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// List of old controller files to backup and remove
	oldControllerFiles := []string{
		"journal_controller.go",
		"journal_entry_controller.go", 
		"journal_entries_controller.go",
		"Journal_controller.go",
		"JournalController.go",
		"JournalEntry_controller.go",
	}

	backedUpCount := 0
	for _, fileName := range oldControllerFiles {
		filePath := fmt.Sprintf("%s/%s", controllersDir, fileName)
		
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		// Read file content to check if it's not an SSOT controller
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to read %s: %v\n", fileName, err)
			continue
		}

		// Skip if it's an SSOT controller
		if strings.Contains(string(content), "SSOT") || strings.Contains(string(content), "ssot") {
			fmt.Printf("   ℹ️  Skipping SSOT controller: %s\n", fileName)
			continue
		}

		// Copy to backup directory
		backupPath := fmt.Sprintf("%s/%s", backupDir, fileName)
		if err := ioutil.WriteFile(backupPath, content, 0644); err != nil {
			fmt.Printf("   ⚠️  Failed to backup %s: %v\n", fileName, err)
			continue
		}

		// Remove original file
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("   ⚠️  Failed to remove %s: %v\n", fileName, err)
			continue
		}

		fmt.Printf("   ✅ Archived and removed %s\n", fileName)
		backedUpCount++
	}

	if backedUpCount == 0 {
		fmt.Println("   ℹ️  No old controller files found to cleanup")
	} else {
		fmt.Printf("   ✅ Cleaned up %d controller files\n", backedUpCount)
	}

	return nil
}

func updateMainImports() error {
	mainFile := "./main.go"
	
	// Read main.go
	content, err := ioutil.ReadFile(mainFile)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %v", err)
	}

	originalContent := string(content)
	
	// Create backup
	backupFile := "./main.go.backup"
	if err := ioutil.WriteFile(backupFile, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Remove old controller imports
	updatedContent := removeOldControllerImports(originalContent)

	// Ensure SSOT controller import is present
	updatedContent = ensureSSOTImport(updatedContent)

	// Write updated content
	if err := ioutil.WriteFile(mainFile, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated main.go: %v", err)
	}

	fmt.Printf("   ✅ Updated main.go imports (backup: %s)\n", backupFile)
	return nil
}

func removeOldControllerImports(content string) string {
	// Remove old journal controller imports
	patterns := []string{
		`\s*"[^"]*journal[^"]*controller"[^s][^s][^o][^t].*\n`,
		`\s*"[^"]*Journal[^"]*Controller"[^S][^S][^O][^T].*\n`,
	}

	updatedContent := content
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		updatedContent = re.ReplaceAllString(updatedContent, "")
	}

	return updatedContent
}

func ensureSSOTImport(content string) string {
	// Check if SSOT import is already present
	if strings.Contains(content, "ssot_journal_controller") {
		return content
	}

	// Add SSOT import if not present
	importPattern := `import \(`
	ssotImport := `import (
	"app-sistem-akuntansi/controllers/ssot_journal_controller"`

	if strings.Contains(content, importPattern) {
		// Replace existing import block
		re := regexp.MustCompile(`import \(`)
		updatedContent := re.ReplaceAllString(content, ssotImport)
		fmt.Println("   ✅ Added SSOT controller import")
		return updatedContent
	}

	fmt.Println("   ℹ️  Could not automatically add SSOT import - please add manually")
	return content
}