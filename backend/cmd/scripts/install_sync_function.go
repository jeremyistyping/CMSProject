package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
)

const sqlBigint = `
CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param BIGINT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    account_type_var VARCHAR(50);
    new_balance DECIMAL(20,2);
BEGIN
    SELECT type INTO account_type_var FROM accounts WHERE id = account_id_param;

    SELECT 
        CASE 
            WHEN account_type_var IN ('ASSET', 'EXPENSE') THEN 
                COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
            ELSE 
                COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
        END
    INTO new_balance
    FROM unified_journal_lines ujl
    LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
    WHERE ujl.account_id = account_id_param
      AND uje.status = 'POSTED';

    UPDATE accounts 
    SET balance = COALESCE(new_balance, 0),
        updated_at = NOW()
    WHERE id = account_id_param;
END;
$$;
`

const sqlInteger = `
CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param INTEGER)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM sync_account_balance_from_ssot(account_id_param::BIGINT);
END;
$$;
`

func main() {
	fmt.Println("🔧 Installing sync_account_balance_from_ssot functions (BIGINT + INTEGER wrapper)...")
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	if err := db.Exec(sqlBigint).Error; err != nil {
		log.Fatalf("❌ Failed to create BIGINT function: %v", err)
	}
	fmt.Println("✅ BIGINT variant installed")

	if err := db.Exec(sqlInteger).Error; err != nil {
		log.Fatalf("❌ Failed to create INTEGER wrapper: %v", err)
	}
	fmt.Println("✅ INTEGER wrapper installed")
	fmt.Println("🎉 Functions installed successfully")
}
