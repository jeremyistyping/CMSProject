package main

import (
	"app-sistem-akuntansi/config"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("🔧 Dynamic Swagger Fixer & Analyzer")
	fmt.Println("=" + string(make([]rune, 40)))

	// Step 1: Analyze current issues
	fmt.Println("\n📋 Analyzing current Swagger issues...")
	issues := config.CheckAndFixCommonIssues()
	
	if len(issues) == 0 {
		fmt.Println("✅ No issues detected in swagger.json")
	} else {
		fmt.Printf("⚠️ Found %d issues:\n", len(issues))
		for i, issue := range issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}

	// Step 2: Apply fixes
	fmt.Println("\n🔧 Applying automatic fixes...")
	swaggerConfig, err := config.ValidateAndFixSwagger()
	if err != nil {
		log.Fatalf("❌ Failed to fix swagger: %v", err)
	}

	// Step 3: Show results
	if len(swaggerConfig.Fixes) > 0 {
		fmt.Printf("\n✅ Applied %d fixes:\n", len(swaggerConfig.Fixes))
		for i, fix := range swaggerConfig.Fixes {
			fmt.Printf("  %d. %s\n", i+1, fix)
		}
	} else {
		fmt.Println("\n✅ No fixes needed - swagger.json is already correct")
	}

	if len(swaggerConfig.Errors) > 0 {
		fmt.Printf("\n❌ %d errors encountered:\n", len(swaggerConfig.Errors))
		for i, error := range swaggerConfig.Errors {
			fmt.Printf("  %d. %s\n", i+1, error)
		}
	}

	// Step 4: Generate report
	fmt.Println("\n📊 Generating fix report...")
	report := map[string]interface{}{
		"timestamp":    fmt.Sprintf("%d", swaggerConfig),
		"host":         swaggerConfig.Host,
		"base_path":    swaggerConfig.BasePath,
		"fixes_count":  len(swaggerConfig.Fixes),
		"errors_count": len(swaggerConfig.Errors),
		"fixes":        swaggerConfig.Fixes,
		"errors":       swaggerConfig.Errors,
		"status":       getStatus(swaggerConfig),
	}

	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	err = os.WriteFile("swagger_fix_report.json", reportJSON, 0644)
	if err != nil {
		log.Printf("⚠️ Failed to write report: %v", err)
	} else {
		fmt.Println("📄 Fix report saved to swagger_fix_report.json")
	}

	// Step 5: Next steps
	fmt.Println("\n🚀 Next Steps:")
	fmt.Println("  1. Restart your Go server to apply the fixes")
	fmt.Println("  2. Visit http://localhost:8080/swagger/index.html")
	fmt.Println("  3. Check that all endpoints now have correct /api/v1 prefixes")
	
	if len(swaggerConfig.Fixes) > 0 {
		fmt.Println("  4. Original swagger.json has been backed up with timestamp")
	}

	fmt.Println("\n🎯 Dynamic Swagger fixing completed!")
}

func getStatus(config *config.DynamicSwaggerConfig) string {
	if len(config.Errors) > 0 {
		return "ERROR"
	} else if len(config.Fixes) > 0 {
		return "FIXED"
	} else {
		return "OK"
	}
}