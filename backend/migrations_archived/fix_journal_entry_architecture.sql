-- =====================================================
-- HIGH PRIORITY #1: Fix Journal Entry Architecture
-- Ensure consistency between database schema and code
-- =====================================================

-- CRITICAL ISSUE: Journal Line vs Direct Journal Entry Inconsistency
-- Problem: Code sometimes creates journal entries WITHOUT journal lines
-- Solution: Enforce proper journal line usage and constraints

-- Step 1: Add missing foreign key constraints
ALTER TABLE journal_lines 
DROP CONSTRAINT IF EXISTS fk_journal_lines_entry,
ADD CONSTRAINT fk_journal_lines_entry 
FOREIGN KEY (journal_entry_id) REFERENCES journal_entries(id) 
ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE journal_lines 
DROP CONSTRAINT IF EXISTS fk_journal_lines_account,
ADD CONSTRAINT fk_journal_lines_account 
FOREIGN KEY (account_id) REFERENCES accounts(id) 
ON DELETE RESTRICT ON UPDATE CASCADE;

-- Step 2: Add triggers to ensure journal entry balance consistency
-- This trigger ensures journal entry totals match their lines
CREATE OR REPLACE FUNCTION sync_journal_entry_totals()
RETURNS TRIGGER AS $$
BEGIN
    -- Update parent journal entry totals when lines change
    UPDATE journal_entries 
    SET 
        total_debit = (
            SELECT COALESCE(SUM(debit_amount), 0) 
            FROM journal_lines 
            WHERE journal_entry_id = COALESCE(NEW.journal_entry_id, OLD.journal_entry_id)
        ),
        total_credit = (
            SELECT COALESCE(SUM(credit_amount), 0) 
            FROM journal_lines 
            WHERE journal_entry_id = COALESCE(NEW.journal_entry_id, OLD.journal_entry_id)
        ),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.journal_entry_id, OLD.journal_entry_id);

    -- Update is_balanced flag
    UPDATE journal_entries 
    SET is_balanced = (ABS(total_debit - total_credit) < 0.01)
    WHERE id = COALESCE(NEW.journal_entry_id, OLD.journal_entry_id);

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create triggers for journal line changes
DROP TRIGGER IF EXISTS trigger_sync_journal_totals_insert ON journal_lines;
CREATE TRIGGER trigger_sync_journal_totals_insert
    AFTER INSERT ON journal_lines
    FOR EACH ROW EXECUTE FUNCTION sync_journal_entry_totals();

DROP TRIGGER IF EXISTS trigger_sync_journal_totals_update ON journal_lines;
CREATE TRIGGER trigger_sync_journal_totals_update
    AFTER UPDATE ON journal_lines
    FOR EACH ROW EXECUTE FUNCTION sync_journal_entry_totals();

DROP TRIGGER IF EXISTS trigger_sync_journal_totals_delete ON journal_lines;
CREATE TRIGGER trigger_sync_journal_totals_delete
    AFTER DELETE ON journal_lines
    FOR EACH ROW EXECUTE FUNCTION sync_journal_entry_totals();

-- Step 3: Add constraint to ensure balanced entries
ALTER TABLE journal_entries 
DROP CONSTRAINT IF EXISTS chk_journal_entries_must_be_balanced,
ADD CONSTRAINT chk_journal_entries_must_be_balanced 
CHECK (
    (status = 'DRAFT') OR 
    (status != 'DRAFT' AND is_balanced = true AND ABS(total_debit - total_credit) < 0.01)
);

-- Step 4: Add constraint to prevent posting without journal lines
-- Create function to check if journal entry has lines
CREATE OR REPLACE FUNCTION journal_entry_has_lines(entry_id INTEGER)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (SELECT 1 FROM journal_lines WHERE journal_entry_id = entry_id);
END;
$$ LANGUAGE plpgsql;

