-- Migration: Add BIGINT-compatible sync_account_balance_from_ssot function
-- Date: 2025-09-26
-- Purpose: Fix runtime error when triggers call sync_account_balance_from_ssot(bigint)
-- Context: Some schemas use BIGINT for account_id in unified_journal_lines.
-- The original function was defined for INTEGER, causing function not found errors.

BEGIN;

-- 1) Create or replace the BIGINT variant used by triggers
CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param BIGINT)
RETURNS VOID AS $$
DECLARE
    account_type_var VARCHAR(50);
    new_balance DECIMAL(20,2);
BEGIN
    -- Get account type (ensure account exists)
    SELECT type INTO account_type_var
    FROM accounts 
    WHERE id = account_id_param;

    -- Compute balance from posted SSOT journals with normal balance rules
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

    -- Update accounts.balance
    UPDATE accounts 
    SET balance = COALESCE(new_balance, 0),
        updated_at = NOW()
    WHERE id = account_id_param;

    -- Optional notice for diagnostics
    -- RAISE NOTICE 'Synced account % to balance %', account_id_param, COALESCE(new_balance, 0);
END;
$$ LANGUAGE plpgsql;

-- 2) Keep an INTEGER overload for backward compatibility; delegate to BIGINT
CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param INTEGER)
RETURNS VOID AS $$
BEGIN
    PERFORM sync_account_balance_from_ssot(account_id_param::BIGINT);
END;
$$ LANGUAGE plpgsql;

-- 3) Comment for documentation
COMMENT ON FUNCTION sync_account_balance_from_ssot(BIGINT) IS 'Synchronizes accounts.balance using posted SSOT journal lines (BIGINT param variant)';
COMMENT ON FUNCTION sync_account_balance_from_ssot(INTEGER) IS 'Wrapper that delegates to BIGINT variant for compatibility';

COMMIT;
