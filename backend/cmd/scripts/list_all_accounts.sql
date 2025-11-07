-- List all accounts in database
SELECT code, name, type, balance, is_active
FROM accounts
WHERE deleted_at IS NULL
  AND (code LIKE '21%' OR code LIKE '41%')
ORDER BY code;