-- Add constraint that posted entries must have lines
ALTER TABLE journal_entries 
DROP CONSTRAINT IF EXISTS chk_journal_entries_posted_must_have_lines,
ADD CONSTRAINT chk_journal_entries_posted_must_have_lines 
CHECK (
    (status = 'DRAFT') OR 
    (status != 'DRAFT' AND journal_entry_has_lines(id))
);

-- Step 5: Enhanced journal entry code uniqueness
-- Add unique constraint with proper index
CREATE UNIQUE INDEX IF NOT EXISTS idx_journal_entries_code_unique 
ON journal_entries(code) WHERE deleted_at IS NULL;

-- Step 6: Add reference integrity constraints
-- Ensure reference_id points to valid records when reference_type is set
CREATE OR REPLACE FUNCTION validate_journal_reference()
RETURNS TRIGGER AS $$
BEGIN
    -- Skip validation if reference_type or reference_id is NULL
    IF NEW.reference_type IS NULL OR NEW.reference_id IS NULL THEN
        RETURN NEW;
    END IF;

    -- Validate reference based on type
    CASE NEW.reference_type
        WHEN 'SALE' THEN
            IF NOT EXISTS (SELECT 1 FROM sales WHERE id = NEW.reference_id) THEN
                RAISE EXCEPTION 'Invalid sale reference_id: %', NEW.reference_id;
            END IF;
        WHEN 'PURCHASE' THEN
            IF NOT EXISTS (SELECT 1 FROM purchases WHERE id = NEW.reference_id) THEN
                RAISE EXCEPTION 'Invalid purchase reference_id: %', NEW.reference_id;
            END IF;
        WHEN 'PAYMENT' THEN
            IF NOT EXISTS (
                SELECT 1 FROM payments WHERE id = NEW.reference_id
                UNION ALL
                SELECT 1 FROM sale_payments WHERE id = NEW.reference_id
            ) THEN
                RAISE EXCEPTION 'Invalid payment reference_id: %', NEW.reference_id;
            END IF;
        -- Add more reference types as needed
    END CASE;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_validate_journal_reference ON journal_entries;
CREATE TRIGGER trigger_validate_journal_reference
    BEFORE INSERT OR UPDATE ON journal_entries
    FOR EACH ROW EXECUTE FUNCTION validate_journal_reference();

-- Step 7: Add account validation for journal lines
CREATE OR REPLACE FUNCTION validate_journal_line_account()
RETURNS TRIGGER AS $$
DECLARE
    account_record RECORD;
BEGIN
    -- Get account details
    SELECT is_active, is_header INTO account_record 
    FROM accounts 
    WHERE id = NEW.account_id AND deleted_at IS NULL;

    -- Check if account exists and is active
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Account ID % not found or deleted', NEW.account_id;
    END IF;

    -- Check if account is active
    IF NOT account_record.is_active THEN
        RAISE EXCEPTION 'Cannot post to inactive account ID %', NEW.account_id;
    END IF;

    -- Check if account is not a header account
    IF account_record.is_header THEN
        RAISE EXCEPTION 'Cannot post to header account ID %', NEW.account_id;
    END IF;

    -- Validate line amounts
    IF NEW.debit_amount < 0 OR NEW.credit_amount < 0 THEN
        RAISE EXCEPTION 'Journal line amounts cannot be negative';
    END IF;

    IF NEW.debit_amount > 0 AND NEW.credit_amount > 0 THEN
        RAISE EXCEPTION 'Journal line cannot have both debit and credit amounts';
    END IF;

    IF NEW.debit_amount = 0 AND NEW.credit_amount = 0 THEN
        RAISE EXCEPTION 'Journal line must have either debit or credit amount';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_validate_journal_line_account ON journal_lines;
CREATE TRIGGER trigger_validate_journal_line_account
    BEFORE INSERT OR UPDATE ON journal_lines
    FOR EACH ROW EXECUTE FUNCTION validate_journal_line_account();

