# 📊 Aplikasi Sistem Akuntansi Modern

Sebuah aplikasi sistem akuntansi komprehensif yang menggabungkan backend API (Go) dan frontend web (Next.js) untuk mengelola seluruh aspek keuangan dan operasional bisnis modern.

## 🏗️ Arsitektur Aplikasi

```
app_sistem_akuntansi/
├── backend/                    # Go REST API dengan Gin & GORM
│   ├── cmd/                   # Entry point dan utilitas CLI
│   ├── controllers/           # HTTP handlers
│   ├── models/               # Database models dan DTOs
│   ├── services/             # Business logic layer
│   ├── repositories/         # Data access layer
│   ├── middleware/           # Auth, RBAC, security
│   ├── routes/               # API routing
│   ├── migrations/           # Database migrations
│   ├── scripts/              # Maintenance scripts
│   └── docs/                 # API documentation
├── frontend/                  # Next.js React App dengan Chakra UI & Tailwind
│   ├── app/                  # Next.js 15 App Router
│   ├── src/
│   │   ├── components/       # React components
│   │   ├── services/         # API services
│   │   ├── hooks/           # Custom React hooks
│   │   ├── contexts/        # React contexts
│   │   ├── utils/           # Helper functions
│   │   └── types/           # TypeScript definitions
│   └── public/              # Static assets
└── README.md
```

## 🚀 Stack Teknologi

### Backend (Go 1.23+)
- **Framework**: Gin Web Framework untuk REST API
- **Database**: PostgreSQL dengan GORM ORM
- **Authentication**: JWT dengan refresh token mechanism
- **Security**: bcrypt, RBAC, rate limiting, audit logging
- **File Processing**: Excel/PDF export dengan excelize & gofpdf
- **Middleware**: CORS, validation, security headers
- **Architecture**: Clean Architecture dengan Repository Pattern

### Frontend (Next.js 15)
- **Framework**: Next.js 15 dengan App Router
- **Language**: TypeScript untuk type safety
- **UI Components**: Chakra UI + Tailwind CSS + Radix UI
- **State Management**: React Context + custom hooks
- **Charts**: Recharts untuk data visualization
- **Forms**: React Hook Form + Zod validation
- **HTTP Client**: Axios dengan interceptors
- **Icons**: React Icons + Lucide React

## 🌟 Fitur Komprehensif

### 🔐 Security & Authentication
- **Multi-layer Authentication**: JWT + Refresh Token dengan auto-refresh
- **Role-Based Access Control (RBAC)**: 5 level user roles
- **Security Middleware**: Rate limiting, CORS, audit logging
- **Token Monitoring**: Track login sessions dan security events
- **Password Security**: bcrypt hashing dengan validation

### 👥 User Management System
- **Admin**: Full system access + user management
- **Director**: Executive dashboard + reporting access
- **Finance**: Financial operations + accounting
- **Inventory Manager**: Stock management + product operations
- **Employee**: Basic operations + data entry

### 📊 Core Business Modules

#### 💼 Sales Management
- **Multi-stage Sales Process**: Quotation → Order → Invoice → Payment
- **Advanced Calculations**: Multi-level discounts, PPN/PPh taxes
- **Payment Tracking**: Partial payments, receivables management
- **Customer Portal**: Sales history, invoice management
- **Returns & Refunds**: Full/partial returns dengan credit notes
- **PDF Generation**: Professional invoices dan reports

#### 🛒 Purchase Management
- **Procurement Workflow**: Request → Approval → Order → Receipt
- **Multi-level Approvals**: Configurable approval workflows
- **Vendor Management**: Supplier tracking, purchase history
- **Three-way Matching**: PO-Receipt-Invoice validation
- **Document Management**: Upload dan track purchase documents
- **Accounting Integration**: Auto journal entries

