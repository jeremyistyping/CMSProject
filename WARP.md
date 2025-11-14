# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

---

## Project overview

This is a monorepo for a **modern accounting system** with:

- **Backend**: Go 1.23+ REST API (`backend/`) using Gin, GORM, PostgreSQL/MySQL, JWT auth, SSOT (Single Source of Truth) journal system, and an advanced **Balance Protection System**.
- **Frontend**: Next.js 15 + TypeScript (`frontend/`) with App Router, Chakra UI + Tailwind, dark/light theming, and multi-language (ID/EN) support.

For detailed product- and feature-level docs, see the root `README.md`.

---

## Common commands

### Backend (Go API – `backend/`)

**Basic dev flow**

```bash
cd backend

# Install Go deps (first time or after go.mod changes)
go mod tidy

# Copy environment template and configure DB / JWT / ports
cp .env.example .env

# Run the backend (recommended entrypoint)
go run cmd/main.go
```

The backend will:
- Connect to the configured PostgreSQL/MySQL database.
- Auto-run Go-based migrations and SQL migrations.
- Install balance-sync and journal functions.
- Seed initial data (users, chart of accounts, etc.).

**Database fixes / first-run issues** (UUID extension, migration glitches, etc.)

```bash
cd backend

# Apply DB fixes if startup complains about extensions/migrations
go run apply_database_fixes.go

# If migrations are still problematic
go run cmd/fix_migrations.go
go run cmd/fix_remaining_migrations.go

# Verify DB health
go run cmd/final_verification.go
```

**Balance Protection System setup (per machine, critical for correct balances)**

From `backend/`:

- **Windows**
  ```bash
  setup_balance_protection.bat
  ```
- **Linux/macOS**
  ```bash
  chmod +x setup_balance_protection.sh
  ./setup_balance_protection.sh
  ```
- **Manual alternative**
  ```bash
  go run cmd/scripts/setup_balance_sync_auto.go
  ```

These scripts install triggers/functions and monitoring tables used by the SSOT/balance system. Do not run ad‑hoc SQL against those internals unless explicitly requested.

**Integration / end-to-end test: sales → invoice → payment → balances**

From `backend/` (see `scripts/README.md` for full details):

- **Go-only execution**
  ```bash
  go run scripts/test_sales_payment_flow.go
  ```

- **Windows PowerShell**
  ```powershell
  .\scripts\run_test.ps1                 # basic run
  .\scripts\run_test.ps1 -Verbose -WaitForServer
  .\scripts\run_test.ps1 -ServerURL http://localhost:8081
  ```

- **Windows CMD**
  ```cmd
  scripts\quick_test.bat
  ```

- **Linux/macOS**
  ```bash
  chmod +x scripts/run_test.sh           # first time only
  ./scripts/run_test.sh                  # basic run
  ./scripts/run_test.sh --verbose --wait-server
  ./scripts/run_test.sh --url http://localhost:8081
  ```

Optional base URL override:

```bash
export TEST_BASE_URL="http://localhost:8080/api/v1"
```

This suite verifies that sales, payments, journal entries, and monitored account balances move consistently through the SSOT system.

**SSOT migration / maintenance helpers (use only when explicitly requested)**

The `backend/Makefile` is focused on migrating to and validating the SSOT journal system:

```bash
cd backend

# Discover available SSOT-related commands
make help

# Install Go deps used by SSOT tools
make deps

# Check SSOT DB status (safe diagnostic)
make check-status

# Spin up backend, hit a basic SSOT endpoint, and tear down (smoke test)
make test-ssot
```

The following targets are **potentially destructive and interactive**; **do not** run them automatically:

- `make migrate-ssot`
- `make cleanup-models`
- `make update-routes`
- `make full-migration`

Run them only if the user explicitly asks for SSOT migration or cleanup.

**Backend build / Docker**

```bash
cd backend

# Local binary builds (examples from docs)
go build -o app cmd/main.go
# or
go build -o sistem-akuntansi cmd/main.go

# Container build (DigitalOcean registry target from repo docs)
docker build --push --platform linux/amd64 \
  -t registry.digitalocean.com/registry-tigapilar/dbm/account-backend:latest .
```

