-- Migration to update tax field sizes to prevent numeric overflow
-- Update sales table tax fields from decimal(8,2) to decimal(15,2)

ALTER TABLE sales 
    ALTER COLUMN tax TYPE decimal(15,2),
    ALTER COLUMN ppn TYPE decimal(15,2),
    ALTER COLUMN pph TYPE decimal(15,2),
    ALTER COLUMN total_tax TYPE decimal(15,2);

-- Update sale_items table tax fields from decimal(8,2) to decimal(15,2)
ALTER TABLE sale_items 
    ALTER COLUMN ppn_amount TYPE decimal(15,2),
    ALTER COLUMN pph_amount TYPE decimal(15,2),
    ALTER COLUMN total_tax TYPE decimal(15,2),
    ALTER COLUMN discount TYPE decimal(15,2),
    ALTER COLUMN tax TYPE decimal(15,2);