#### 📦 Inventory Control
- **Real-time Stock Tracking**: Multi-location inventory
- **Smart Notifications**: Minimum stock alerts dengan dashboard integration
- **Stock Operations**: Adjustments, transfers, opname
- **Valuation Methods**: FIFO, LIFO, Average costing
- **Product Variants**: Multiple SKUs per product
- **Bulk Operations**: Price updates, stock adjustments

#### 💰 Financial Management
- **Chart of Accounts**: Hierarchical account structure
- **Cash & Bank Management**: Multi-account, transfers, reconciliation
- **Payment Processing**: Multiple payment methods
- **Tax Management**: PPN, PPh calculations
- **Financial Reports**: P&L, Balance Sheet, Cash Flow
- **Multi-currency Support**: Exchange rates, currency conversion

#### 🏢 Asset Management
- **Fixed Asset Tracking**: Complete asset lifecycle
- **Depreciation Calculations**: Multiple methods
- **Asset Categories**: Organized asset classification
- **Maintenance Scheduling**: Asset maintenance tracking
- **Document Attachments**: Asset photos dan documents

#### 📈 Analytics & Reporting
- **Executive Dashboard**: KPIs, trends, analytics untuk setiap role
- **Financial Reports**: Comprehensive financial reporting
- **Export Capabilities**: PDF, Excel export untuk semua reports
- **Real-time Data**: Live dashboard updates
- **Custom Filters**: Advanced filtering dan search

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
```bash
cd backend

# Install Go dependencies
go mod tidy

# Setup database PostgreSQL
created sistem_akuntansi
# atau untuk MySQL, buat database: CREATE DATABASE app_sistem_akuntansi;

# Copy dan konfigurasi environment variables
cp .env.example .env
# Edit .env dengan konfigurasi database Anda:
# DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME, JWT_SECRET

# Jalankan database migrations dan seeding
go run cmd/main.go
# Server akan otomatis:
# - Migrate database schema
# - Seed initial data (users, accounts, categories)
# - Fix account header status
# - Start HTTP server
```
**Backend Server**: `http://localhost:8080`  
**API Documentation**: `http://localhost:8080/api/v1/health`

### 3. Setup Frontend
```bash
cd frontend

# Install Node.js dependencies
npm install
# atau
yarn install

# Setup environment variables (optional)
# Buat .env.local dan set NEXT_PUBLIC_API_URL jika berbeda dari default
echo "NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1" > .env.local

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

Employee User:
  Username: employee@company.com
  Password: password123

Director User:
  Username: director@company.com
  Password: password123

Inventory User:
  Username: inventory@company.com
  Password: password123

```

## 📚 API Documentation

### Authentication & User Management
```http
POST /api/v1/auth/register     # Register user baru
POST /api/v1/auth/login        # Login user
POST /api/v1/auth/refresh      # Refresh JWT token
GET  /api/v1/profile           # Get user profile
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
```

### Reports & Analytics
```http
GET    /api/v1/reports/sales               # Sales reports
GET    /api/v1/reports/purchases           # Purchase reports
GET    /api/v1/reports/inventory           # Inventory reports
GET    /api/v1/reports/financial           # Financial reports
GET    /api/v1/dashboard/summary           # Dashboard data
```

### System Monitoring (Admin Only)
```http
GET    /api/v1/monitoring/status           # System status
GET    /api/v1/monitoring/audit-logs       # Audit trails
GET    /api/v1/notifications               # System notifications
GET    /api/v1/health                      # Health check
```

## 🗃️ Database Schema

### Core Business Tables
- **users** - User authentication, roles, dan profile
- **contacts** - Customers, vendors, employees, sales persons
- **products** - Master products dengan variants
- **product_categories** - Hierarchical product categorization
- **accounts** - Chart of accounts dengan hierarchy

### Transaction Tables
- **sales** & **sale_items** - Sales transactions (quotation→invoice→payment)
- **sale_payments** & **sale_returns** - Payment tracking & returns
- **purchases** & **purchase_items** - Purchase transactions
- **purchase_receipts** - Goods receipt tracking
- **inventories** - Stock movement logs
- **cash_banks** - Bank accounts dan cash management
- **payments** - Universal payment records