---

### Frontend (Next.js – `frontend/`)

**Install & dev server**

```bash
cd frontend

# Install Node deps
npm install

# (Optional) configure API base URL
# creates .env.local with NEXT_PUBLIC_API_URL
echo "NEXT_PUBLIC_API_URL=http://localhost:8080/" > .env.local

# Development server (Next.js 15, Turbopack)
npm run dev
```

Alternative dev modes (from `package.json`):

- `npm run dev:clean` – dev without the custom deprecation suppressor.
- `npm run dev:verbose` – dev without extra NODE_OPTIONS.

**Lint, typecheck, build**

```bash
cd frontend

# ESLint via Next
npm run lint

# TypeScript typecheck (from root README)
npx tsc --noEmit

# Production build & start
npm run build
npm run start
```

**Frontend Docker build**

From `frontend/`:

```bash
docker build --push --platform linux/amd64 \
  -t registry.digitalocean.com/registry-tigapilar/dbm/account-fe:latest .
```

---

### Database & migrations

Migrations live in `backend/migrations/` and are managed via **golang-migrate** plus a Go `MigrationService` that runs at backend startup.

**Normal development path**

In most cases you **only** need to:

1. Ensure the database exists (e.g., `sistem_akuntansi`).
2. Configure `.env`.
3. Run `go run cmd/main.go`.

The application will automatically apply pending migrations on startup.

**Creating new migrations** (see `backend/migrations/README.md` for full details)

Typical workflow:

```bash
cd backend/migrations

# Create a new sequential SQL migration pair
migrate create -ext sql -dir . -seq add_my_feature

# Then edit the generated .up.sql and .down.sql files
```

Key conventions (from the migrations README):

- Always create both `.up.sql` (apply) and `.down.sql` (rollback).
- Make migrations idempotent (`CREATE TABLE IF NOT EXISTS`, etc.).
- Use transactions (`BEGIN; ... COMMIT;`).
- Do **not** edit previously applied migration files; create new ones instead.

Manual CLI usage for debugging (not normally needed in dev because startup runs them):

```bash
# From backend/
# Apply all pending migrations
migrate -path ./migrations -database "<DB_URL>" up

# Roll back last migration
migrate -path ./migrations -database "<DB_URL>" down 1
```

Refer to `backend/migrations/README.md` for troubleshooting dirty versions, out-of-order migrations, and CI/CD integration examples.

---

## Architecture overview

### Monorepo layout

- `backend/` – Go REST API, database migrations, SSOT journal, balance protection, test utilities.
- `frontend/` – Next.js 15 app with App Router, theming, translations, and UI components.
- Root `README.md` – canonical, detailed product/feature documentation and quick start.

### Backend architecture (`backend/`)

The backend follows a layered, clean-ish architecture as described in the root `README.md` and backed by `go.mod`:

- **Entry points (`cmd/` and root `main.go`)**
  - `cmd/main.go` – primary server entrypoint used by the main docs (`go run cmd/main.go`).
  - Additional commands under `cmd/` and `cmd/scripts/` for DB verification, migration fixes, SSOT setup, and balance synchronization.

- **HTTP/API layer**
  - `controllers/` – HTTP handlers, including enhanced security controllers and monitoring endpoints.
  - `routes/` – route registration and API versioning (`/api/v1/...`), including debug and monitoring routes.
  - `middleware/` – JWT auth, RBAC, CORS, validation, security headers, rate limiting, logging.

- **Domain & application layer**
  - `services/` – business logic for sales, purchases, inventory, cash/bank, reporting, security, balance monitoring, and background workers.
  - `models/` – GORM entities, DTOs, and types tying business concepts to the DB schema.
  - `repositories/` – data access layer encapsulating GORM/Postgres/MySQL specifics.

