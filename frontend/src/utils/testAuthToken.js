/**
 * Test script to verify centralized auth token functionality
 * Run this in browser console to test token retrieval across different storage scenarios
 */

import { getAuthToken, getAuthHeaders, isAuthenticated, clearAuthTokens } from './authTokenUtils';

// Test scenarios
const testScenarios = [
  {
    name: 'Test 1: Token in correct location (localStorage.token)',
    setup: () => {
      clearAuthTokens();
      localStorage.setItem('token', 'test-token-123');
    },
    expected: 'test-token-123'
  },
  {
    name: 'Test 2: Token in legacy location (localStorage.authToken)',
    setup: () => {
      clearAuthTokens();
      localStorage.setItem('authToken', 'legacy-token-456');
    },
    expected: 'legacy-token-456'
  },
  {
    name: 'Test 3: Token in sessionStorage',
    setup: () => {
      clearAuthTokens();
      sessionStorage.setItem('token', 'session-token-789');
    },
    expected: 'session-token-789'
  },
  {
    name: 'Test 4: No token available',
    setup: () => {
      clearAuthTokens();
    },
    expected: null,
    shouldThrow: true
  }
];

// Run tests
export function runAuthTokenTests() {
  console.log('🔧 Starting Auth Token Utility Tests...\n');
  
  testScenarios.forEach((scenario, index) => {
    console.log(`${index + 1}. ${scenario.name}`);
    
    try {
      // Setup scenario
      scenario.setup();
      
      // Test getAuthToken
      if (scenario.shouldThrow) {
        try {
          getAuthToken();
          console.log('   ❌ FAIL: Expected error but got result');
        } catch (error) {
          console.log('   ✅ PASS: Correctly threw error:', error.message);
        }
        
        // Test with throwOnError = false
        const result = getAuthToken(false);
        if (result === scenario.expected) {
          console.log('   ✅ PASS: getAuthToken(false) returned null as expected');
        } else {
          console.log('   ❌ FAIL: Expected null, got:', result);
        }
      } else {
        const token = getAuthToken();
        if (token === scenario.expected) {
          console.log('   ✅ PASS: getAuthToken returned expected token');
          
          // Verify token was moved to correct location
          const correctToken = localStorage.getItem('token');
          if (correctToken === scenario.expected) {
            console.log('   ✅ PASS: Token correctly stored in localStorage.token');
          } else {
            console.log('   ❌ FAIL: Token not stored correctly in localStorage.token');
          }
        } else {
          console.log(`   ❌ FAIL: Expected "${scenario.expected}", got "${token}"`);
        }
      }
      
      // Test getAuthHeaders
      try {
        const headers = getAuthHeaders();
        if (!scenario.shouldThrow) {
          const expectedAuth = `Bearer ${scenario.expected}`;
          if (headers.Authorization === expectedAuth) {
            console.log('   ✅ PASS: getAuthHeaders returned correct Authorization header');
          } else {
            console.log(`   ❌ FAIL: Expected "${expectedAuth}", got "${headers.Authorization}"`);
          }
          
          if (headers['Content-Type'] === 'application/json') {
            console.log('   ✅ PASS: getAuthHeaders included Content-Type header');
          } else {
            console.log('   ❌ FAIL: Content-Type header missing or incorrect');
          }
        }
      } catch (error) {
        if (scenario.shouldThrow) {
          console.log('   ✅ PASS: getAuthHeaders correctly threw error');
        } else {
          console.log('   ❌ FAIL: getAuthHeaders unexpectedly threw error:', error.message);
        }
      }
      
      // Test isAuthenticated
      const authenticated = isAuthenticated();
      const expectedAuth = !scenario.shouldThrow;
      if (authenticated === expectedAuth) {
        console.log(`   ✅ PASS: isAuthenticated returned ${authenticated} as expected`);
      } else {
        console.log(`   ❌ FAIL: isAuthenticated returned ${authenticated}, expected ${expectedAuth}`);
      }
      
    } catch (error) {
      console.log('   ❌ ERROR: Test failed with error:', error.message);
    }
    
    console.log(''); // Empty line for readability
  });
  
  // Test clearAuthTokens
  console.log('5. Test clearAuthTokens functionality');
  localStorage.setItem('token', 'test');
  localStorage.setItem('authToken', 'test');
  localStorage.setItem('user', 'test');
  sessionStorage.setItem('token', 'test');
  
  clearAuthTokens();
  
  const allCleared = !localStorage.getItem('token') && 
                    !localStorage.getItem('authToken') && 
                    !localStorage.getItem('user') && 
                    !sessionStorage.getItem('token');
  
  if (allCleared) {
    console.log('   ✅ PASS: clearAuthTokens removed all auth data');
  } else {
    console.log('   ❌ FAIL: clearAuthTokens did not clear all data');
  }
  
  console.log('\n✨ Auth Token Utility Tests Completed!');
}

// Export for manual testing
if (typeof window !== 'undefined') {
  window.runAuthTokenTests = runAuthTokenTests;
  console.log('💡 Auth token test utility loaded. Run window.runAuthTokenTests() to test.');
}