-- Check for triggers on cash_banks table
SELECT 
    tgname as trigger_name,
    tgrelid::regclass as table_name,
    tgtype,
    tgenabled,
    pg_get_triggerdef(oid) as trigger_definition
FROM pg_trigger
WHERE tgrelid = 'cash_banks'::regclass
AND tgisinternal = FALSE
ORDER BY tgname;

-- Check for triggers on cash_bank_transactions table
SELECT 
    tgname as trigger_name,
    tgrelid::regclass as table_name,
    tgtype,
    tgenabled,
    pg_get_triggerdef(oid) as trigger_definition
FROM pg_trigger
WHERE tgrelid = 'cash_bank_transactions'::regclass
AND tgisinternal = FALSE
ORDER BY tgname;

-- Check for triggers on accounts table
SELECT 
    tgname as trigger_name,
    tgrelid::regclass as table_name,
    tgtype,
    tgenabled,
    pg_get_triggerdef(oid) as trigger_definition
FROM pg_trigger
WHERE tgrelid = 'accounts'::regclass
AND tgisinternal = FALSE
ORDER BY tgname;

