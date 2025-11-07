-- Test query to check the sales data and relationships
SELECT 
    s.id,
    s.code,
    s.invoice_number,
    s.due_date,
    s.payment_terms,
    c.name as customer_name,
    sp.name as sales_person_name
FROM sales s
LEFT JOIN contacts c ON s.customer_id = c.id
LEFT JOIN contacts sp ON s.sales_person_id = sp.id
WHERE s.id = 1;

-- Check if customer and sales person exist
SELECT id, name, type FROM contacts WHERE id IN (
    SELECT customer_id FROM sales WHERE id = 1
    UNION
    SELECT sales_person_id FROM sales WHERE id = 1 AND sales_person_id IS NOT NULL
);

-- Full sales data
SELECT * FROM sales WHERE id = 1;