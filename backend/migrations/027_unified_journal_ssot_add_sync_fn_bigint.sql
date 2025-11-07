-- Migration: Add BIGINT-compatible sync_account_balance_from_ssot (complex parser)
-- Date: 2025-09-26
-- This file name includes 'unified_journal_ssot' to ensure the complex SQL parser runs,
-- so CREATE FUNCTION blocks are executed correctly.

BEGIN;

-- BIGINT variant used by triggers and status-change routines
CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param BIGINT)
RETURNS VOID AS $$
DECLARE
    account_type_var VARCHAR(50);
    new_balance DECIMAL(20,2);
BEGIN
    -- Determine account type
    SELECT type INTO account_type_var
    FROM accounts 
    WHERE id = account_id_param;

    -- Compute balance from POSTED SSOT journal lines
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

    -- Update account balance
    UPDATE accounts 
    SET balance = COALESCE(new_balance, 0),
        updated_at = NOW()
    WHERE id = account_id_param;
END;
$$ LANGUAGE plpgsql;

-- INTEGER overload (delegates to BIGINT) for backward compatibility
CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param INTEGER)
RETURNS VOID AS $$
BEGIN
    PERFORM sync_account_balance_from_ssot(account_id_param::BIGINT);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION sync_account_balance_from_ssot(BIGINT) IS 'Synchronize accounts.balance from SSOT journal (BIGINT param)';
COMMENT ON FUNCTION sync_account_balance_from_ssot(INTEGER) IS 'Wrapper delegating to BIGINT variant';

COMMIT;