-- Step 8: Create journal entry audit trail
CREATE OR REPLACE FUNCTION journal_entry_audit()
RETURNS TRIGGER AS $$
BEGIN
    -- Log journal entry changes
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_logs (table_name, record_id, action, old_values, new_values, user_id, created_at)
        VALUES ('journal_entries', NEW.id, 'INSERT', NULL, row_to_json(NEW), NEW.user_id, CURRENT_TIMESTAMP);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Only log if significant fields changed
        IF OLD.status != NEW.status OR OLD.total_debit != NEW.total_debit OR OLD.total_credit != NEW.total_credit THEN
            INSERT INTO audit_logs (table_name, record_id, action, old_values, new_values, user_id, created_at)
            VALUES ('journal_entries', NEW.id, 'UPDATE', row_to_json(OLD), row_to_json(NEW), NEW.user_id, CURRENT_TIMESTAMP);
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_logs (table_name, record_id, action, old_values, new_values, user_id, created_at)
        VALUES ('journal_entries', OLD.id, 'DELETE', row_to_json(OLD), NULL, OLD.user_id, CURRENT_TIMESTAMP);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_journal_entry_audit ON journal_entries;
CREATE TRIGGER trigger_journal_entry_audit
    AFTER INSERT OR UPDATE OR DELETE ON journal_entries
    FOR EACH ROW EXECUTE FUNCTION journal_entry_audit();

-- Step 9: Create data consistency check function
CREATE OR REPLACE FUNCTION check_journal_consistency()
RETURNS TABLE(
    entry_id INTEGER,
    issue_type TEXT,
    description TEXT,
    severity TEXT
) AS $$
BEGIN
    -- Check for unbalanced entries
    RETURN QUERY
    SELECT 
        je.id::INTEGER,
        'UNBALANCED'::TEXT,
        FORMAT('Entry %s: Debit=%.2f, Credit=%.2f, Difference=%.2f', 
               je.code, je.total_debit, je.total_credit, ABS(je.total_debit - je.total_credit))::TEXT,
        'HIGH'::TEXT
    FROM journal_entries je
    WHERE je.deleted_at IS NULL 
      AND (NOT je.is_balanced OR ABS(je.total_debit - je.total_credit) >= 0.01);

    -- Check for posted entries without lines
    RETURN QUERY
    SELECT 
        je.id::INTEGER,
        'NO_LINES'::TEXT,
        FORMAT('Posted entry %s has no journal lines', je.code)::TEXT,
        'CRITICAL'::TEXT
    FROM journal_entries je
    LEFT JOIN journal_lines jl ON je.id = jl.journal_entry_id
    WHERE je.deleted_at IS NULL 
      AND je.status != 'DRAFT' 
      AND jl.id IS NULL;

    -- Check for entries with mismatched totals vs lines
    RETURN QUERY
    SELECT 
        je.id::INTEGER,
        'TOTAL_MISMATCH'::TEXT,
        FORMAT('Entry %s: Header totals (D:%.2f C:%.2f) != Line totals (D:%.2f C:%.2f)', 
               je.code, je.total_debit, je.total_credit, 
               COALESCE(lines.total_debit, 0), COALESCE(lines.total_credit, 0))::TEXT,
        'HIGH'::TEXT
    FROM journal_entries je
    LEFT JOIN (
        SELECT 
            journal_entry_id,
            SUM(debit_amount) as total_debit,
            SUM(credit_amount) as total_credit
        FROM journal_lines
        GROUP BY journal_entry_id
    ) lines ON je.id = lines.journal_entry_id
    WHERE je.deleted_at IS NULL
      AND (ABS(je.total_debit - COALESCE(lines.total_debit, 0)) >= 0.01 
           OR ABS(je.total_credit - COALESCE(lines.total_credit, 0)) >= 0.01);

    -- Check for invalid references
    RETURN QUERY
    SELECT 
        je.id::INTEGER,
        'INVALID_REFERENCE'::TEXT,
        FORMAT('Entry %s has invalid %s reference_id: %s', 
               je.code, je.reference_type, je.reference_id)::TEXT,
        'MEDIUM'::TEXT
    FROM journal_entries je
    WHERE je.deleted_at IS NULL 
      AND je.reference_type IS NOT NULL 
      AND je.reference_id IS NOT NULL
      AND NOT CASE je.reference_type
          WHEN 'SALE' THEN EXISTS (SELECT 1 FROM sales WHERE id = je.reference_id)
          WHEN 'PURCHASE' THEN EXISTS (SELECT 1 FROM purchases WHERE id = je.reference_id)
          WHEN 'PAYMENT' THEN EXISTS (SELECT 1 FROM payments WHERE id = je.reference_id) OR
                               EXISTS (SELECT 1 FROM sale_payments WHERE id = je.reference_id)
          ELSE true  -- Allow other reference types for now
      END;