### System Tables
- **approval_workflows** - Configurable approval processes
- **notifications** - System notifications
- **audit_logs** - Complete audit trail
- **assets** - Fixed asset management
- **stock_alerts** - Minimum stock monitoring

## 🔧 Development Guide

### Backend Development
```bash
cd backend

# Development mode dengan auto-reload
go run cmd/main.go

# Run specific maintenance scripts
go run scripts/fix_accounts.go
go run scripts/check_sales_codes.go

# Database operations
go run cmd/cleanup_accounts.go
go run cmd/fix_sales_foreign_key.go

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
NEXT_PUBLIC_API_URL=https://your-api-domain.com/api/v1
```

### Database Migration
- Auto-migration saat aplikasi start
- Manual migration dengan `go run migrations/migrations.go`
- Backup database sebelum deployment

### Security Checklist
- [ ] Update JWT_SECRET untuk production
- [ ] Enable HTTPS untuk API dan frontend
- [ ] Configure database SSL
- [ ] Set proper CORS origins
- [ ] Enable rate limiting
- [ ] Setup monitoring dan logging

## 🧪 Testing

### Backend Testing
```bash
cd backend
go test ./...
```

### Frontend Testing
```bash
cd frontend
npm run test
```

## ✅ Implementation Status

### ✅ Completed Features
- [x] **Complete Authentication System** - JWT + Refresh tokens
- [x] **Role-Based Access Control** - 5 user roles dengan granular permissions
- [x] **Sales Module** - End-to-end sales process
- [x] **Purchase Module** - Full procurement workflow dengan approvals
- [x] **Inventory Management** - Real-time stock tracking
- [x] **Financial Management** - Chart of accounts, payments
- [x] **Asset Management** - Fixed asset tracking dengan depreciation
- [x] **Dashboard Analytics** - Role-based dashboards
- [x] **Notification System** - Real-time alerts
- [x] **Export Features** - PDF/Excel reports
- [x] **Audit Trail** - Complete activity logging
- [x] **Security Features** - Rate limiting, token monitoring

### 🚧 In Progress
- [ ] **Advanced Reporting** - Financial statements, analytics
- [ ] **Multi-location Support** - Warehouse management
- [ ] **API Documentation** - Swagger/OpenAPI specs
- [ ] **Unit Testing** - Backend dan frontend test coverage

### 📋 Future Roadmap
- [ ] **Docker Support** - Containerized deployment
- [ ] **PWA Features** - Offline capability
- [ ] **Mobile App** - React Native companion app
- [ ] **Integration APIs** - Third-party integrations
- [ ] **Advanced Analytics** - ML-powered insights
- [ ] **Multi-currency** - Full international support

## 🤝 Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 🎯 Key Highlights

### 🏆 Production Ready
- **Comprehensive Business Logic** - Real-world accounting principles
- **Enterprise Security** - Multi-layer security dengan audit trails
- **Scalable Architecture** - Clean architecture dengan separation of concerns
- **Modern Tech Stack** - Latest versions dengan best practices
- **Responsive Design** - Mobile-first UI/UX

### 📊 Business Impact
- **Streamlined Operations** - Integrated sales-to-payment workflow
- **Financial Control** - Real-time financial visibility
- **Inventory Optimization** - Smart stock management
- **Compliance Ready** - Indonesian tax regulations (PPN/PPh)
- **Decision Support** - Executive dashboards dan analytics

### Development Standards
- Follow clean code principles
- Add proper error handling
- Include type definitions
- Write descriptive commit messages
- Update documentation

## 📞 Support & Documentation

- **Issues**: Create GitHub issue untuk bugs atau feature requests
- **Documentation**: Lihat folder `backend/docs/` untuk detailed documentation
- **API Testing**: Gunakan `/api/v1/debug/` endpoints untuk testing

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**🚀 Built with modern technologies for efficient business management**  
*Sistem Akuntansi Modern - Comprehensive ERP Solution*
