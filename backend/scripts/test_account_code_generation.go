package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🧪 Account Code Generation System Test")
	fmt.Println("=====================================")

	generator := services.NewAccountCodeGenerator(db)

	// Test 1: Check non-standard codes
	fmt.Println("\n📋 Test 1: Identifying Non-Standard Codes")
	testNonStandardCodes(generator)

	// Test 2: Generate codes for different account types
	fmt.Println("\n📋 Test 2: Code Generation for Different Account Types")
	testCodeGeneration(generator)

	// Test 3: Sequential child code generation
	fmt.Println("\n📋 Test 3: Sequential Child Code Generation")
	testChildCodeGeneration(generator)

	// Test 4: Validation tests
	fmt.Println("\n📋 Test 4: Code Validation Tests")
	testCodeValidation(generator)

	// Test 5: Intelligent code suggestions
	fmt.Println("\n📋 Test 5: Intelligent Code Suggestions")
	testIntelligentSuggestions(generator)

	// Test 6: Fix "Kas Kecil" issue
	fmt.Println("\n📋 Test 6: Fix Kas Kecil Non-Standard Code Issue")
	testKasKecilFix(generator)

	fmt.Println("\n🎉 All tests completed!")
}

func testNonStandardCodes(generator *services.AccountCodeGenerator) {
	fixes, err := generator.FixNonStandardCodes()
	if err != nil {
		fmt.Printf("❌ Error getting non-standard codes: %v\n", err)
		return
	}

	fmt.Printf("Found %d non-standard codes:\n", len(fixes))
	for _, fix := range fixes {
		fmt.Printf("  ❌ %s (%s) %s → %s\n", 
			fix.CurrentCode, fix.AccountType, fix.AccountName, fix.SuggestedCode)
		fmt.Printf("     Reason: %s\n", fix.Reason)
	}

	if len(fixes) == 0 {
		fmt.Println("  ✅ No non-standard codes found!")
	}
}

func testCodeGeneration(generator *services.AccountCodeGenerator) {
	testCases := []struct {
		accountType string
		parentCode  string
		name        string
		description string
	}{
		{"ASSET", "", "Kas Besar", "Main cash account"},
		{"ASSET", "", "Bank BNI", "New bank account"},
		{"ASSET", "1100", "Kas Kecil Cabang A", "Child of current assets"},
		{"LIABILITY", "", "Utang Pajak", "Tax payable"},
		{"EQUITY", "", "Modal Tambahan", "Additional capital"},
		{"REVENUE", "", "Pendapatan Bunga", "Interest income"},
		{"EXPENSE", "", "Beban Internet", "Internet expense"},
	}

	fmt.Println("Testing code generation:")
	for _, tc := range testCases {
		code, err := generator.GetRecommendedCode(tc.accountType, tc.parentCode, tc.name)
		if err != nil {
			fmt.Printf("  ❌ %s (%s): Error - %v\n", tc.name, tc.accountType, err)
		} else {
			fmt.Printf("  ✅ %s (%s): %s\n", tc.name, tc.accountType, code)
		}
	}
}

func testChildCodeGeneration(generator *services.AccountCodeGenerator) {
	// Test child code generation for different parent formats
	parentCodes := []string{"1100", "1200", "2100"}
	
	fmt.Println("Testing child code generation:")
	for _, parent := range parentCodes {
		fmt.Printf("  Parent: %s\n", parent)
		for i := 0; i < 3; i++ {
			childCode, err := generator.GenerateNextCode("ASSET", parent)
			if err != nil {
				fmt.Printf("    ❌ Child %d: Error - %v\n", i+1, err)
			} else {
				fmt.Printf("    ✅ Child %d: %s\n", i+1, childCode)
			}
		}
	}
}

func testCodeValidation(generator *services.AccountCodeGenerator) {
	testCases := []struct {
		code        string
		accountType string
		parentCode  string
		shouldPass  bool
		description string
	}{
		{"1999", "ASSET", "", true, "Valid asset code"},
		{"2999", "ASSET", "", false, "Wrong type (2xxx for asset)"},
		{"5101", "EXPENSE", "", true, "Valid expense code"},
		{"1100-001", "ASSET", "", false, "Dash in main code (non-standard)"},
		{"1101", "ASSET", "1100", true, "Valid child code"},
		{"2101", "ASSET", "1100", false, "Child doesn't match parent"},
		{"", "ASSET", "", false, "Empty code"},
		{"INVALID", "ASSET", "", false, "Non-numeric code"},
	}

	fmt.Println("Testing code validation:")
	for _, tc := range testCases {
		err := generator.ValidateAccountCode(tc.code, tc.accountType, tc.parentCode)
		passed := (err == nil) == tc.shouldPass
		
		status := "✅"
		if !passed {
			status = "❌"
		}
		
		result := "VALID"
		if err != nil {
			result = err.Error()
		}
		
		fmt.Printf("  %s %s (%s): %s\n", status, tc.code, tc.description, result)
	}
}

