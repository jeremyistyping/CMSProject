-- Add category progress fields to daily_updates table

DO $$
BEGIN
    -- Add progress column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_updates' AND column_name = 'progress') THEN
        ALTER TABLE daily_updates ADD COLUMN progress NUMERIC(5,2) DEFAULT 0;
    END IF;

    -- Add foundation_progress column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_updates' AND column_name = 'foundation_progress') THEN
        ALTER TABLE daily_updates ADD COLUMN foundation_progress NUMERIC(5,2) DEFAULT 0;
    END IF;

    -- Add utilities_progress column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_updates' AND column_name = 'utilities_progress') THEN
        ALTER TABLE daily_updates ADD COLUMN utilities_progress NUMERIC(5,2) DEFAULT 0;
    END IF;

    -- Add interior_progress column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_updates' AND column_name = 'interior_progress') THEN
        ALTER TABLE daily_updates ADD COLUMN interior_progress NUMERIC(5,2) DEFAULT 0;
    END IF;

    -- Add equipment_progress column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_updates' AND column_name = 'equipment_progress') THEN
        ALTER TABLE daily_updates ADD COLUMN equipment_progress NUMERIC(5,2) DEFAULT 0;
    END IF;
END $$;
