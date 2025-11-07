-- Disable Auto Balance Sync Triggers
-- These triggers cause DOUBLE POSTING because:
-- 1. Journal system already updates COA balances
-- 2. Triggers update COA AGAIN when CashBank changes
-- Result: COA gets updated TWICE!

-- Disable the main sync trigger
DROP TRIGGER IF EXISTS trigger_sync_cashbank_coa ON cash_banks;

-- Disable the recalc triggers (these also cause issues)
DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_insert ON cash_bank_transactions;
DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_update ON cash_bank_transactions;
DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_delete ON cash_bank_transactions;

-- KEEP the account balance validation trigger (it's useful)
-- DROP TRIGGER IF EXISTS trigger_validate_account_balance ON accounts;

SELECT 'Triggers disabled successfully' as status;

-- Show remaining triggers
SELECT 
    tgname as trigger_name,
    tgrelid::regclass as table_name,
    tgenabled as enabled,
    CASE tgenabled
        WHEN 'O' THEN 'ENABLED'
        WHEN 'D' THEN 'DISABLED'
        ELSE 'UNKNOWN'
    END as status
FROM pg_trigger
WHERE tgrelid IN ('cash_banks'::regclass, 'cash_bank_transactions'::regclass, 'accounts'::regclass)
AND tgisinternal = FALSE
ORDER BY table_name, tgname;