END;
$$ LANGUAGE plpgsql;

-- Step 10: Create maintenance procedures
CREATE OR REPLACE FUNCTION fix_journal_totals()
RETURNS TEXT AS $$
DECLARE
    fixed_count INTEGER := 0;
    entry_record RECORD;
BEGIN
    -- Fix all journal entries with incorrect totals
    FOR entry_record IN 
        SELECT je.id, 
               COALESCE(SUM(jl.debit_amount), 0) as correct_debit,
               COALESCE(SUM(jl.credit_amount), 0) as correct_credit
        FROM journal_entries je
        LEFT JOIN journal_lines jl ON je.id = jl.journal_entry_id
        WHERE je.deleted_at IS NULL
        GROUP BY je.id
        HAVING ABS(je.total_debit - COALESCE(SUM(jl.debit_amount), 0)) >= 0.01
            OR ABS(je.total_credit - COALESCE(SUM(jl.credit_amount), 0)) >= 0.01
    LOOP
        UPDATE journal_entries 
        SET 
            total_debit = entry_record.correct_debit,
            total_credit = entry_record.correct_credit,
            is_balanced = (ABS(entry_record.correct_debit - entry_record.correct_credit) < 0.01),
            updated_at = CURRENT_TIMESTAMP
        WHERE id = entry_record.id;
        
        fixed_count := fixed_count + 1;
    END LOOP;
    
    RETURN FORMAT('Fixed %s journal entries with incorrect totals', fixed_count);
END;
$$ LANGUAGE plpgsql;

-- Create a view for quick journal consistency monitoring
CREATE OR REPLACE VIEW journal_consistency_monitor AS
SELECT 
    'TOTAL_ENTRIES' as metric,
    COUNT(*)::TEXT as value,
    'INFO' as severity
FROM journal_entries 
WHERE deleted_at IS NULL

UNION ALL

SELECT 
    'UNBALANCED_ENTRIES' as metric,
    COUNT(*)::TEXT as value,
    CASE WHEN COUNT(*) > 0 THEN 'HIGH' ELSE 'OK' END as severity
FROM journal_entries 
WHERE deleted_at IS NULL 
  AND (NOT is_balanced OR ABS(total_debit - total_credit) >= 0.01)

UNION ALL

SELECT 
    'POSTED_WITHOUT_LINES' as metric,
    COUNT(*)::TEXT as value,
    CASE WHEN COUNT(*) > 0 THEN 'CRITICAL' ELSE 'OK' END as severity
FROM journal_entries je
LEFT JOIN journal_lines jl ON je.id = jl.journal_entry_id
WHERE je.deleted_at IS NULL 
  AND je.status != 'DRAFT' 
  AND jl.id IS NULL

ORDER BY severity DESC;

-- SUMMARY MESSAGE
SELECT 'Journal Entry Architecture Fix Applied Successfully!' as message,
       'Run SELECT * FROM journal_consistency_monitor; to check status' as next_step;