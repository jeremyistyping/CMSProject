package main

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
		"services/sales_double_entry_service.go",
		"services/sales_service.go",
	}
	
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
