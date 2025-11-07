package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	fmt.Println("=== CREATING HEADER ACCOUNT BALANCE PREVENTION ===")
	
	// Create a database trigger to prevent header accounts from having non-zero balance
	triggerSQL := `
		-- Create function to prevent header accounts from having non-zero balance
		CREATE OR REPLACE FUNCTION prevent_header_account_balance()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Check if account is a header account and trying to set non-zero balance
			IF NEW.is_header = true AND NEW.balance != 0 THEN
				RAISE EXCEPTION 'Header accounts cannot have non-zero balance. Account: % (%), Balance: %', 
					NEW.code, NEW.name, NEW.balance
				USING ERRCODE = 'P0001';
			END IF;
			
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		-- Drop existing trigger if exists
		DROP TRIGGER IF EXISTS trg_prevent_header_balance ON accounts;

		-- Create trigger
		CREATE TRIGGER trg_prevent_header_balance
			BEFORE INSERT OR UPDATE ON accounts
			FOR EACH ROW
			EXECUTE FUNCTION prevent_header_account_balance();
	`

	// Execute the trigger creation
	if err := db.Exec(triggerSQL).Error; err != nil {
		log.Fatalf("Failed to create trigger: %v", err)
	}

	fmt.Println("✅ Header account balance prevention trigger created successfully!")
	
	// Test the trigger by trying to set a header account balance
	fmt.Println("\n=== TESTING TRIGGER ===")
	
	testSQL := `
		UPDATE accounts 
		SET balance = 100 
		WHERE code = '1000' AND is_header = true
	`
	
	err := db.Exec(testSQL).Error
	if err != nil {
		fmt.Printf("✅ Trigger working correctly - prevented header account balance update: %v\n", err)
	} else {
		fmt.Println("❌ Trigger not working - header account balance was updated!")
	}

	fmt.Println("\n✅ Header account balance prevention system is active!")
}
