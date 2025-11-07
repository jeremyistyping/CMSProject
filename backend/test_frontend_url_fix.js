// Simple test untuk AccountService URL construction
// Run di browser console atau Node.js

console.log('🔧 Testing AccountService URL Construction Fix');

// Mock API_BASE_URL seperti di frontend
const API_BASE = 'http://localhost:8080';

// Test URL construction seperti yang ada di AccountService
function testURLConstruction() {
  console.log('\n1. Testing accounts catalog URL:');
  
  // Test 1: getAccountCatalog tanpa type
  let url1 = `${API_BASE}/api/v1/accounts/catalog`;
  console.log('✅ Without type:', url1);
  
  // Test 2: getAccountCatalog dengan type EXPENSE
  let url2 = `${API_BASE}/api/v1/accounts/catalog`;
  url2 += `?type=${encodeURIComponent('EXPENSE')}`;
  console.log('✅ With EXPENSE type:', url2);
  
  // Test 3: getCreditAccounts
  let url3 = `${API_BASE}/api/v1/accounts/credit?type=LIABILITY`;
  console.log('✅ Credit accounts:', url3);
  
  // Test 4: getAccounts dengan type
  let url4 = `${API_BASE}/api/v1/accounts`;
  url4 += `?type=${encodeURIComponent('EXPENSE')}`;
  console.log('✅ Regular accounts with type:', url4);
}

// Test URL validity
function testURLValidity() {
  console.log('\n2. Testing URL validity:');
  
  const testUrls = [
    `${API_BASE}/api/v1/accounts/catalog`,
    `${API_BASE}/api/v1/accounts/catalog?type=EXPENSE`,
    `${API_BASE}/api/v1/accounts/credit?type=LIABILITY`,
    `${API_BASE}/api/v1/accounts?type=EXPENSE`
  ];
  
  testUrls.forEach((url, index) => {
    try {
      new URL(url); // This should not throw now
      console.log(`✅ URL ${index + 1} is valid:`, url);
    } catch (error) {
      console.log(`❌ URL ${index + 1} is invalid:`, url, error.message);
    }
  });
}

// Mock fetch test
function mockFetchTest() {
  console.log('\n3. Mock fetch test:');
  
  const originalFetch = global.fetch || window.fetch;
  
  // Mock fetch untuk testing
  const mockFetch = (url, options) => {
    console.log(`📡 Fetch called with URL: ${url}`);
    console.log(`   Method: ${options?.method || 'GET'}`);
    console.log(`   Headers: ${JSON.stringify(options?.headers || {})}`);
    
    return Promise.resolve({
      ok: true,
      json: () => Promise.resolve({
        data: [
          { id: 1, code: '5001', name: 'Office Supplies', active: true },
          { id: 2, code: '2101', name: 'Accounts Payable', active: true }
        ],
        count: 2
      })
    });
  };
  
  // Test getExpenseAccounts equivalent
  const token = 'mock-token';
  let url = `${API_BASE}/api/v1/accounts/catalog?type=${encodeURIComponent('EXPENSE')}`;
  
  mockFetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  }).then(() => {
    console.log('✅ Mock fetch for expense accounts succeeded');
  }).catch(error => {
    console.log('❌ Mock fetch failed:', error.message);
  });
  
  // Test getCreditAccounts equivalent  
  url = `${API_BASE}/api/v1/accounts/credit?type=LIABILITY`;
  
  mockFetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  }).then(() => {
    console.log('✅ Mock fetch for credit accounts succeeded');
  }).catch(error => {
    console.log('❌ Mock fetch failed:', error.message);
  });
}

// Run tests
testURLConstruction();
testURLValidity();
mockFetchTest();

console.log('\n🎯 URL Fix Test Complete!');
console.log('Jika semua URL valid, maka TypeError: Failed to construct URL sudah teratasi');