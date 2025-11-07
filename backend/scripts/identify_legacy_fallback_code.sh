#!/bin/bash

# =====================================================
# IDENTIFY LEGACY FALLBACK CODE
# =====================================================
# This script identifies all references to legacy journal system
# in the codebase to help with cleanup

echo "==========================================="
echo "🔍 IDENTIFYING LEGACY FALLBACK CODE"
echo "==========================================="
echo ""

cd "$(dirname "$0")/.." || exit 1

echo "Searching in backend/services directory..."
echo ""

# 1. Search for legacy journal service references
echo "1️⃣  Legacy JournalService References:"
echo "--------------------------------------"
grep -rn "journalService\.Create\|journalServiceV2" services/ --include="*.go" | head -20
echo ""

# 2. Search for direct journal_entries table references
echo "2️⃣  Direct journal_entries Table References:"
echo "---------------------------------------------"
grep -rn "journal_entries\|journal_lines" services/ --include="*.go" | grep -v "unified_journal" | head -20
echo ""

# 3. Search for legacy fallback functions
echo "3️⃣  Legacy Fallback Functions:"
echo "-------------------------------"
grep -rn "createAndPostPurchaseJournalEntries\|createSimpleSSOTPurchaseJournalFallback" services/ --include="*.go"
echo ""

# 4. Search for legacy create functions in sales
echo "4️⃣  Legacy Sales Journal Creation:"
echo "-----------------------------------"
grep -rn "CreateSalesJournal.*legacy\|salesJournalService\.Create" services/ --include="*.go"
echo ""

# 5. Count files with legacy references
echo "5️⃣  Files With Legacy References:"
echo "----------------------------------"
FILES_WITH_LEGACY=$(grep -rl "journal_entries\|journalService\.Create" services/ --include="*.go" | grep -v "unified_journal" | wc -l)
echo "Total files with legacy code: $FILES_WITH_LEGACY"
echo ""

# 6. Generate detailed report
echo "6️⃣  Generating Detailed Report..."
echo "----------------------------------"

# Create output file
OUTPUT_FILE="legacy_code_report.txt"
echo "Legacy Code Analysis Report" > "$OUTPUT_FILE"
echo "Generated: $(date)" >> "$OUTPUT_FILE"
echo "===========================================" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

echo "FILES WITH LEGACY JOURNAL REFERENCES:" >> "$OUTPUT_FILE"
echo "--------------------------------------" >> "$OUTPUT_FILE"
grep -rl "journal_entries\|journalService\.Create" services/ --include="*.go" | grep -v "unified_journal" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

echo "DETAILED OCCURRENCES:" >> "$OUTPUT_FILE"
echo "---------------------" >> "$OUTPUT_FILE"
grep -rn "journal_entries\|journalService\.Create\|createAndPost" services/ --include="*.go" | grep -v "unified_journal" >> "$OUTPUT_FILE"

echo ""
echo "✅ Report saved to: $OUTPUT_FILE"
echo ""

# 7. Priority cleanup list
echo "7️⃣  PRIORITY CLEANUP LIST:"
echo "--------------------------"
echo ""
echo "HIGH PRIORITY (Remove Fallback):"
echo "  📄 purchase_service.go - createAndPostPurchaseJournalEntries()"
echo "  📄 sales_service_v2.go - salesJournalService fallback"
echo "  📄 payment_service.go - legacy journal creation"
echo ""
echo "MEDIUM PRIORITY (Cleanup References):"
echo "  📄 enhanced_report_service.go - journal_entries queries"
echo "  📄 report_validation_service.go - legacy table checks"
echo ""
echo "LOW PRIORITY (Documentation/Comments):"
echo "  📄 Files with comments about journal_entries"
echo ""

echo "==========================================="
echo "✅ IDENTIFICATION COMPLETE"
echo "==========================================="
echo ""
echo "Next Steps:"
echo "1. Review $OUTPUT_FILE for details"
echo "2. Run check_legacy_journals.sql to verify data"
echo "3. If safe, run archive_legacy_journals.sql"
echo "4. Remove fallback code from identified files"
echo ""
