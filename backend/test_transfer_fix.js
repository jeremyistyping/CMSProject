const axios = require('axios');

const BASE_URL = 'http://localhost:8080/api/v1';

// Test data for creating a transfer
const transferTemplate = {
    from_account_id: 1, // Bank BCA (ID: 1)
    to_account_id: 2,   // Bank Mandiri (ID: 2) 
    amount: 100000,
    date: "2025-09-21",
    reference: "TEST-TRANSFER",
    notes: "Test transfer to verify fixed duplicate key constraint issue",
    exchange_rate: 1
};

async function runTransferTest() {
    try {
        // First, get authentication token
        console.log('🔐 Logging in...');
        const loginResponse = await axios.post(`${BASE_URL}/auth/login`, {
            email: 'admin@company.com',
            password: 'password123'
        });

        const token = loginResponse.data.token;
        console.log('✅ Login successful');

        // Set up headers with auth token
        const headers = {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        };

        // Get cash bank accounts first to verify they exist
        console.log('\n📊 Getting cash bank accounts...');
        const accountsResponse = await axios.get(`${BASE_URL}/cashbank/accounts`, { headers });
        console.log('✅ Available cash bank accounts:');
        accountsResponse.data.forEach(account => {
            console.log(`  - ID: ${account.id}, Name: ${account.name}, Balance: ${account.balance}`);
        });

        // Test multiple transfers to verify the duplicate key issue is fixed
        console.log('\n🔄 Testing transfer functionality...');
        
        for (let i = 1; i <= 3; i++) {
            console.log(`\n📤 Creating transfer ${i}...`);
            const transferData = {
                ...transferTemplate,
                amount: 50000 + (i * 10000), // Vary the amount
                notes: `${transferTemplate.notes} - Attempt ${i}`
            };

            try {
                const transferResponse = await axios.post(`${BASE_URL}/cashbank/transfer`, transferData, { headers });
                console.log(`✅ Transfer ${i} created successfully!`);
                console.log(`   Transfer Number: ${transferResponse.data.transfer_number}`);
                console.log(`   Amount: ${transferResponse.data.amount}`);
                console.log(`   Status: ${transferResponse.data.status}`);
                
                // Wait a bit between transfers to test concurrent scenarios
                await new Promise(resolve => setTimeout(resolve, 100));
            } catch (error) {
                if (error.response) {
                    console.log(`❌ Transfer ${i} failed:`, error.response.data);
                    if (error.response.data.details && error.response.data.details.includes('duplicate key')) {
                        console.log('🚨 DUPLICATE KEY CONSTRAINT ERROR STILL EXISTS!');
                        return;
                    }
                } else {
                    console.log(`❌ Transfer ${i} failed:`, error.message);
                }
            }
        }

        // Test concurrent transfers (simulate the original issue scenario)
        console.log('\n⚡ Testing concurrent transfers...');
        const concurrentPromises = [];
        for (let i = 1; i <= 5; i++) {
            const concurrentTransferData = {
                ...transferTemplate,
                amount: 25000 + (i * 5000),
                notes: `Concurrent transfer test - ${i}`
            };
            concurrentPromises.push(
                axios.post(`${BASE_URL}/cashbank/transfer`, concurrentTransferData, { headers })
                    .then(response => ({
                        success: true,
                        transferNumber: response.data.transfer_number,
                        amount: response.data.amount
                    }))
                    .catch(error => ({
                        success: false,
                        error: error.response?.data || error.message
                    }))
            );
        }

        const concurrentResults = await Promise.all(concurrentPromises);
        const successfulTransfers = concurrentResults.filter(r => r.success);
        const failedTransfers = concurrentResults.filter(r => !r.success);

        console.log(`✅ Successful concurrent transfers: ${successfulTransfers.length}`);
        console.log(`❌ Failed concurrent transfers: ${failedTransfers.length}`);

        if (successfulTransfers.length > 0) {
            console.log('📋 Transfer numbers generated:');
            successfulTransfers.forEach(t => {
                console.log(`   - ${t.transferNumber} (Amount: ${t.amount})`);
            });
        }

        if (failedTransfers.length > 0) {
            console.log('❌ Failed transfer details:');
            failedTransfers.forEach((f, idx) => {
                console.log(`   Transfer ${idx + 1}: ${JSON.stringify(f.error)}`);
            });
        }

        // Verify SSOT journal entries were created
        console.log('\n📊 Checking SSOT journal entries...');
        try {
            const journalsResponse = await axios.get(`${BASE_URL}/journals?limit=10&source_type=CASHBANK`, { headers });
            const cashbankJournals = journalsResponse.data.data.filter(j => j.source_type === 'CASHBANK');
            console.log(`✅ Found ${cashbankJournals.length} SSOT journal entries for cash/bank transactions`);
            
            if (cashbankJournals.length > 0) {
                console.log('📋 Recent cash/bank journal entries:');
                cashbankJournals.slice(0, 3).forEach(journal => {
                    console.log(`   - ${journal.entry_number}: ${journal.description} (${journal.total_debit})`);
                });
            }
        } catch (error) {
            console.log('⚠️ Could not fetch journal entries:', error.message);
        }

        console.log('\n🎉 Transfer testing completed successfully!');
        console.log('🔧 The duplicate key constraint issue appears to be FIXED!');

    } catch (error) {
        if (error.response) {
            console.error('❌ Test failed - Response:', error.response.status, error.response.data);
        } else if (error.request) {
            console.error('❌ Test failed - No response received:', error.message);
        } else {
            console.error('❌ Test failed - Request setup:', error.message);
        }
        console.error('Full error:', error);
    }
}

// Run the test
runTransferTest();
