// Test script for frontend AccountService public endpoints
// This script simulates the frontend calling the backend public endpoints

const API_BASE_URL = 'http://localhost:8080';

// Mock AccountService methods based on our fixed implementation
class TestAccountService {
  
  async handleResponse(response) {
    if (!response.ok) {
      let errorData;
      try {
        errorData = await response.json();
      } catch {
        errorData = {
          error: 'Network error',
          code: 'NETWORK_ERROR',
        };
      }

      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }
    
    return response.json();
  }

  // Get account catalog (PUBLIC ENDPOINT - no auth required)
  async getAccountCatalog(token, type) {
    let url = `${API_BASE_URL}/api/v1/accounts/catalog`;
    if (type) {
      url += `?type=${encodeURIComponent(type)}`;
    }
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    const result = await this.handleResponse(response);
    return result.data;
  }
  
  // Get expense accounts specifically for purchase items (PUBLIC ENDPOINT)
  async getExpenseAccounts(token) {
    return this.getAccountCatalog(undefined, 'EXPENSE');
  }
  
  // Get liability accounts for credit payment methods (PUBLIC ENDPOINT)
  async getCreditAccounts(token) {
    const url = `${API_BASE_URL}/api/v1/accounts/credit?type=LIABILITY`;
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    const result = await this.handleResponse(response);
    return result.data;
  }
}

// Run tests
async function runTests() {
  const accountService = new TestAccountService();
  
  console.log('🧪 Testing Frontend AccountService Public Endpoints\n');
  
  try {
    // Test 1: Get all account catalog
    console.log('📝 Test 1: Get all account catalog');
    const allAccounts = await accountService.getAccountCatalog();
    console.log(`✅ SUCCESS: Retrieved ${allAccounts.length} accounts`);
    console.log(`   Sample: ${allAccounts[0]?.code} - ${allAccounts[0]?.name}\n`);
    
    // Test 2: Get expense accounts
    console.log('📝 Test 2: Get expense accounts for purchase form');
    const expenseAccounts = await accountService.getExpenseAccounts();
    console.log(`✅ SUCCESS: Retrieved ${expenseAccounts.length} expense accounts`);
    console.log(`   Sample: ${expenseAccounts[0]?.code} - ${expenseAccounts[0]?.name}\n`);
    
    // Test 3: Get credit accounts (LIABILITY)
    console.log('📝 Test 3: Get credit accounts for payment method');
    const creditAccounts = await accountService.getCreditAccounts();
    console.log(`✅ SUCCESS: Retrieved ${creditAccounts.length} liability accounts`);
    if (creditAccounts.length > 0) {
      console.log(`   Sample: ${creditAccounts[0]?.code} - ${creditAccounts[0]?.name}\n`);
    } else {
      console.log(`   (No LIABILITY accounts found, this might be expected)\n`);
    }
    
    // Test 4: Get specific account types
    console.log('📝 Test 4: Get ASSET accounts');
    const assetAccounts = await accountService.getAccountCatalog(undefined, 'ASSET');
    console.log(`✅ SUCCESS: Retrieved ${assetAccounts.length} asset accounts\n`);
    
    console.log('🎉 ALL TESTS PASSED! Frontend service should work correctly.\n');
    
    console.log('📋 SUMMARY:');
    console.log(`   - Total accounts: ${allAccounts.length}`);
    console.log(`   - Expense accounts: ${expenseAccounts.length}`);
    console.log(`   - Credit accounts: ${creditAccounts.length}`);
    console.log(`   - Asset accounts: ${assetAccounts.length}`);
    
    console.log('\n✅ The purchase form dropdowns should now work without "Limited Access" errors!');
    
  } catch (error) {
    console.error('❌ TEST FAILED:', error.message);
    console.log('\n🔍 Troubleshooting:');
    console.log('   1. Ensure backend server is running on port 8080');
    console.log('   2. Verify public endpoints are correctly configured');
    console.log('   3. Check network connectivity');
  }
}

// Check if fetch is available (Node.js vs Browser)
if (typeof fetch === 'undefined') {
  console.log('⚠️  This test requires fetch API. Running with node-fetch...\n');
  import('node-fetch').then(({ default: fetch }) => {
    global.fetch = fetch;
    runTests();
  }).catch(() => {
    console.log('❌ Please install node-fetch: npm install node-fetch');
    console.log('   Or run this test in a browser console');
  });
} else {
  runTests();
}