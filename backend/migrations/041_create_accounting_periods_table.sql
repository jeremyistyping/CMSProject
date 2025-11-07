-- Migration: Create accounting_periods table for flexible period closing
-- Purpose: Track closed accounting periods (bulanan, triwulan, semester, tahunan)
-- Date: 2025-01-04

DO $$
BEGIN
    -- Check if table already exists
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'accounting_periods') THEN
        RAISE NOTICE '⏭️  Table accounting_periods already exists, skipping creation';
        
        -- Check if closing_journal_id column exists, if not add it
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                      WHERE table_name = 'accounting_periods' AND column_name = 'closing_journal_id') THEN
            RAISE NOTICE '🔧 Adding missing closing_journal_id column';
            ALTER TABLE accounting_periods ADD COLUMN closing_journal_id BIGINT;
            ALTER TABLE accounting_periods ADD CONSTRAINT fk_closing_journal 
                FOREIGN KEY (closing_journal_id) REFERENCES journal_entries(id) ON DELETE SET NULL;
        END IF;
        
        -- Check if net_income column exists, if not add it
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                      WHERE table_name = 'accounting_periods' AND column_name = 'net_income') THEN
            RAISE NOTICE '🔧 Adding missing net_income column';
            ALTER TABLE accounting_periods ADD COLUMN net_income DECIMAL(20,2) DEFAULT 0;
        END IF;
        
        -- Check if total_revenue column exists, if not add it
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                      WHERE table_name = 'accounting_periods' AND column_name = 'total_revenue') THEN
            RAISE NOTICE '🔧 Adding missing total_revenue column';
            ALTER TABLE accounting_periods ADD COLUMN total_revenue DECIMAL(20,2) DEFAULT 0;
        END IF;
        
        -- Check if total_expense column exists, if not add it
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                      WHERE table_name = 'accounting_periods' AND column_name = 'total_expense') THEN
            RAISE NOTICE '🔧 Adding missing total_expense column';
            ALTER TABLE accounting_periods ADD COLUMN total_expense DECIMAL(20,2) DEFAULT 0;
        END IF;
    ELSE
        RAISE NOTICE '🔧 Creating accounting_periods table';
        
        -- Create accounting_periods table
        CREATE TABLE accounting_periods (
            id BIGSERIAL PRIMARY KEY,
            start_date TIMESTAMP NOT NULL,
            end_date TIMESTAMP NOT NULL,
            description TEXT,
            is_closed BOOLEAN DEFAULT FALSE,
            is_locked BOOLEAN DEFAULT FALSE,
            closed_by BIGINT,
            closed_at TIMESTAMP,
            
            -- Closing summary
            total_revenue DECIMAL(20,2) DEFAULT 0,
            total_expense DECIMAL(20,2) DEFAULT 0,
            net_income DECIMAL(20,2) DEFAULT 0,
            closing_journal_id BIGINT,
            
            notes TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP,
            
            -- Foreign keys
            CONSTRAINT fk_closed_by_user FOREIGN KEY (closed_by) REFERENCES users(id) ON DELETE SET NULL,
            CONSTRAINT fk_closing_journal FOREIGN KEY (closing_journal_id) REFERENCES journal_entries(id) ON DELETE SET NULL
        );
    END IF;
    
    -- Create indexes (these are idempotent with IF NOT EXISTS)
    CREATE INDEX IF NOT EXISTS idx_accounting_periods_dates ON accounting_periods(start_date, end_date);
    CREATE INDEX IF NOT EXISTS idx_accounting_periods_status ON accounting_periods(is_closed, is_locked);
    CREATE INDEX IF NOT EXISTS idx_accounting_periods_closed_by ON accounting_periods(closed_by);
    CREATE INDEX IF NOT EXISTS idx_accounting_periods_journal ON accounting_periods(closing_journal_id);
END $$;

-- Comments
COMMENT ON TABLE accounting_periods IS 'Tracks closed accounting periods with flexible duration (monthly, quarterly, semester, annual)';
COMMENT ON COLUMN accounting_periods.is_closed IS 'Whether the period is closed';
COMMENT ON COLUMN accounting_periods.is_locked IS 'Whether the period is hard-locked (cannot be easily reopened)';
COMMENT ON COLUMN accounting_periods.net_income IS 'Net income for the period (total_revenue - total_expense)';
