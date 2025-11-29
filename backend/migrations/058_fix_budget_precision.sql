-- Fix budget precision issue for projects table
-- Change budget and cost columns to use NUMERIC instead of DECIMAL to avoid precision loss

-- Alter projects table columns to use NUMERIC(20,0) for exact integer storage
ALTER TABLE projects ALTER COLUMN budget TYPE NUMERIC(20,0);
ALTER TABLE projects ALTER COLUMN actual_cost TYPE NUMERIC(20,0);
ALTER TABLE projects ALTER COLUMN material_cost TYPE NUMERIC(20,0);
ALTER TABLE projects ALTER COLUMN labor_cost TYPE NUMERIC(20,0);
ALTER TABLE projects ALTER COLUMN equipment_cost TYPE NUMERIC(20,0);
ALTER TABLE projects ALTER COLUMN overhead_cost TYPE NUMERIC(20,0);
ALTER TABLE projects ALTER COLUMN variance TYPE NUMERIC(20,0);

-- Fix any existing data that might have precision issues
-- Round all values to nearest integer
UPDATE projects SET budget = ROUND(budget);
UPDATE projects SET actual_cost = ROUND(actual_cost);
UPDATE projects SET material_cost = ROUND(material_cost);
UPDATE projects SET labor_cost = ROUND(labor_cost);
UPDATE projects SET equipment_cost = ROUND(equipment_cost);
UPDATE projects SET overhead_cost = ROUND(overhead_cost);
UPDATE projects SET variance = ROUND(variance);
