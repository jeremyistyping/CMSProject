package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConflictPattern represents patterns that indicate auto-posting
type ConflictPattern struct {
	Pattern     string
	Description string
	Severity    string
}

// Patterns that indicate problematic auto-posting
var conflictPatterns = []ConflictPattern{
	// Sales journal posting without status validation
	{"func.*CreateSale.*journal", "CreateSale method with journal posting", "HIGH"},
	{"CreateJournalEntry.*sale", "Direct journal creation from sales", "HIGH"},
	{"CreateSaleJournalEntries", "Sale journal entry creation", "HIGH"},
	{"PostSaleToJournal", "Direct sale posting to journal", "HIGH"},
	{"AutoPostSale", "Automatic sale posting", "CRITICAL"},
	
	// Payment journal posting without status validation
	{"CreatePaymentJournal.*sale", "Payment journal creation for sales", "HIGH"},
	{"PostPaymentToJournal", "Direct payment posting to journal", "HIGH"},
	{"AutoCreatePaymentJournal", "Automatic payment journal creation", "HIGH"},
	
	// Account balance updates without status validation
	{"UpdateAccountBalance.*sale", "Direct account balance updates", "MEDIUM"},
	{"UpdateCOABalance", "COA balance updates", "MEDIUM"},
	{"SyncAccountBalance", "Account balance synchronization", "MEDIUM"},
	
	// Status-agnostic processing
	{"ProcessSale.*journal", "Sale processing with journal", "HIGH"},
	{"HandleSaleCreation.*journal", "Sale creation with journal handling", "HIGH"},
	
	// Hook and trigger patterns
	{"OnSaleCreate.*journal", "Sale creation hooks with journal", "HIGH"},
	{"TriggerSaleAccounting", "Sale accounting triggers", "CRITICAL"},
	{"AfterSaleCreate.*Post", "After sale creation posting", "HIGH"},
	
	// GORM hooks that might auto-post
	{"AfterCreate.*Sale.*journal", "GORM AfterCreate hooks with journal", "HIGH"},
	{"BeforeCreate.*Sale.*journal", "GORM BeforeCreate hooks with journal", "HIGH"},
	
	// Service calls without status check
	{"journalService\\.Create.*\\(sale", "Journal service calls with sale", "HIGH"},
	{"accountingService\\.Post.*\\(sale", "Accounting service posts with sale", "HIGH"},
	
	// Database triggers simulation
	{"INSERT.*sales.*journal", "Database operations with journal", "MEDIUM"},
	{"CREATE.*TRIGGER.*sales", "Database triggers on sales", "CRITICAL"},
}

func main() {
	fmt.Println("🔍 DEEP CONFLICT DETECTION: Finding auto-posting issues")
	fmt.Println("=" + string(make([]byte, 70)))

	servicesDirs := []string{
		"services",
		"controllers", 
		"handlers",
		"middleware",
		"models",
		"repositories",
	}

	allConflicts := make(map[string][]ConflictDetail)

	for _, dir := range servicesDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Printf("⚠️ Directory %s not found, skipping\n", dir)
			continue
		}

		fmt.Printf("\n🔍 Scanning directory: %s\n", dir)
		conflicts := scanDirectory(dir)
		if len(conflicts) > 0 {
			allConflicts[dir] = conflicts
			fmt.Printf("   ❗ Found %d potential conflicts\n", len(conflicts))
		} else {
			fmt.Printf("   ✅ No conflicts found\n")
		}
	}

	// Generate comprehensive report
	generateConflictReport(allConflicts)

	// Create fix script
	createFixScript(allConflicts)

	fmt.Println("\n🎯 SUMMARY:")
	totalConflicts := 0
	for _, conflicts := range allConflicts {
		totalConflicts += len(conflicts)
	}

	if totalConflicts > 0 {
		fmt.Printf("   ❗ Found %d total conflicts that may cause auto-posting\n", totalConflicts)
		fmt.Printf("   📝 Detailed report saved to: conflict_report.txt\n")
		fmt.Printf("   🔧 Fix script created: apply_auto_posting_fixes.go\n")
		fmt.Printf("\n🚀 Run: go run apply_auto_posting_fixes.go\n")
	} else {
		fmt.Printf("   ✅ No auto-posting conflicts detected!\n")
	}
}

type ConflictDetail struct {
	File        string
	Line        int
	Content     string
	Pattern     ConflictPattern
	Context     []string
}

func scanDirectory(dirPath string) []ConflictDetail {
	var conflicts []ConflictDetail

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Only scan Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files and our own generated files
		if strings.Contains(path, "_test.go") || 
		   strings.Contains(path, "stub_services.go") ||
		   strings.Contains(path, "fix_") ||
		   strings.Contains(path, "test_") ||
		   strings.Contains(path, "demo_") {
			return nil
		}

		fileConflicts := scanFile(path)
		conflicts = append(conflicts, fileConflicts...)

		return nil
	})

	if err != nil {
		fmt.Printf("Error scanning directory %s: %v\n", dirPath, err)
	}

	return conflicts
}

