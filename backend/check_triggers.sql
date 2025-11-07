-- Check if validate_journal_balance function exists
SELECT 
    'Function exists' as check_type,
    proname as name,
    CASE 
        WHEN prosrc LIKE '%Cannot post unbalanced journal entry%' THEN 'Yes - validates balance'
        ELSE 'Unknown'
    END as validates_balance
FROM pg_proc 
WHERE proname = 'validate_journal_balance'

UNION ALL

-- Check if trigger exists and is enabled
SELECT 
    'Trigger exists' as check_type,
    tgname as name,
    CASE 
        WHEN tgenabled = 'O' THEN 'ENABLED'
        WHEN tgenabled = 'D' THEN 'DISABLED'
        WHEN tgenabled = 'R' THEN 'ENABLED (replica)'
        WHEN tgenabled = 'A' THEN 'ENABLED (always)'
        ELSE 'UNKNOWN'
    END as validates_balance
FROM pg_trigger 
WHERE tgname = 'trg_validate_journal_balance';

-- Show trigger definition
SELECT pg_get_triggerdef(oid) as trigger_definition
FROM pg_trigger 
WHERE tgname = 'trg_validate_journal_balance';
