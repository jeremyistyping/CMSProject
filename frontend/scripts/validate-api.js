#!/usr/bin/env node

/**
 * Quick API Endpoint Validation Script
 * 
 * This script validates that the API endpoints are correctly configured
 * and can be reached from the frontend.
 */

const { compareEndpointsWithSwagger, validateAPIAgainstSwagger } = require('../src/utils/apiDocumentation');

async function main() {
  console.log('🚀 Running API Endpoint Validation...\n');

  try {
    // Run endpoint comparison with Swagger
    console.log('📋 Step 1: Comparing endpoints with Swagger documentation...');
    const comparison = compareEndpointsWithSwagger();
    
    console.log('\n📊 Comparison Results:');
    console.log(`✅ Total Matches: ${comparison.matches}`);
    console.log(`🔄 Mismatches: ${comparison.mismatches}`);
    console.log(`⬅️ Missing in Frontend: ${comparison.missingInFrontend}`);
    console.log(`➡️ Missing in Swagger: ${comparison.missingInSwagger}`);
    console.log(`📈 Match Rate: ${Math.round((comparison.matches / comparison.totalChecked) * 100)}%`);

    if (comparison.mismatches > 0) {
      console.log('\n⚠️ Mismatched Endpoints:');
      comparison.comparisons
        .filter(c => c.status === 'mismatch')
        .forEach(c => {
          console.log(`  • Frontend: ${c.frontendPath}`);
          console.log(`    Swagger:  ${c.swaggerPath}`);
        });
    }

    if (comparison.recommendations.length > 0) {
      console.log('\n💡 Recommendations:');
      comparison.recommendations.forEach(rec => console.log(`  • ${rec}`));
    }

    // Run comprehensive validation
    console.log('\n📋 Step 2: Running comprehensive API validation...');
    const validation = await validateAPIAgainstSwagger();
    
    console.log('\n🏥 Health Check Results:');
    console.log(`✅ Healthy Endpoints: ${validation.healthCheck.healthyEndpoints}/${validation.healthCheck.totalEndpoints}`);
    
    if (validation.healthCheck.failedEndpoints.length > 0) {
      console.log('❌ Failed Endpoints:');
      validation.healthCheck.failedEndpoints.forEach(endpoint => {
        console.log(`  • ${endpoint}`);
      });
    }

    console.log('\n📋 Overall Recommendations:');
    validation.recommendations.forEach(rec => console.log(`  • ${rec}`));

    // Success criteria
    const matchRate = (comparison.matches / comparison.totalChecked) * 100;
    const healthRate = (validation.healthCheck.healthyEndpoints / validation.healthCheck.totalEndpoints) * 100;
    
    console.log('\n🎯 Production Readiness Score:');
    console.log(`📊 API Configuration: ${matchRate.toFixed(1)}%`);
    console.log(`🏥 API Health: ${healthRate.toFixed(1)}%`);
    
    const overallScore = (matchRate + healthRate) / 2;
    console.log(`🚀 Overall Score: ${overallScore.toFixed(1)}%`);
    
    if (overallScore >= 90) {
      console.log('\n🎉 READY FOR PRODUCTION! ✅');
    } else if (overallScore >= 70) {
      console.log('\n⚠️ NEEDS IMPROVEMENT before production');
    } else {
      console.log('\n🚨 NOT READY for production - Critical issues need fixing');
    }

  } catch (error) {
    console.error('❌ Validation failed:', error.message);
    process.exit(1);
  }
}

// Run if called directly
if (require.main === module) {
  main().catch(error => {
    console.error('💥 Fatal error:', error);
    process.exit(1);
  });
}

module.exports = { main };