-- Emergency: Drop ALL problematic triggers NOW
-- This should be run manually to fix immediate errors

-- Drop all triggers on accounts table
DROP TRIGGER IF EXISTS trigger_sync_cashbank_coa ON cash_banks;
DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_insert ON cash_bank_transactions;
DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_update ON cash_bank_transactions;
DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_delete ON cash_bank_transactions;
DROP TRIGGER IF EXISTS trigger_validate_account_balance ON accounts;
DROP TRIGGER IF EXISTS trigger_update_parent_balances ON accounts;
DROP TRIGGER IF EXISTS trigger_update_parent_account_balances ON accounts;
DROP TRIGGER IF EXISTS trigger_sync_account_balance ON accounts;
DROP TRIGGER IF EXISTS trigger_update_account_balance ON accounts;

-- List remaining triggers for verification
SELECT 
    trigger_name, 
    event_object_table, 
    action_statement
FROM information_schema.triggers 
WHERE event_object_schema = 'public'
AND event_object_table IN ('accounts', 'cash_banks', 'cash_bank_transactions')
ORDER BY event_object_table, trigger_name;