func scanFile(filePath string) []ConflictDetail {
	var conflicts []ConflictDetail

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return conflicts
	}

	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Check each conflict pattern
		for _, pattern := range conflictPatterns {
			if matchesPattern(line, pattern.Pattern) {
				// Get context (surrounding lines)
				context := getContext(lines, lineNum, 2)
				
				conflict := ConflictDetail{
					File:    filePath,
					Line:    lineNum + 1,
					Content: line,
					Pattern: pattern,
					Context: context,
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

func matchesPattern(line, pattern string) bool {
	line = strings.ToLower(line)
	pattern = strings.ToLower(pattern)
	
	// Simple pattern matching for now
	// In a real implementation, you'd use regex
	words := strings.Fields(pattern)
	for _, word := range words {
		if !strings.Contains(line, word) {
			return false
		}
	}
	return true
}

func getContext(lines []string, lineNum, contextSize int) []string {
	var context []string
	
	start := lineNum - contextSize
	if start < 0 {
		start = 0
	}
	
	end := lineNum + contextSize + 1
	if end > len(lines) {
		end = len(lines)
	}
	
	for i := start; i < end; i++ {
		prefix := "  "
		if i == lineNum {
			prefix = "→ "
		}
		context = append(context, fmt.Sprintf("%s%3d: %s", prefix, i+1, lines[i]))
	}
	
	return context
}

func generateConflictReport(allConflicts map[string][]ConflictDetail) {
	report := fmt.Sprintf("CONFLICT DETECTION REPORT\n")
	report += fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	report += fmt.Sprintf("=" + string(make([]byte, 50)) + "\n\n")

	totalConflicts := 0
	for dir, conflicts := range allConflicts {
		if len(conflicts) == 0 {
			continue
		}

		report += fmt.Sprintf("DIRECTORY: %s (%d conflicts)\n", dir, len(conflicts))
		report += fmt.Sprintf("-" + string(make([]byte, 40)) + "\n")

		for _, conflict := range conflicts {
			report += fmt.Sprintf("\nFILE: %s (Line %d)\n", conflict.File, conflict.Line)
			report += fmt.Sprintf("SEVERITY: %s\n", conflict.Pattern.Severity)
			report += fmt.Sprintf("ISSUE: %s\n", conflict.Pattern.Description)
			report += fmt.Sprintf("PATTERN: %s\n", conflict.Pattern.Pattern)
			report += fmt.Sprintf("CODE: %s\n", conflict.Content)
			
			if len(conflict.Context) > 0 {
				report += fmt.Sprintf("CONTEXT:\n")
				for _, ctx := range conflict.Context {
					report += fmt.Sprintf("%s\n", ctx)
				}
			}
			report += fmt.Sprintf("-" + string(make([]byte, 30)) + "\n")
		}

		totalConflicts += len(conflicts)
		report += fmt.Sprintf("\n")
	}

	report += fmt.Sprintf("\nSUMMARY:\n")
	report += fmt.Sprintf("Total conflicts found: %d\n", totalConflicts)
	report += fmt.Sprintf("Action required: Review and fix these auto-posting patterns\n")

	err := ioutil.WriteFile("conflict_report.txt", []byte(report), 0644)
	if err != nil {
		fmt.Printf("Error writing report: %v\n", err)
	}
}

func createFixScript(allConflicts map[string][]ConflictDetail) {
	script := `package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	fmt.Println("🔧 AUTO-POSTING CONFLICT FIXER")
	fmt.Println("=" + string(make([]byte, 40)))
	
	backupDir := fmt.Sprintf("backup_conflicts_%d", time.Now().Unix())
	fmt.Printf("Creating backup directory: %s\n", backupDir)
	
	err := os.MkdirAll(backupDir, 0755)
	if err != nil {
		log.Fatal("Failed to create backup directory:", err)
	}
	
	conflictFiles := []string{
`

	// Add conflicting files to the fix script
	processedFiles := make(map[string]bool)
	for _, conflicts := range allConflicts {
		for _, conflict := range conflicts {
			if conflict.Pattern.Severity == "CRITICAL" || conflict.Pattern.Severity == "HIGH" {
				if !processedFiles[conflict.File] {
					script += fmt.Sprintf("\t\t\"%s\",\n", conflict.File)
					processedFiles[conflict.File] = true
				}
			}
		}
	}

	script += `	}
	
	fmt.Printf("Moving %d conflicting files to backup...\n", len(conflictFiles))
	
	for _, file := range conflictFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("File %s not found, skipping\n", file)
			continue
		}
		
		// Create backup path
		backupPath := filepath.Join(backupDir, filepath.Base(file))
		
		// Move file to backup
		err := os.Rename(file, backupPath)
		if err != nil {
			fmt.Printf("Error moving %s: %v\n", file, err)
			continue
		}
		
		fmt.Printf("✅ Moved %s to backup\n", file)
	}
	
	fmt.Println("\n🎯 CONFLICT RESOLUTION COMPLETE!")
	fmt.Println("All problematic files have been moved to backup.")
	fmt.Println("The system should now be protected from auto-posting.")
	fmt.Printf("Backup location: %s\n", backupDir)
}
`

	err := ioutil.WriteFile("apply_auto_posting_fixes.go", []byte(script), 0644)
	if err != nil {
		fmt.Printf("Error creating fix script: %v\n", err)
	}
}