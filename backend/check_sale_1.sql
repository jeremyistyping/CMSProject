-- Check Sale #1 data to diagnose journal entry imbalance
SELECT 
    id,
    subtotal,
    ppn,
    pph,
    pph21_amount,
    pph23_amount,
    other_tax_additions,
    other_tax_deductions,
    shipping_cost,
    discount_amount,
    discount_percent,
    total_amount,
    -- Calculate expected total
    (subtotal + ppn + COALESCE(other_tax_additions, 0) + COALESCE(shipping_cost, 0) - COALESCE(pph, 0) - COALESCE(pph21_amount, 0) - COALESCE(pph23_amount, 0) - COALESCE(other_tax_deductions, 0) - COALESCE(discount_amount, 0)) as expected_total,
    -- Calculate difference
    total_amount - (subtotal + ppn + COALESCE(other_tax_additions, 0) + COALESCE(shipping_cost, 0) - COALESCE(pph, 0) - COALESCE(pph21_amount, 0) - COALESCE(pph23_amount, 0) - COALESCE(other_tax_deductions, 0) - COALESCE(discount_amount, 0)) as difference
FROM sales 
WHERE id = 1;
