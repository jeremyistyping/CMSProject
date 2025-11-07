-- Check for accounts with code '1008' including soft deleted ones
SELECT id, code, name, deleted_at, created_at, updated_at 
FROM accounts 
WHERE code = '1008';

-- If soft deleted records exist, permanently delete them
DELETE FROM accounts WHERE code = '1008' AND deleted_at IS NOT NULL;

-- Check if there are any remaining records
SELECT id, code, name, deleted_at, created_at, updated_at 
FROM accounts 
WHERE code = '1008';

-- Also check for any other potential duplicate codes
SELECT code, COUNT(*) as count
FROM accounts 
GROUP BY code 
HAVING COUNT(*) > 1;
