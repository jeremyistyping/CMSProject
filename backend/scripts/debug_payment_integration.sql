-- Debug Payment Integration between Sales and Payment Management
-- Run this after creating a payment through Sales Management to verify data

-- Check if Payment records exist
SELECT 'PAYMENTS TABLE' as table_name;
SELECT 
    p.id,
    p.code,
    p.contact_id,
    c.name as contact_name,
    c.type as contact_type,
    p.date,
    p.amount,
    p.method,
    p.status,
    p.reference,
    p.notes,
    p.created_at
FROM payments p
JOIN contacts c ON p.contact_id = c.id
WHERE p.created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
ORDER BY p.created_at DESC
LIMIT 10;

SELECT '';
SELECT 'SALE_PAYMENTS TABLE' as table_name;
-- Check if SalePayment records exist (cross-reference)
SELECT 
    sp.id,
    sp.sale_id,
    s.code as sale_code,
    s.invoice_number,
    sp.payment_number,
    sp.date,
    sp.amount,
    sp.method,
    sp.payment_id as linked_payment_id,
    sp.created_at
FROM sale_payments sp
JOIN sales s ON sp.sale_id = s.id
WHERE sp.created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
ORDER BY sp.created_at DESC
LIMIT 10;

SELECT '';
SELECT 'PAYMENT_ALLOCATIONS TABLE' as table_name;
-- Check if PaymentAllocation records exist
SELECT 
    pa.id,
    pa.payment_id,
    p.code as payment_code,
    pa.invoice_id,
    s.code as sale_code,
    s.invoice_number,
    pa.allocated_amount,
    pa.created_at
FROM payment_allocations pa
LEFT JOIN payments p ON pa.payment_id = p.id
LEFT JOIN sales s ON pa.invoice_id = s.id
WHERE pa.created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
ORDER BY pa.created_at DESC
LIMIT 10;

SELECT '';
SELECT 'CROSS-REFERENCE CHECK' as check_name;
-- Check cross-references between payments and sales
SELECT 
    'Payment to Sale Cross-Reference' as reference_type,
    p.id as payment_id,
    p.code as payment_code,
    sp.id as sale_payment_id,
    sp.sale_id,
    s.code as sale_code,
    s.invoice_number,
    p.amount as payment_amount,
    sp.amount as sale_payment_amount
FROM payments p
LEFT JOIN sale_payments sp ON sp.payment_id = p.id
LEFT JOIN sales s ON sp.sale_id = s.id
WHERE p.created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
   OR sp.created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
ORDER BY p.created_at DESC, sp.created_at DESC
LIMIT 10;