func testIntelligentSuggestions(generator *services.AccountCodeGenerator) {
	testCases := []struct {
		name        string
		accountType string
		expected    string
	}{
		{"Kas", "ASSET", "1101"},
		{"Kas Kecil", "ASSET", "1102"},
		{"Bank Mandiri", "ASSET", "1103"},
		{"Piutang Dagang", "ASSET", "1201"},
		{"Persediaan Barang", "ASSET", "1301"},
		{"Utang Usaha", "LIABILITY", "2101"},
		{"PPN Keluaran", "LIABILITY", "2103"},
		{"Pendapatan Penjualan", "REVENUE", "4101"},
		{"Beban Gaji", "EXPENSE", "5101"},
		{"Beban Listrik", "EXPENSE", "5201"},
	}

	fmt.Println("Testing intelligent suggestions:")
	for _, tc := range testCases {
		code, err := generator.GetRecommendedCode(tc.accountType, "", tc.name)
		if err != nil {
			fmt.Printf("  ❌ %s: Error - %v\n", tc.name, err)
		} else {
			match := ""
			if code == tc.expected {
				match = " ✅ (expected)"
			} else {
				match = fmt.Sprintf(" (expected %s)", tc.expected)
			}
			fmt.Printf("  ✅ %s: %s%s\n", tc.name, code, match)
		}
	}
}

func testKasKecilFix(generator *services.AccountCodeGenerator) {
	fmt.Println("Demonstrating fix for 'Kas Kecil' issue:")
	fmt.Println("Current problematic scenario: Account 'Kas Kecil' with code '1100-001'")
	fmt.Println("Solution: Use sequential numbering instead of dash format")
	
	// Show what the current system would generate
	kasKecilCode, err := generator.GetRecommendedCode("ASSET", "", "Kas Kecil")
	if err != nil {
		fmt.Printf("❌ Error generating code for Kas Kecil: %v\n", err)
	} else {
		fmt.Printf("✅ Recommended code for 'Kas Kecil': %s\n", kasKecilCode)
	}
	
	// Show proper child codes for current assets
	fmt.Println("\nProper sequential codes for Current Assets (1100 parent):")
	for i := 0; i < 5; i++ {
		childCode, err := generator.GenerateNextCode("ASSET", "1100")
		if err != nil {
			fmt.Printf("  ❌ Child %d: Error - %v\n", i+1, err)
			break
		}
		
		// Simulate different account names
		accountNames := []string{"Kas", "Kas Kecil", "Kas USD", "Kas Petty Cash", "Kas Operasional"}
		accountName := "Account " + fmt.Sprintf("%d", i+1)
		if i < len(accountNames) {
			accountName = accountNames[i]
		}
		
		fmt.Printf("  ✅ %s: %s\n", accountName, childCode)
	}
	
	fmt.Println("\nRecommendations:")
	fmt.Println("1. Replace '1100-001' with '1101' or next sequential number")
	fmt.Println("2. Use consistent 4-digit numbering without dashes")
	fmt.Println("3. Implement auto-generation for new accounts")
	fmt.Println("4. Validate codes before creation to prevent non-standard formats")
}

func demonstrateAPIUsage() {
	fmt.Println("\n📡 API Usage Examples:")
	fmt.Println("================================")
	
	fmt.Println("1. Generate account code:")
	fmt.Println("   POST /api/v1/account-codes/generate")
	fmt.Println("   Body: {\"account_type\": \"ASSET\", \"account_name\": \"Kas Kecil\"}")
	
	fmt.Println("\n2. Validate account code:")
	fmt.Println("   GET /api/v1/account-codes/validate?code=1102&account_type=ASSET")
	
	fmt.Println("\n3. Check code availability:")
	fmt.Println("   GET /api/v1/account-codes/availability?code=1102")
	
	fmt.Println("\n4. Get next sequential code:")
	fmt.Println("   GET /api/v1/account-codes/next?account_type=ASSET&parent_code=1100")
	
	fmt.Println("\n5. Get code suggestions based on name:")
	fmt.Println("   GET /api/v1/account-codes/suggest?account_name=Kas Kecil&account_type=ASSET")
	
	fmt.Println("\n6. Get non-standard codes (admin only):")
	fmt.Println("   GET /api/v1/account-codes/non-standard")
	
	fmt.Println("\n7. Get accounting structure:")
	fmt.Println("   GET /api/v1/account-codes/structure")
}