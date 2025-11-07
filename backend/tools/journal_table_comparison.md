# Journal Table Structure Comparison

## UNIFIED JOURNAL TABLES (Legacy)

### unified_journal_ledger
```sql
CREATE TABLE unified_journal_ledger (
    id SERIAL PRIMARY KEY,
    entry_number VARCHAR,
    entry_date DATE,
    description TEXT,
    status VARCHAR DEFAULT 'POSTED',
    -- Other legacy fields
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### unified_journal_lines  
```sql
CREATE TABLE unified_journal_lines (
    id SERIAL PRIMARY KEY,
    journal_id INTEGER REFERENCES unified_journal_ledger(id),
    account_id INTEGER REFERENCES accounts(id),
    debit_amount DECIMAL(15,2) DEFAULT 0,
    credit_amount DECIMAL(15,2) DEFAULT 0,
    description TEXT,
    line_number INTEGER,
    created_at TIMESTAMP
);
```

## SSOT JOURNAL TABLES (New/Corrected)

### ssot_journal_entries
```sql
CREATE TABLE ssot_journal_entries (
    id SERIAL PRIMARY KEY,
    entry_number VARCHAR UNIQUE,
    entry_date DATE,
    description TEXT,
    status VARCHAR DEFAULT 'POSTED',
    total_debit DECIMAL(15,2),
    total_credit DECIMAL(15,2),
    is_balanced BOOLEAN,
    is_auto_generated BOOLEAN,
    posted_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### ssot_journal_lines
```sql
CREATE TABLE ssot_journal_lines (
    id SERIAL PRIMARY KEY,
    journal_entry_id INTEGER REFERENCES ssot_journal_entries(id),
    account_id INTEGER REFERENCES accounts(id),
    debit_amount DECIMAL(15,2),
    credit_amount DECIMAL(15,2),
    description TEXT,
    line_number INTEGER,
    created_at TIMESTAMP
);
```

## KEY DIFFERENCES

| Feature | Unified (Legacy) | SSOT (New) |
|---------|-----------------|------------|
| **Naming** | `unified_journal_ledger` | `ssot_journal_entries` |
| **Lines Reference** | `journal_id` | `journal_entry_id` |
| **Balance Tracking** | ❌ No totals | ✅ `total_debit`, `total_credit` |
| **Balance Check** | ❌ No validation | ✅ `is_balanced` field |
| **Auto Generation** | ❌ Not tracked | ✅ `is_auto_generated` |
| **Posted Status** | ❌ Basic status | ✅ `posted_at` timestamp |
| **Data Quality** | ⚠️ May have errors | ✅ Clean, validated data |

## CONTROLLER QUERY DIFFERENCES

### Before Fix (Wrong):
```sql
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
WHERE uje.status = 'POSTED'
```

### After Fix (Correct):
```sql  
FROM accounts a
LEFT JOIN ssot_journal_lines sjl ON sjl.account_id = a.id
LEFT JOIN ssot_journal_entries sje ON sje.id = sjl.journal_entry_id
WHERE sje.status = 'POSTED'
```

## IMPACT ON SYSTEM

### Legacy Issues:
- ❌ Potential duplicate entries
- ❌ Inconsistent accounting logic  
- ❌ No balance validation
- ❌ Hard to audit

### SSOT Benefits:
- ✅ Single source of truth
- ✅ Correct double-entry logic
- ✅ Automatic balance validation
- ✅ Complete audit trail
- ✅ Data integrity enforced