#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

console.log('🔍 FRONTEND PURCHASE REPORT VALIDATION');
console.log('======================================');

const frontendPath = path.join(__dirname, '..', '..', 'frontend');
const reportsPagePath = path.join(frontendPath, 'app', 'reports', 'page.tsx');

// Check if files exist
console.log('\n📁 CHECKING FILES');
console.log('=================');

const filesToCheck = [
    { path: reportsPagePath, name: 'Reports Page' },
    { path: path.join(frontendPath, 'src', 'translations', 'en.json'), name: 'English Translations' },
    { path: path.join(frontendPath, 'src', 'translations', 'id.json'), name: 'Indonesian Translations' }
];

for (const file of filesToCheck) {
    if (fs.existsSync(file.path)) {
        console.log(`✅ ${file.name}: EXISTS`);
    } else {
        console.log(`❌ ${file.name}: NOT FOUND`);
    }
}

// Check contents
console.log('\n🔍 CONTENT VALIDATION');
console.log('=====================');

try {
    const reportsContent = fs.readFileSync(reportsPagePath, 'utf8');
    
    const checks = [
        { 
            name: 'Purchase Report ID', 
            pattern: /purchase-report/g,
            content: reportsContent
        },
        { 
            name: 'Vendor Analysis Removed', 
            pattern: /vendor-analysis/g,
            content: reportsContent,
            shouldNotExist: true
        },
        { 
            name: 'Purchase Report Modal', 
            pattern: /ssotPROpen/g,
            content: reportsContent
        },
        { 
            name: 'Purchase Report API Call', 
            pattern: /ssot-reports\/purchase-report/g,
            content: reportsContent
        },
        { 
            name: 'Purchase Report State', 
            pattern: /setSSOTPRData/g,
            content: reportsContent
        }
    ];
    
    let allValid = true;
    
    for (const check of checks) {
        const matches = check.content.match(check.pattern);
        const matchCount = matches ? matches.length : 0;
        
        if (check.shouldNotExist) {
            if (matchCount === 0) {
                console.log(`✅ ${check.name}: REMOVED (${matchCount} occurrences)`);
            } else {
                console.log(`❌ ${check.name}: STILL EXISTS (${matchCount} occurrences)`);
                allValid = false;
            }
        } else {
            if (matchCount > 0) {
                console.log(`✅ ${check.name}: FOUND (${matchCount} occurrences)`);
            } else {
                console.log(`❌ ${check.name}: NOT FOUND`);
                allValid = false;
            }
        }
    }
    
    console.log('\n📋 TRANSLATION VALIDATION');
    console.log('=========================');
    
    // Check translations
    const enTranslations = JSON.parse(fs.readFileSync(path.join(frontendPath, 'src', 'translations', 'en.json'), 'utf8'));
    const idTranslations = JSON.parse(fs.readFileSync(path.join(frontendPath, 'src', 'translations', 'id.json'), 'utf8'));
    
    const translationChecks = [
        { key: 'reports.purchaseReport', lang: 'EN', translations: enTranslations, expected: 'Purchase Report' },
        { key: 'reports.purchaseReport', lang: 'ID', translations: idTranslations, expected: 'Laporan Pembelian' }
    ];
    
    for (const check of translationChecks) {
        const keys = check.key.split('.');
        let value = check.translations;
        
        for (const key of keys) {
            value = value[key];
            if (!value) break;
        }
        
        if (value === check.expected) {
            console.log(`✅ ${check.lang} Translation (${check.key}): "${value}"`);
        } else {
            console.log(`❌ ${check.lang} Translation (${check.key}): Expected "${check.expected}", got "${value}"`);
            allValid = false;
        }
    }
    
    console.log('\n🏆 FINAL ASSESSMENT');
    console.log('==================');
    
    if (allValid) {
        console.log('🎉 ALL VALIDATIONS PASSED!');
        console.log('✅ Frontend successfully updated for Purchase Report');
        console.log('✅ Vendor Analysis Report properly replaced');
        console.log('✅ Translations updated correctly');
        console.log('✅ Modal and API calls configured');
        
        console.log('\n🚀 NEXT STEPS');
        console.log('=============');
        console.log('1. Start the frontend server: npm run dev');
        console.log('2. Navigate to Reports page');
        console.log('3. Click "Purchase Report" card');
        console.log('4. Verify the new Purchase Report loads correctly');
        console.log('5. Test with date range: 2025-09-01 to 2025-09-30');
        
    } else {
        console.log('⚠️ SOME VALIDATIONS FAILED!');
        console.log('Please check the failed items above');
    }
    
} catch (error) {
    console.error('❌ Error reading files:', error.message);
}

console.log('\n📱 EXPECTED UI CHANGES');
console.log('=====================');
console.log('• Card title changed from "Vendor Analysis Report" to "Purchase Report"');
console.log('• Description updated to mention credible purchase analysis');
console.log('• Modal shows "Purchase Report (SSOT)" with purchase-focused UI');
console.log('• Data includes: Total Purchases, Amount, Paid, Outstanding');
console.log('• Vendor breakdown with payment method analysis');
console.log('• Tax analysis and financial metrics');
console.log('• Error handling for deprecated vendor analysis endpoint');