# 🧪 Comprehensive Sales-Payment Flow Testing Suite

## 📋 Overview

This testing suite provides comprehensive end-to-end testing for the sales-to-payment flow in the accounting system. It verifies 100% data integrity across all accounting transactions, ensuring that:

- **Piutang Usaha** (Accounts Receivable) decreases when payment is received
- **Kas/Bank** (Cash/Bank) increases when payment is recorded
- **Pendapatan** (Revenue) is properly recorded during sales
- **Journal entries** are created correctly for all transactions
- **Data consistency** is maintained across all modules

## 🚀 Quick Start

### Windows (PowerShell)
```powershell
# Simple test run
.\scripts\run_test.ps1

# Verbose output with server wait
.\scripts\run_test.ps1 -Verbose -WaitForServer

# Custom server URL
.\scripts\run_test.ps1 -ServerURL http://localhost:8081
```

### Windows (Command Prompt)
```cmd
REM Quick test
scripts\quick_test.bat
```

### Linux/Mac (Bash)
```bash
# Make script executable (first time only)
chmod +x scripts/run_test.sh

# Simple test run
./scripts/run_test.sh

# Verbose output with server wait
./scripts/run_test.sh --verbose --wait-server

# Custom server URL
./scripts/run_test.sh --url http://localhost:8081
```

### Manual Go Execution
```bash
# Direct execution
go run scripts/test_sales_payment_flow.go
```

## 📁 Files Structure

```
scripts/
├── test_sales_payment_flow.go  # Main test implementation
├── run_test.ps1               # PowerShell runner (Windows)
├── run_test.sh                # Bash runner (Linux/Mac)
├── quick_test.bat             # Quick Windows batch runner
└── README.md                  # This documentation
```

## 🔧 Test Configuration

### Environment Variables
```bash
export TEST_BASE_URL="http://localhost:8080/api/v1"
```

### Test Data Constants (in test_sales_payment_flow.go)
```go
const (
    TestCustomerID = 1        # Customer to use for testing
    TestProductID  = 1        # Product to use for sales
    TestCashBankID = 1        # Cash/Bank account for payments
    
    // Account codes to monitor
    ReceivableAccount = "1201" # Piutang Usaha
    CashAccount      = "1101" # Kas
    BankAccount      = "1102" # Bank BCA  
    RevenueAccount   = "4101" # Pendapatan Penjualan
)
```

## 🧪 Test Scenarios

### 1. Complete Sales-to-Payment Flow
1. **Sales Creation**: Creates a sales order with 2 products @ 1M each (Total: 2.22M including PPN)
2. **Invoice Generation**: Converts sales order to invoice
3. **Payment Recording**: Records payment allocation against the invoice
4. **Balance Verification**: Verifies all account balances updated correctly
5. **Integrity Check**: Ensures data consistency across all modules

### 2. Account Balance Verification
The test monitors these critical account changes:

```
📈 Before → After (Expected Changes)
────────────────────────────────────
Piutang Usaha (1201): +2,220,000 → 0       (decrease by payment)
Kas (1101):           +2,220,000            (increase by payment)
Pendapatan (4101):    +2,220,000            (increase by sale)
```

### 3. Journal Entry Validation
- Sales journal entries (DR: Piutang, CR: Pendapatan + PPN)
- Payment journal entries (DR: Kas, CR: Piutang)
- Balanced debit/credit amounts
- Proper reference linking

## 📊 Test Report Example

```
======================================================================
📋 COMPREHENSIVE TEST REPORT
======================================================================
⏱️  Test Duration: 3.45s
📅 Start Time: 2025-01-15 10:30:00
📅 End Time: 2025-01-15 10:30:03

🎯 TEST RESULTS:
   ✅ Sales Created
   ✅ Sales Invoiced
   ✅ Payment Recorded
   ✅ Journal Entries Created
   ✅ Accounts Updated
   ✅ Data Integrity Verified

💰 FINANCIAL SUMMARY:
   📊 Sale Amount: 2220000.00
   💸 Payment Amount: 2220000.00

📈 BALANCE CHANGES:
   📈 1201: 0.00 -> 2220000.00 (Change: +2220000.00)
   📈 1101: 5000000.00 -> 7220000.00 (Change: +2220000.00)
   📈 4101: 4000000.00 -> 6000000.00 (Change: +2000000.00)

📊 JOURNAL ENTRIES: 4 entries recorded

======================================================================
🎉 OVERALL RESULT: ✅ ALL TESTS PASSED - 100% SUCCESS!
   🔹 Sales-to-Payment flow working perfectly
   🔹 Account balances updated correctly
   🔹 Data integrity maintained
   🔹 Journal entries properly recorded
   🔹 System is production-ready! 🚀
======================================================================
```

