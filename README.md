# 📋 Aplikasi Sistem Akuntansi Modern

**Version 2.0** | **Last Updated: October 2025** | **Production Ready**

Sebuah aplikasi sistem akuntansi komprehensif yang menggabungkan backend API (Go) dan frontend web (Next.js) untuk mengelola seluruh aspek keuangan dan operasional bisnis modern. **Dilengkapi dengan Dark/Light Mode yang User-Friendly dan Multi-Language Support (Bahasa Indonesia & English)**.

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![Next.js](https://img.shields.io/badge/Next.js-15-black?logo=next.js)](https://nextjs.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-12+-336791?logo=postgresql)](https://postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ Key Features Terbaru

### 🎨 **User Experience Excellence**
- **🌓 Smart Dark/Light Mode** - Theme system yang responsif dengan deteksi sistem otomatis
- **🌍 Multi-Language Support** - Dukungan Bahasa Indonesia dan English dengan translation engine
- **⚡ Lightning Fast Performance** - Next.js 15 dengan Turbopack untuk development yang super cepat
- **📱 Responsive Design** - Mobile-first approach dengan Tailwind CSS + Chakra UI
- **🎭 Smooth Animations** - CSS variables dan transitions untuk UX yang premium

### 🔒 **Enterprise Security & Monitoring**
- **🛡️ Advanced Security Controller** - Real-time security incident monitoring dan management
- **📊 Balance Monitoring System** - Automated balance sync dan anomaly detection
- **🔍 Comprehensive Audit Trail** - Complete activity logging dengan forensic capabilities
- **⚠️ Smart Notifications** - Intelligent alert system dengan customizable rules
- **🔐 Session Management** - Auto-cleanup expired sessions dengan background worker
- **🎯 RBAC System** - 7-level role-based access control dengan granular permissions
- **🔄 Token Monitoring** - Advanced JWT token tracking dan auto-refresh

### 📈 **Enhanced Financial Reporting**
- **📋 Professional Financial Reports** - PDF/Excel export dengan formatting profesional
- **📊 Real-time Financial Dashboard** - Live metrics dan KPIs untuk decision making
- **🧮 Advanced Financial Ratios** - Automated calculation untuk analisis mendalam
- **📈 Unified Reporting Engine** - Standardized reporting dengan multiple output formats

## 🏗️ Arsitektur Aplikasi

```
app_sistem_akuntansi/
├── backend/                    # Go REST API dengan Gin & GORM
│   ├── cmd/                   # Entry point dan utilitas CLI
│   ├── controllers/           # HTTP handlers & enhanced security
│   ├── models/               # Database models & DTOs
│   ├── services/             # Business logic & advanced reporting
│   ├── repositories/         # Data access layer
│   ├── middleware/           # Auth, RBAC, enhanced security
│   ├── routes/               # API routing & unified endpoints
│   ├── migrations/           # Database migrations
│   ├── scripts/              # Maintenance & monitoring scripts
│   ├── docs/                 # API & system documentation
│   └── integration/          # Third-party integrations
├── frontend/                  # Next.js React App dengan Modern UI
│   ├── app/                  # Next.js 15 App Router
│   │   ├── globals.css       # Advanced theming dengan CSS variables
│   │   ├── layout.tsx        # Root layout dengan theme initialization
│   │   └── ClientProviders.tsx # Provider wrapper untuk contexts
│   ├── src/
│   │   ├── components/       # React components dengan theme support
│   │   │   ├── common/       # Reusable components
│   │   │   │   └── SimpleThemeToggle.tsx # Theme switcher
│   │   │   ├── reports/      # Enhanced reporting components
│   │   │   ├── settings/     # System configuration UI
│   │   │   └── users/        # User management dengan permissions
│   │   ├── contexts/         # React contexts
│   │   │   ├── SimpleThemeContext.tsx    # Theme management
│   │   │   ├── LanguageContext.tsx       # Multi-language support
│   │   │   └── AuthContext.tsx           # Authentication state
│   │   ├── hooks/            # Custom React hooks
│   │   │   ├── useTranslation.ts         # Translation hook
│   │   │   └── usePermissions.ts         # Permission management
│   │   ├── services/         # API services & financial reporting
│   │   ├── translations/     # Language files (ID/EN)
│   │   ├── utils/           # Helper functions
│   │   └── types/           # TypeScript definitions
│   └── public/              # Static assets
└── README.md
```

## 🚀 Stack Teknologi

### Backend (Go 1.23+)
- **Framework**: Gin Web Framework v1.10+ untuk REST API
- **Database**: PostgreSQL 12+ dengan GORM v1.30 ORM
- **Authentication**: JWT (golang-jwt/jwt v5.3) dengan refresh token mechanism
- **Enhanced Security**: Advanced security monitoring, incident tracking, audit logging
- **File Processing**: 
  - Excel export: excelize v2.9+
  - PDF generation: gofpdf v1.16+
  - Professional formatting templates
- **Reporting Engine**: Multi-format report generation dengan standardized templates
- **Decimal Handling**: shopspring/decimal v1.4 untuk presisi keuangan
- **Migration**: Auto-migration system dengan rollback support
- **Logging**: Structured logging dengan logrus v1.9
- **Middleware**: CORS, validation, security headers, rate limiting, RBAC
- **Architecture**: Clean Architecture dengan Repository Pattern
- **API Documentation**: Swagger/OpenAPI dengan swaggo v1.16

### Frontend (Next.js 15)
- **Framework**: Next.js 15 dengan App Router dan Turbopack
- **Language**: TypeScript untuk type safety
- **UI Components**: Chakra UI + Tailwind CSS + Radix UI
- **Theme System**: Advanced Dark/Light mode dengan CSS variables
- **Internationalization**: Multi-language support (ID/EN) dengan custom translation engine
- **State Management**: React Context + custom hooks
- **Charts**: Recharts untuk data visualization dengan theme-aware colors
- **Forms**: React Hook Form + Zod validation
- **HTTP Client**: Axios dengan interceptors dan error handling
- **Icons**: React Icons + Lucide React
- **Performance**: SSR/SSG optimization dengan hydration mismatch prevention

## 🌟 Comprehensive Feature Set

### 🎨 **Modern User Interface**
- **🌓 Intelligent Theme System**
  - Automatic dark/light mode detection based on system preference
  - Manual theme toggle dengan smooth transitions
  - CSS variables untuk consistent theming
  - Theme persistence dengan localStorage
  - Chakra UI integration untuk component theming

- **🌍 Multi-Language Support**
  - Complete Indonesian dan English translations
  - Context-based translation system
  - Real-time language switching
  - Nested translation keys dengan dot notation
  - Language preference persistence

- **📱 Responsive & Accessible**
  - Mobile-first design approach
  - Accessibility-compliant dengan ARIA standards
  - Keyboard navigation support
  - Screen reader compatible
  - High contrast mode support

### 🔐 Enhanced Security & Authentication
- **Multi-layer Authentication**: JWT + Refresh Token dengan auto-refresh
- **Role-Based Access Control (RBAC)**: 7 level user roles (Admin, Director, Finance, Inventory, Employee, Auditor, Operational)
- **Advanced Security Monitoring**: Real-time incident tracking dan threat detection
- **Security Controller**: Comprehensive security incident management system
- **Enhanced Middleware**: Rate limiting, CORS, audit logging, dan security headers
- **Token Monitoring**: Advanced session tracking dan security events
- **Password Security**: bcrypt hashing dengan advanced validation rules

### 👥 Enhanced User Management System
- **Admin**: Full system access + user management + security monitoring
- **Director**: Executive dashboard + comprehensive reporting + approval workflows
- **Finance**: Financial operations + advanced reporting + audit capabilities
- **Inventory Manager**: Stock management + product operations + supply chain analytics
- **Employee**: Basic operations + data entry + self-service features


### 📊 Core Business Modules

#### 💼 Sales Management
- **Multi-stage Sales Process**: Quotation → Order → Invoice → Payment
- **Advanced Calculations**: Multi-level discounts, PPN/PPh taxes
- **Payment Tracking**: Partial payments, receivables management
- **Customer Portal**: Sales history, invoice management
- **Returns & Refunds**: Full/partial returns dengan credit notes
- **Professional PDF Generation**: Industry-standard invoices dan reports

#### 🛒 Purchase Management
- **Procurement Workflow**: Request → Approval → Order → Receipt
- **Multi-level Approvals**: Configurable approval workflows
- **Vendor Management**: Supplier tracking, purchase history
- **Three-way Matching**: PO-Receipt-Invoice validation
- **Document Management**: Upload dan track purchase documents
- **Accounting Integration**: Automated journal entries

#### 📦 Inventory Control
- **Real-time Stock Tracking**: Multi-location inventory
- **Smart Notifications**: Minimum stock alerts dengan dashboard integration
- **Stock Operations**: Adjustments, transfers, opname
- **Valuation Methods**: FIFO, LIFO, Average costing
- **Product Variants**: Multiple SKUs per product
- **Bulk Operations**: Price updates, stock adjustments

#### 💰 Enhanced Financial Management
- **Chart of Accounts**: Hierarchical account structure
- **Cash & Bank Management**: Multi-account, transfers
- **Payment Processing**: Multiple payment methods
- **Tax Management**: PPN, PPh calculations
- **Advanced Financial Reports**: Professional P&L, Balance Sheet, Cash Flow
- **Balance Monitoring**: Automated balance sync dan anomaly detection
- **Journal Entry Management**: Manual journal entries dengan audit trail

#### 🏢 Asset Management
- **Fixed Asset Tracking**: Complete asset lifecycle
- **Depreciation Calculations**: Multiple methods
- **Asset Categories**: Organized asset classification
- **Maintenance Scheduling**: Asset maintenance tracking
- **Document Attachments**: Asset photos dan documents

#### 📈 Advanced Analytics & Reporting
- **Executive Dashboard**: Role-specific KPIs, trends, dan real-time analytics
- **Enhanced Financial Reports**: Professional-grade financial statements
- **Multiple Export Formats**: PDF, Excel, JSON dengan professional formatting
- **Real-time Metrics**: Live dashboard updates dengan WebSocket support
- **Advanced Filtering**: Multi-criteria search dengan saved filter profiles
- **Financial Ratios Calculator**: Automated calculation untuk liquidity, profitability, efficiency ratios
- **Unified Reporting Engine**: Standardized reporting framework untuk consistency
- **Balance Monitoring**: Automated balance reconciliation dan anomaly detection
- **Professional Report Templates**: Industry-standard formatting untuk compliance

## 🛠️ Quick Start

### Prerequisites
- **Node.js 18+** dengan npm/yarn
- **Go 1.23+** dengan module support
- **PostgreSQL 12+** atau MySQL 8+
- **Git** untuk version control

### 1. Clone Repository
```bash
git clone [repository-url]
cd sistem_akuntansi
```

### 2. Setup Backend

#### 2.1 Persiapan Environment

```bash
cd backend

# Install Go dependencies
go mod tidy

# Setup database PostgreSQL
createdb sistem_akuntansi
# atau untuk MySQL: CREATE DATABASE app_sistem_akuntansi;

# Copy dan konfigurasi environment variables
cp .env.example .env
```

#### 2.2 Konfigurasi `.env`

```env
# Database Configuration (pilih salah satu)
# Option 1: Full Connection String (Recommended)
DATABASE_URL=postgres://username:password@localhost:5432/sistem_akuntansi?sslmode=disable

# Option 2: Individual Parameters
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=sistem_akuntansi

# Application Configuration
SERVER_PORT=8080
ENVIRONMENT=development  # development | production

# Security
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=168h  # 7 days

# Optional: Swagger
ENABLE_SWAGGER=true  # Set false di production
```

#### 2.3 Jalankan Aplikasi

```bash
# Jalankan backend server
go run cmd/main.go

# ✅ Server akan OTOMATIS menjalankan:
# 1. ⚡ Connect ke database dan verify connection
# 2. 🗄️  Auto-migrate database schema (GORM models)
# 3. 🔄 Run SQL migrations (40+ migration files)
# 4. 🛡️  Install balance sync system (triggers + functions)
# 5. 📊 Install SSOT journal functions
# 6. ✅ Verify invoice types system
# 7. 📋 Setup approval workflows (Standard Purchase Approval)
# 8. 🌱 Seed initial data:
#    - Default users (admin, finance, director, employee, etc.)
#    - Chart of accounts (Indonesian standard)
#    - Product categories
#    - Default invoice types
# 9. 🔧 Fix account header status
# 10. 🚀 Start HTTP server on port 8080
# 11. 🧹 Start background session cleanup worker
# 12. 📚 Generate Swagger documentation (if enabled)

# 🎯 TIDAK perlu menjalankan script manual!
# Semua setup otomatis pada first run.
```

**Backend Server**: `http://localhost:8080`  
**Health Check**: `http://localhost:8080/api/v1/health`  
**API Documentation**: `http://localhost:8080/swagger/index.html` (jika ENABLE_SWAGGER=true)

#### 2.4 Troubleshooting

**Problem: UUID Extension Error**
```bash
# Jalankan fix script
go run apply_database_fixes.go
go run cmd/main.go
```

**Problem: Migration Failed**
```bash
# Check migration logs
psql -d sistem_akuntansi -c "SELECT * FROM migration_logs ORDER BY executed_at DESC LIMIT 10;"

# Re-run aplikasi (migrations are idempotent)
go run cmd/main.go
```

**Problem: Port Already in Use**
```bash
# Ubah port di .env
SERVER_PORT=8081
```

### 3. Setup Frontend
```bash
cd frontend

# Install Node.js dependencies
npm install
# atau
yarn install

# Setup environment variables (optional)
# Buat .env.local dan set NEXT_PUBLIC_API_URL jika berbeda dari default
echo "NEXT_PUBLIC_API_URL=http://localhost:8080/" > .env.local

# Jalankan development server dengan Turbopack
npm run dev
# atau
yarn dev
```
**Frontend Application**: `http://localhost:3000`

### 4. Default Login Credentials
```
Admin User:
  Username: admin@company.com
  Password: password123

Finance User:
  Username: finance@company.com
  Password: password123

Director User:
  Username: director@company.com
  Password: password123

Employee User:
  Username: employee@company.com
  Password: password123

Inventory User:
  Username: inventory@company.com
  Password: password123

Auditor User:
  Username: auditor@company.com
  Password: password123
```

## 📚 Comprehensive API Documentation

### Authentication & User Management
```http
POST /api/v1/auth/register     # Register user baru
POST /api/v1/auth/login        # Login user
POST /api/v1/auth/refresh      # Refresh JWT token
GET  /api/v1/profile           # Get user profile
PUT  /api/v1/profile           # Update user profile
```

### Sales Management
```http
GET    /api/v1/sales                      # List sales dengan filters
POST   /api/v1/sales                      # Create new sale
GET    /api/v1/sales/{id}                 # Get sale details
PUT    /api/v1/sales/{id}                 # Update sale
POST   /api/v1/sales/{id}/confirm         # Confirm sale
POST   /api/v1/sales/{id}/invoice         # Generate invoice
POST   /api/v1/sales/{id}/payments        # Record payment
GET    /api/v1/sales/analytics            # Sales analytics
GET    /api/v1/sales/{id}/invoice/pdf     # Export invoice PDF
```

### Purchase Management
```http
GET    /api/v1/purchases                   # List purchases
POST   /api/v1/purchases                   # Create purchase
POST   /api/v1/purchases/{id}/submit-approval  # Submit for approval
POST   /api/v1/purchases/{id}/approve      # Approve purchase
GET    /api/v1/purchases/pending-approval  # Get pending approvals
POST   /api/v1/purchases/receipts          # Create receipt
```

### Inventory & Products
```http
GET    /api/v1/products                    # List products
POST   /api/v1/products                    # Create product
POST   /api/v1/products/adjust-stock       # Adjust stock
POST   /api/v1/products/opname            # Stock opname
GET    /api/v1/inventory/movements         # Stock movements
GET    /api/v1/inventory/low-stock         # Low stock alerts
```

### Financial Management
```http
GET    /api/v1/accounts                    # Chart of accounts
GET    /api/v1/cash-banks                  # Cash & bank accounts
POST   /api/v1/payments                    # Record payments
GET    /api/v1/payments/dashboard          # Payment dashboard
POST   /api/v1/cash-banks/transfer         # Bank transfers
POST   /api/v1/journal-entries             # Manual journal entries
```

### Enhanced Reports & Analytics
```http
GET    /api/v1/reports/sales               # Sales reports
GET    /api/v1/reports/purchases           # Purchase reports
GET    /api/v1/reports/inventory           # Inventory reports
GET    /api/v1/reports/financial           # Financial reports
GET    /api/v1/dashboard/summary           # Dashboard data

# Enhanced Financial Reporting
GET    /api/v1/enhanced-reports/balance-sheet    # Comprehensive balance sheet
GET    /api/v1/enhanced-reports/profit-loss     # Enhanced P&L statement
GET    /api/v1/enhanced-reports/cash-flow       # Cash flow statement
POST   /api/v1/financial-reports/trial-balance  # Generate trial balance
GET    /api/v1/financial-reports/general-ledger/{account_id} # General ledger by account
GET    /api/v1/financial-reports/dashboard      # Financial dashboard
GET    /api/v1/financial-reports/metrics        # Real-time financial metrics
GET    /api/v1/financial-reports/ratios         # Calculate financial ratios

# Unified Reporting System
GET    /api/v1/unified-reports/comprehensive    # Multi-format comprehensive reports
POST   /api/v1/unified-reports/custom          # Custom report generation
```

### Enhanced System Monitoring & Security
```http
GET    /api/v1/monitoring/status           # System status
GET    /api/v1/monitoring/audit-logs       # Audit trails
GET    /api/v1/notifications               # System notifications
GET    /api/v1/health                      # Health check

# Security Management
GET    /api/v1/admin/security/incidents    # List security incidents
GET    /api/v1/admin/security/incidents/{id} # Get incident details
PUT    /api/v1/admin/security/incidents/{id}/resolve # Resolve incident
GET    /api/v1/security/dashboard          # Security monitoring dashboard
POST   /api/v1/security/report-incident    # Report security incident

# Balance Monitoring
GET    /api/v1/balance-monitor/status      # Balance monitoring status
POST   /api/v1/balance-monitor/sync        # Manual balance synchronization
GET    /api/v1/balance-monitor/anomalies   # Detect balance anomalies

# Settings Management
GET    /api/v1/settings                    # Get system settings
PUT    /api/v1/settings                    # Update system settings
GET    /api/v1/settings/company            # Company information
PUT    /api/v1/settings/company            # Update company info
```

## 🗃️ Enhanced Database Schema

### Core Business Tables
- **users** - User authentication, roles, permissions, dan profile
- **contacts** - Customers, vendors, employees, sales persons
- **products** - Master products dengan variants dan advanced tracking
- **product_categories** - Hierarchical product categorization
- **accounts** - Enhanced Chart of accounts dengan hierarchy

### Transaction Tables
- **sales** & **sale_items** - Sales transactions (quotation→invoice→payment)
- **sale_payments** & **sale_returns** - Payment tracking & returns
- **purchases** & **purchase_items** - Purchase transactions
- **purchase_receipts** - Goods receipt tracking
- **inventories** - Stock movement logs dengan real-time tracking
- **cash_banks** - Bank accounts dan cash management
- **payments** - Universal payment records

### Enhanced System Tables
- **approval_workflows** - Configurable approval processes
- **notifications** - System notifications dan smart alerts
- **notification_configs** - User notification preferences
- **audit_logs** - Complete audit trail dengan forensic capabilities
- **security_incidents** - Security monitoring dan incident tracking
- **assets** - Fixed asset management dengan depreciation
- **stock_alerts** - Advanced minimum stock monitoring
- **settings** - System configuration dan company information
- **permissions** - Granular permission management
- **financial_reports** - Cached financial report data
- **journal_entries** - Manual journal entry tracking

## 🔧 Development Guide

### Backend Development
```bash
cd backend

# Development mode dengan auto-reload
go run cmd/main.go

# Run specific maintenance scripts
go run scripts/maintenance/fix_accounts.go
go run scripts/maintenance/check_sales_codes.go

# Security and balance monitoring scripts
go run scripts/test_security_system.go
go run scripts/maintenance/run_balance_monitor.go

# Database operations
go run scripts/maintenance/reset_transaction_data.go
go run scripts/maintenance/sync_cashbank_gl_balance.go

# Build for production
go build -o app cmd/main.go
```

### Frontend Development
```bash
cd frontend

# Development dengan Turbopack (faster)
npm run dev

# Type checking
npx tsc --noEmit

# Linting
npm run lint

# Production build
npm run build
npm run start
```

### Development Features
- **Hot Reload**: Backend dan frontend auto-refresh
- **TypeScript**: Full type safety dengan strict mode
- **Theme Development**: Live theme switching untuk development
- **Multi-language Testing**: Real-time language switching
- **API Interceptors**: Auto token refresh
- **Error Boundaries**: Comprehensive error handling
- **Debug Routes**: `/api/v1/debug/*` untuk testing

## 📦 Production Deployment

### Backend Deployment
```bash
# Build production binary
go build -o sistem-akuntansi cmd/main.go

# Setup production database
createdb sistem_akuntansi_prod

# Set production environment
export DB_HOST=prod-db-host
export DB_NAME=sistem_akuntansi_prod
export JWT_SECRET=your-secure-jwt-secret
export GIN_MODE=release

# Run application
./sistem-akuntansi
```

### Frontend Deployment
```bash
# Build untuk production
npm run build

# Deploy ke Vercel (recommended)
npx vercel --prod

# Atau deploy ke server dengan PM2
npm install -g pm2
pm2 start npm --name "sistem-akuntansi" -- start

# Environment variables untuk production
NEXT_PUBLIC_API_URL=https://your-api-domain.com/
```

### Enhanced Security Checklist
- [ ] Update JWT_SECRET untuk production
- [ ] Enable HTTPS untuk API dan frontend
- [ ] Configure database SSL
- [ ] Set proper CORS origins
- [ ] Enable advanced rate limiting
- [ ] Setup security monitoring alerts
- [ ] Configure balance monitoring notifications
- [ ] Enable audit log retention policies
- [ ] Setup incident response procedures

## ✅ Implementation Status

### ✅ Completed Features (Production Ready)

#### 🔐 Authentication & Security
- [x] **JWT Authentication** - JWT v5.3 + Refresh tokens dengan auto-refresh
- [x] **RBAC System** - 7 user roles (Admin, Director, Finance, Inventory, Employee, Auditor, Operational)
- [x] **Session Management** - Auto-cleanup dengan background worker
- [x] **Token Monitoring** - Advanced JWT tracking dan security events
- [x] **Security Incidents** - Real-time monitoring dan incident management
- [x] **Audit Trail** - Forensic-level activity logging
- [x] **Password Security** - bcrypt hashing dengan validation rules

#### 📊 Core Business Modules
- [x] **Sales Module** - End-to-end sales process (Quotation → Invoice → Payment)
- [x] **Purchase Module** - Full procurement workflow dengan multi-level approvals
- [x] **Inventory Management** - Real-time stock tracking dengan smart notifications
- [x] **Financial Management** - Chart of accounts, journal entries, balance monitoring
- [x] **Asset Management** - Fixed asset tracking dengan depreciation calculations
- [x] **Cash & Bank** - Multi-account management, transfers

#### 📈 Reporting & Analytics
- [x] **Financial Reports** - P&L, Balance Sheet, Cash Flow dengan PDF/Excel export
- [x] **Role-Specific Dashboards** - Personalized analytics untuk setiap user role
- [x] **Financial Ratio Calculator** - Automated liquidity, profitability, efficiency analysis
- [x] **Unified Reporting Engine** - Standardized framework dengan multiple formats
- [x] **Invoice Types System** - Customizable invoice numbering dengan auto-counter

#### 🎨 UI/UX Features
- [x] **Dark/Light Theme** - Smart detection dengan smooth transitions
- [x] **Multi-Language** - Complete ID/EN translation system
- [x] **Responsive Design** - Mobile-first dengan Tailwind CSS + Chakra UI
- [x] **Smart Notifications** - Context-aware alerts dengan user preferences

#### 🔧 Developer Features
- [x] **Auto-Migration System** - Zero-manual-setup database initialization
- [x] **Swagger Documentation** - Auto-generated API docs dengan swaggo
- [x] **Debug Endpoints** - Development testing tools
- [x] **Migration Logs** - Complete database change tracking
- [x] **Balance Sync System** - Automated balance reconciliation dengan triggers
- [x] **SSOT Journal** - Single Source of Truth accounting system

### 🚧 Roadmap

- [ ] **Comprehensive Unit Testing** - Backend dan frontend test coverage (Target: 80%+)
- [ ] **Performance Optimization** - Database query optimization, Redis caching
- [ ] **Advanced Permissions** - Field-level access control
- [ ] **Export Templates** - Customizable report templates
- [ ] **Budget Module** - Budget planning dan tracking
- [ ] **Multi-Company** - Multi-entity support
- [ ] **API Rate Limiting** - Per-user rate limiting
- [ ] **Webhook Integration** - External system notifications

## 🎯 Key Highlights

### 🏆 Production Ready
- **Comprehensive Business Logic** - Real-world accounting principles dengan industry standards
- **Enterprise Security** - Multi-layer security dengan forensic audit trails
- **Scalable Architecture** - Clean architecture dengan separation of concerns
- **Modern Tech Stack** - Latest versions dengan best practices
- **Responsive Design** - Mobile-first UI/UX dengan accessibility compliance
- **User-Friendly Interface** - Intuitive dark/light theme dengan smooth transitions
- **Multi-Language Ready** - Complete internationalization support

### 📊 Business Impact
- **Streamlined Operations** - Integrated sales-to-payment workflow dengan automation
- **Real-time Financial Control** - Live financial visibility dengan advanced analytics
- **Smart Inventory Optimization** - AI-powered stock management dengan predictive alerts
- **Regulatory Compliance** - Indonesian tax regulations (PPN/PPh) dengan automatic updates
- **Data-Driven Decision Support** - Executive dashboards dengan actionable insights
- **Enhanced User Experience** - Intuitive dark/light theme dengan multi-language support
- **Enterprise Security** - Advanced threat detection dan incident management
- **Professional Reporting** - Industry-standard financial statements dan analysis

## 🤝 Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Development Standards
- Follow clean code principles
- Add proper error handling
- Include type definitions untuk TypeScript
- Write descriptive commit messages
- Update documentation dan translations
- Test dalam dark/light theme
- Verify multi-language functionality
- Include security considerations

## 📞 Support & Documentation

- **Issues**: Create GitHub issue untuk bugs atau feature requests
- **Documentation**: Lihat folder `backend/docs/` untuk detailed technical documentation
- **API Testing**: Gunakan `/api/v1/debug/` endpoints untuk development testing
- **Security Reports**: Use secure channels untuk reporting security vulnerabilities
- **Feature Requests**: Submit detailed requirements dengan business justification

## 🔥 What's New in Version 2.0 (October 2025)

### ⚡ **Performance & Stability**
- 🚀 **Auto-Migration System** - Zero-manual-setup database initialization
- 🔄 **Idempotent Migrations** - Safe to re-run, smart duplicate detection
- 🏃 **Background Workers** - Session cleanup, balance sync automation
- 📊 **40+ SQL Migrations** - Comprehensive database schema management
- 🎯 **SSOT Journal System** - Single Source of Truth untuk accounting
- 💾 **Decimal Precision** - shopspring/decimal untuk accurate financial calculations

### 🎨 **UI/UX Enhancements**
- ✨ **Smart Dark/Light Theme** - Automatic theme detection dengan smooth transitions
- 🌍 **Multi-Language Support** - Complete ID/EN translations dengan real-time switching
- 📱 **Enhanced Mobile Experience** - Improved responsive design
- 🎭 **Smooth Animations** - CSS-based transitions untuk professional feel
- 🖼️ **Modern UI Components** - Chakra UI + Tailwind CSS + Radix UI

### 🔒 **Security Improvements**
- 🛡️ **Security Monitoring Dashboard** - Real-time incident tracking
- 📊 **Balance Monitoring System** - Automated reconciliation dengan anomaly detection
- 🔍 **Enhanced Audit Trail** - Forensic-level logging capabilities
- ⚠️ **Smart Alert System** - Context-aware notifications
- 🔐 **Token Monitoring** - JWT session tracking dengan auto-refresh
- 🧹 **Session Cleanup** - Auto-purge expired sessions

### 📈 **Advanced Reporting**
- 📋 **Professional Report Templates** - Industry-standard formatting
- 📊 **Real-time Financial Metrics** - Live dashboard updates
- 🧮 **Financial Ratio Calculator** - Automated analysis tools
- 📈 **Unified Reporting Engine** - Standardized framework
- 📄 **Excel & PDF Export** - excelize v2.9 & gofpdf v1.16
- 📊 **Invoice Types System** - Customizable invoice numbering

### 🔧 **Developer Experience**
- 📚 **Swagger Documentation** - Auto-generated API docs
- 🧪 **Testing Tools** - Comprehensive test scripts
- 🐛 **Debug Endpoints** - `/api/v1/debug/*` untuk development
- 📝 **Migration Logs** - Track all database changes
- 🔍 **Enhanced Error Messages** - Detailed error context

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**🚀 Built with cutting-edge technologies for modern business management**  
*Sistem Akuntansi Modern - Complete Enterprise Solution with Dark/Light Theme & Multi-Language Support*

**Latest Features**: Dark/Light Mode • Multi-Language (ID/EN) • Enhanced Security • Advanced Reporting • Balance Monitoring • Professional UI/UX
