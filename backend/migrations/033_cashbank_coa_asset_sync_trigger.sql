-- 033_cashbank_coa_asset_sync_trigger.sql
-- Ensures Cash & Bank (CashBank) balances stay synced with linked COA ASSET accounts
-- Idempotent: drops/replaces function and trigger safely

-- Drop existing to avoid conflicts
DROP TRIGGER IF EXISTS trg_sync_cashbank_coa ON cash_bank_transactions;
DROP FUNCTION IF EXISTS sync_cashbank_balance_to_coa();

-- Function guarded to only sync when linked COA account is of type ASSET
CREATE OR REPLACE FUNCTION sync_cashbank_balance_to_coa()
RETURNS TRIGGER AS $$
DECLARE
    coa_account_id INTEGER;
    transaction_sum DECIMAL(18,2);
    acct_type TEXT;
BEGIN
    -- Determine affected cash bank id
    -- and get linked account id
    IF TG_OP = 'DELETE' THEN
        SELECT account_id INTO coa_account_id FROM cash_banks WHERE id = OLD.cash_bank_id;
    ELSE
        SELECT account_id INTO coa_account_id FROM cash_banks WHERE id = NEW.cash_bank_id;
    END IF;

    -- Skip when not linked
    IF coa_account_id IS NULL THEN
        RETURN COALESCE(NEW, OLD);
    END IF;

    -- Only for ASSET accounts
    SELECT type INTO acct_type FROM accounts WHERE id = coa_account_id;
    IF acct_type IS NULL OR acct_type <> 'ASSET' THEN
        RETURN COALESCE(NEW, OLD);
    END IF;

    -- Source of truth: sum of transactions
    SELECT COALESCE(SUM(amount), 0) INTO transaction_sum
      FROM cash_bank_transactions
     WHERE cash_bank_id = COALESCE(NEW.cash_bank_id, OLD.cash_bank_id)
       AND deleted_at IS NULL;

    -- Update CashBank
    UPDATE cash_banks
       SET balance = transaction_sum,
           updated_at = NOW()
     WHERE id = COALESCE(NEW.cash_bank_id, OLD.cash_bank_id);

    -- Update linked COA account
    UPDATE accounts
       SET balance = transaction_sum,
           updated_at = NOW()
     WHERE id = coa_account_id;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create trigger (re-create)
CREATE TRIGGER trg_sync_cashbank_coa
    AFTER INSERT OR UPDATE OR DELETE ON cash_bank_transactions
    FOR EACH ROW
    EXECUTE FUNCTION sync_cashbank_balance_to_coa();