# Purchase Balance Migration Tester

This script helps you test and verify the Purchase Balance Migration system in your accounting application.

## What it does

- ✅ Tests database connectivity (PostgreSQL first, then MySQL fallback)
- 🔍 Checks if the migration has been applied
- 📊 Verifies installed functions and components
- 🧪 Tests function execution
- 📋 Validates prerequisites

## How to use

1. **Configure your .env file** in the backend directory:
   ```env
   # For PostgreSQL
   DATABASE_URL=postgres://username:password@localhost/accounting_db?sslmode=disable
   
   # Or for MySQL
   DATABASE_URL=mysql://username:password@localhost:3306/accounting_db
   ```

2. **Run the test**:
   ```bash
   cd backend
   go run cmd/scripts/test_migration_simple.go
   ```

3. **The script will automatically**:
   - Load database configuration from .env file
   - Connect to your database
   - Parse the DATABASE_URL to detect database type
   - Test migration status and functions

## Configuration

### .env File Setup
The script reads database configuration from your `.env` file. Make sure you have:

```env
# Database Configuration
DATABASE_URL=postgres://username:password@localhost/database_name?sslmode=disable

# Other configurations...
JWT_SECRET=your_jwt_secret
SERVER_PORT=8080
```

### Supported Database URLs
- **PostgreSQL**: `postgres://user:pass@host:port/dbname?sslmode=disable`
- **MySQL**: `mysql://user:pass@host:port/dbname`

### Security Features
- Password is automatically masked in output for security
- .env file is searched in current directory and parent directories
- Graceful fallback if .env file is not found

## Expected Output

### Before Migration
```
🧪 PURCHASE BALANCE MIGRATION TESTER
====================================
📄 Loading configuration from: D:\path\to\backend\.env
🔗 Database URL: postgres://postgres:***@localhost/accounting_db?sslmode=disable

🔌 Testing Database Connection...
   Database: accounting_db
   Host: localhost:5432
   User: postgres
✅ Database connection successful! (Using PostgreSQL)

🔍 Checking Migration Status...
⚠️  Purchase balance migration not found

📋 WHAT NEEDS TO BE DONE:
   1. Backend startup will automatically run migration
   2. Or manually run: 022_purchase_balance_validation_postgresql.sql
   3. Functions will be available after migration
```

### After Migration
```
🧪 PURCHASE BALANCE MIGRATION TESTER
====================================
📄 Loading configuration from: D:\path\to\backend\.env
🔗 Database URL: postgres://postgres:***@localhost/accounting_db?sslmode=disable

🔌 Testing Database Connection...
   Database: accounting_db
   Host: localhost:5432
   User: postgres
✅ Database connection successful! (Using PostgreSQL)

🔍 Checking Migration Status...
✅ Purchase Balance Migration already installed!

🔍 Checking installed components...
📊 Functions found: 3/3
   ✅ validate_purchase_balances
   ✅ sync_purchase_balances
   ✅ get_purchase_balance_status

🧪 Testing Functions...
   ✅ validate_purchase_balances() working
   ✅ sync_purchase_balances() working
   ✅ get_purchase_balance_status() working

🎯 SUMMARY:
   - Database: ✅ Connected
   - Migration: ✅ Installed
   - Functions: ✅ Available
   - Status: 🎉 Ready to use!
```

## Troubleshooting

### .env Configuration Issues
- Ensure `.env` file exists in backend directory
- Check DATABASE_URL format is correct
- Verify no extra spaces in DATABASE_URL
- Make sure database credentials are valid

### Connection Issues
- Make sure PostgreSQL or MySQL is running
- Check database credentials in .env file
- Verify the database exists
- Ensure user has proper permissions
- Check host and port are accessible

### Migration Issues
- Check if migration files exist in your migrations directory
- Verify migration_logs table exists
- Run migrations manually if needed

### Function Issues
- Ensure migration was applied successfully
- Check PostgreSQL logs for errors
- Verify database user has function execution permissions

## Files in this system

- `test_migration_simple.go` - This test script
- `022_purchase_balance_validation_postgresql.sql` - PostgreSQL-compatible migration
- `021_install_purchase_balance_validation.sql` - Original MySQL migration
- `purchase_balance_migration_guide.md` - Complete documentation

## Next Steps

After confirming everything works:
1. Apply the migration: `022_purchase_balance_validation_postgresql.sql`
2. Run this test script again to verify installation
3. Use the functions in your application code
4. Consider setting up automated testing