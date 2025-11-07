-- Migration: Create BIGINT-compatible sync_account_balance_from_ssot without DECLARE (parser-safe)
-- Date: 2025-09-26
-- Notes: Uses no DECLARE so the custom SQL parser won't split before BEGIN.

-- BIGINT variant
CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param BIGINT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE accounts a
    SET 
        balance = COALESCE((
            SELECT CASE 
                WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                    COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
                ELSE 
                    COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
            END
            FROM unified_journal_lines ujl 
            LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
            WHERE ujl.account_id = account_id_param 
              AND uje.status = 'POSTED'
        ), 0),
        updated_at = NOW()
    WHERE a.id = account_id_param;
END;
$$;

-- INTEGER overload delegating to BIGINT
CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param INTEGER)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM sync_account_balance_from_ssot(account_id_param::BIGINT);
END;
$$;