- **Infrastructure & support**
  - `migrations/` – SQL migrations managed by golang-migrate; tightly coupled to the `MigrationService` invoked at startup.
  - `database/`, `config/`, `startup/` – DB connections, configuration loading (e.g., `.env` via `godotenv`), and bootstrapping (migrations, seeding, background workers, Swagger).
  - `integration/` – third-party integrations (external services / providers).
  - `docs/` – API/system documentation (Swagger/OpenAPI and internal guides).
  - `tools/`, `scripts/`, `debug_scripts/` – operational and debugging utilities (balance sync, account fixes, SSOT migration helpers, comprehensive test scripts).

- **Cross-cutting systems**
  - **SSOT Journal System** – centralized journal data model and reporting layer. The `Makefile` and various `cmd/scripts/*` tools are dedicated to migrating existing journals into SSOT and keeping routes/models aligned.
  - **Balance Protection System** – DB triggers, monitoring tables, and Go helpers (plus shell/batch wrappers) that:
    - Auto-sync account balances on transactional changes.
    - Detect and report mismatches via monitoring views/tables.
    - Provide scripted fixes and health checks.

When modifying core accounting behavior (journals, balances, account tables), keep these two systems in mind and prefer updating migrations and dedicated scripts rather than bypassing them.

### Frontend architecture (`frontend/`)

The frontend architecture is summarized in the root `README.md` and centered on the Next.js App Router:

- **App shell (`app/`)**
  - `app/layout.tsx` – global layout, theme initialization, providers.
  - `app/ClientProviders.tsx` – wrapper for React context providers (theme, language, auth, etc.).
  - Route-specific subtrees under `app/` (e.g., project edit flows) use this shared shell.

- **Feature & UI layer (`src/`)**
  - `src/components/` – feature-oriented component folders, including `common/`, `reports/`, `settings/`, `users/`, `payments/` and others.
    - `common/SimpleThemeToggle.tsx` – core theme switcher used across the app.
  - `src/contexts/` – React contexts for:
    - `SimpleThemeContext.tsx` – dark/light theme state.
    - `LanguageContext.tsx` – ID/EN language selection and translation lookup.
    - `AuthContext.tsx` – authentication state, tokens, and user role/permissions.
  - `src/hooks/` – custom hooks for translations (`useTranslation`), permissions (`usePermissions`), and other cross-cutting concerns.
  - `src/services/` – API client layer (Axios) for talking to the Go backend, including financial reporting endpoints.
  - `src/translations/` – language files and translation keys.
  - `src/utils/`, `src/types/` – shared helpers and TypeScript types.

The UI is tightly integrated with theme and language contexts; when editing UI or adding routes, ensure components participate in the existing theming and translation patterns.

---

## Important conventions & invariants

- **Backend startup is responsible for DB shape.** Whenever possible, let `go run cmd/main.go` drive migrations and seeding instead of manually editing the DB.
- **Balance Protection and SSOT systems are critical invariants.**
  - Always ensure `setup_balance_protection.(bat|sh)` (or `setup_balance_sync_auto.go`) has been run on any new environment before relying on balances.
  - If you change journal-related models/tables, plan corresponding migrations and SSOT scripts rather than ad‑hoc table edits.
- **Migrations are append-only.** Do not rewrite existing `.up.sql`/`.down.sql` once deployed; add new sequential migrations.
- **Frontend changes must respect theming and i18n.** When updating UI:
  - Wire new components through existing theme context and CSS variable system.
  - Add/adjust translation keys in `src/translations/` and ensure both ID and EN variants are maintained.
- **Swagger and debug endpoints exist for exploration.**
  - API health: `GET /api/v1/health`.
  - Swagger: `/swagger/index.html` when `ENABLE_SWAGGER=true` in `.env`.
  - Debug/testing routes live under `/api/v1/debug/*` and are intended for development.

For any deeper behavior or module-specific details, consult:

- Root `README.md` – full feature and API overview.
- `backend/README.md` – backend-specific quick start and Balance Protection details.
- `backend/migrations/README.md` – migration tooling and best practices.
- `backend/scripts/README.md` – comprehensive sales→payment test suite documentation.