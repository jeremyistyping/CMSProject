-- Migration to fix numeric overflow in purchase_items table
-- Update discount and tax fields from decimal(8,2) to decimal(15,2)

ALTER TABLE purchase_items 
    ALTER COLUMN discount TYPE decimal(15,2),
    ALTER COLUMN tax TYPE decimal(15,2);
