// Test script to verify SSOT payment endpoints
const axios = require('axios');

const API_BASE_URL = 'http://localhost:8080/api/v1';

// Test function
async function testPaymentEndpoints() {
    console.log('🔍 Testing SSOT Payment Endpoints...\n');
    
    // Test 1: Check if SSOT endpoints exist
    console.log('1. Checking SSOT endpoint availability:');
    
    try {
        // This should return 401/403 (authentication required) instead of 404
        const response = await axios.get(`${API_BASE_URL}/payments/ssot`, {
            validateStatus: () => true // Accept all status codes
        });
        
        console.log(`   GET /payments/ssot - Status: ${response.status}`);
        
        if (response.status === 404) {
            console.log('   ❌ SSOT endpoint not found!');
        } else if (response.status === 401 || response.status === 403) {
            console.log('   ✅ SSOT endpoint exists (requires authentication)');
        } else {
            console.log(`   ℹ️  Unexpected status: ${response.status}`);
        }
        
    } catch (error) {
        if (error.code === 'ECONNREFUSED') {
            console.log('   ❌ Backend server is not running');
            return;
        }
        console.log(`   ❌ Error: ${error.message}`);
    }
    
    // Test 2: Check deprecated notice endpoint
    console.log('\n2. Checking deprecated notice:');
    
    try {
        const response = await axios.get(`${API_BASE_URL}/payments/deprecated-notice`, {
            validateStatus: () => true
        });
        
        console.log(`   Status: ${response.status}`);
        if (response.status === 200) {
            console.log('   ✅ Deprecated notice available');
            console.log('   Available endpoints:');
            response.data.available_endpoints?.forEach(endpoint => {
                console.log(`     - ${endpoint}`);
            });
        }
        
    } catch (error) {
        console.log(`   ❌ Error: ${error.message}`);
    }
    
    // Test 3: Test specific SSOT payment endpoints (should return 401/403)
    const endpoints = [
        'POST /payments/ssot/receivable',
        'POST /payments/ssot/payable',
        'POST /payments/ssot/preview-journal'
    ];
    
    console.log('\n3. Testing SSOT payment endpoints (expect 401/403):');
    
    for (const endpointDesc of endpoints) {
        const [method, path] = endpointDesc.split(' ');
        
        try {
            let response;
            if (method === 'POST') {
                response = await axios.post(`${API_BASE_URL}${path}`, {}, {
                    validateStatus: () => true
                });
            } else {
                response = await axios.get(`${API_BASE_URL}${path}`, {
                    validateStatus: () => true
                });
            }
            
            console.log(`   ${endpointDesc} - Status: ${response.status}`);
            
            if (response.status === 404) {
                console.log('     ❌ Endpoint not found');
            } else if (response.status === 401 || response.status === 403) {
                console.log('     ✅ Endpoint exists (requires authentication)');
            } else if (response.status === 400) {
                console.log('     ✅ Endpoint exists (bad request - expected without auth data)');
            }
            
        } catch (error) {
            console.log(`     ❌ Error: ${error.message}`);
        }
    }
    
    console.log('\n📋 Summary:');
    console.log('If you see status codes 401, 403, or 400 instead of 404, the endpoints exist.');
    console.log('The 404 error in frontend should now be fixed after updating payment service.\n');
}

// Run the test
testPaymentEndpoints().catch(console.error);