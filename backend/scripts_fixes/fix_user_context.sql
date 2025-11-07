-- Fix User Context Issues for Receipt Creation
-- This script helps identify and fix user authentication issues

-- 1. Check current users in the database
SELECT 'Current Users in Database:' as info;
SELECT id, username, email, role, created_at, updated_at 
FROM users 
ORDER BY id;

-- 2. Check recent receipt creation attempts that might have failed
SELECT 'Recent Purchase Receipts:' as info;
SELECT pr.id, pr.purchase_id, pr.receipt_number, pr.received_by, 
       pr.status, pr.created_at, u.username as receiver_name
FROM purchase_receipts pr
LEFT JOIN users u ON u.id = pr.received_by
ORDER BY pr.created_at DESC
LIMIT 10;

-- 3. Check for any orphaned receipt records (receipts with invalid user references)
SELECT 'Orphaned Receipt Records (Invalid User References):' as info;
SELECT pr.id, pr.purchase_id, pr.receipt_number, pr.received_by, pr.status, pr.created_at
FROM purchase_receipts pr
LEFT JOIN users u ON u.id = pr.received_by
WHERE u.id IS NULL;

-- 4. Show current user ID range to help with debugging
SELECT 'User ID Range:' as info;
SELECT MIN(id) as min_user_id, MAX(id) as max_user_id, COUNT(*) as total_users
FROM users;

-- 5. Show recent purchases that can receive receipts
SELECT 'Recent Approved Purchases (Can Receive Receipts):' as info;
SELECT p.id, p.code, p.status, p.created_at, c.name as vendor_name,
       (SELECT COUNT(*) FROM purchase_receipts pr WHERE pr.purchase_id = p.id) as receipt_count
FROM purchases p
JOIN contacts c ON c.id = p.vendor_id
WHERE p.status IN ('APPROVED', 'PENDING')
ORDER BY p.created_at DESC
LIMIT 5;

-- 6. Recommendation: If you need to clean up any test receipts, uncomment below:
-- DELETE FROM purchase_receipt_items WHERE receipt_id IN (
--     SELECT id FROM purchase_receipts WHERE receipt_number LIKE 'TEST-%'
-- );
-- DELETE FROM purchase_receipts WHERE receipt_number LIKE 'TEST-%';

-- 7. Create a simple test user if needed (uncomment if you need a test user):
-- INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
-- VALUES ('test_user', 'test@test.com', 'dummy_hash', 'employee', NOW(), NOW())
-- ON CONFLICT (username) DO NOTHING;