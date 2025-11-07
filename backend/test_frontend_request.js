// Simple Node.js script to mimic frontend request format

const testRequest = {
  account_codes: [],
  account_ids: [],  
  start_date: "2025-09-01",  // Frontend sends string
  end_date: "2025-09-17",    // Frontend sends string
  report_type: "PROFIT_LOSS",
  line_item_name: "Gross Profit", 
  page: 1,
  limit: 20
};

// Convert dates to RFC3339 like frontend should
const convertToRFC3339 = (dateString) => {
  if (!dateString) return new Date().toISOString();
  
  if (dateString.includes('T')) {
    return dateString;
  }
  
  const date = new Date(dateString + 'T00:00:00.000Z');
  return date.toISOString();
};

const convertedRequest = {
  ...testRequest,
  start_date: convertToRFC3339(testRequest.start_date),
  end_date: convertToRFC3339(testRequest.end_date)
};

console.log('Original request (what frontend might send):');
console.log(JSON.stringify(testRequest, null, 2));

console.log('\nConverted request (what should be sent):');
console.log(JSON.stringify(convertedRequest, null, 2));

// Test the conversion
async function testAPI() {
  try {
    const response = await fetch('http://localhost:8080/api/v1/journal-drilldown', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + process.env.TEST_TOKEN // You'd need to set this
      },
      body: JSON.stringify(convertedRequest)
    });
    
    console.log(`\nAPI Response Status: ${response.status}`);
    
    if (!response.ok) {
      const errorText = await response.text();
      console.log('Error response:', errorText);
    } else {
      const result = await response.json();
      console.log('Success! Entries found:', result.data?.journal_entries?.length || 0);
    }
  } catch (error) {
    console.error('Request failed:', error.message);
  }
}

// Uncomment to test (need to login first to get token)
// testAPI();