## 🛠️ Prerequisites

### Required Software
- **Go 1.19+** - Programming language runtime
- **curl** - For server connectivity checks (Linux/Mac)
- **PowerShell** - For Windows scripts

### Required Data
Ensure your database has:
1. **Customer record** with ID = 1
2. **Product record** with ID = 1
3. **Cash/Bank accounts** with proper setup
4. **Chart of Accounts** with standard account codes:
   - 1201: Piutang Usaha
   - 1101: Kas  
   - 1102: Bank BCA
   - 4101: Pendapatan Penjualan

### Server Requirements
- Server running on `localhost:8080` (or specify custom URL)
- Database connected and migrated
- All API endpoints functional

## 🔍 Troubleshooting

### Common Issues

#### 1. Server Not Running
```bash
❌ Cannot connect to server at http://localhost:8080
```
**Solution**: Start the server with `go run cmd/main.go`

#### 2. Missing Test Data
```bash
❌ Customer not found: contact with ID 1 not found
```
**Solution**: Ensure test customers and products exist in database

#### 3. Account Code Mismatch
```bash
❌ Account not found: 1201
```
**Solution**: Verify chart of accounts has required account codes

#### 4. Database Connection Issues
```bash
❌ Database connection failed
```
**Solution**: Check database connectivity and credentials

### Debug Mode

Enable verbose logging:
```powershell
# PowerShell
.\scripts\run_test.ps1 -Verbose

# Bash
./scripts/run_test.sh --verbose
```

### Manual Test Execution
```bash
# Run test directly with Go
go run scripts/test_sales_payment_flow.go

# Check server health manually
curl http://localhost:8080/api/v1/health

# View account balance manually
curl http://localhost:8080/api/v1/accounts/1201
```

## 🔒 Data Safety

### Production Safety Features
- **Read-mostly operations**: Test primarily reads data and creates test transactions
- **Isolated test data**: Uses specific test customer/product IDs
- **Transaction rollback**: Failed tests don't leave partial data
- **Balance preservation**: Critical account balances are monitored and reported

### Recommended Usage
- **Development**: Safe to run anytime
- **Staging**: Excellent for CI/CD pipeline integration
- **Production**: Use with caution - creates real transactions

## 🚦 CI/CD Integration

### GitHub Actions Example
```yaml
name: Sales-Payment Integration Test

on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.19'
      
      - name: Start Server
        run: |
          go run cmd/main.go &
          sleep 10
      
      - name: Run Integration Tests
        run: ./scripts/run_test.sh --wait-server
```

### Docker Integration
```dockerfile
# Add to your Dockerfile
COPY scripts/ /app/scripts/
RUN chmod +x /app/scripts/run_test.sh

# Health check
HEALTHCHECK --interval=30s --timeout=10s \
  CMD /app/scripts/run_test.sh || exit 1
```

## 📈 Performance Benchmarks

### Expected Performance
- **Test Duration**: < 5 seconds
- **Database Operations**: ~20-30 queries
- **Memory Usage**: < 50MB
- **API Calls**: ~15-20 requests

### Performance Monitoring
The test automatically tracks:
- Total execution time
- Individual step timing
- API response times
- Database query performance

## 🔄 Extending the Tests

### Adding New Test Scenarios
1. **Edit** `test_sales_payment_flow.go`
2. **Add** new test functions following the pattern
3. **Update** `runComprehensiveTest()` to include new tests
4. **Test** your changes with `go run scripts/test_sales_payment_flow.go`

### Custom Account Testing
```go
// Add custom accounts to monitor
const (
    CustomAccount = "2101"  // Your custom account
)

// Add to captureInitialBalances function
accounts := []string{
    ReceivableAccount, 
    CashAccount, 
    CustomAccount,  // Add here
}
```

## 📞 Support

### Getting Help
1. **Check logs** for detailed error messages
2. **Verify prerequisites** are met
3. **Test server connectivity** manually
4. **Check database** for required test data

### Contributing
1. Fork the repository
2. Create feature branch
3. Add your test improvements
4. Run existing tests to ensure compatibility
5. Submit pull request with detailed description

---

**Last Updated**: January 2025  
**Version**: 1.0.0  
**Compatibility**: Go 1.19+, PostgreSQL, Windows/Linux/Mac