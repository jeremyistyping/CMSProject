# Design Document: VPS Docker Deployment

## Overview

Sistem deployment aplikasi full-stack (Go backend + Next.js frontend) ke VPS menggunakan Docker Compose dengan konfigurasi yang sepenuhnya fleksibel melalui environment variables. Design ini mendukung deployment dengan atau tanpa domain name, menggunakan IP address sebagai fallback.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         VPS Server                           │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Nginx Reverse Proxy                       │ │
│  │         (Port 80/443 - Public Access)                  │ │
│  └─────────────┬──────────────────────┬───────────────────┘ │
│                │                      │                      │
│                ▼                      ▼                      │
│  ┌─────────────────────┐   ┌──────────────────────┐        │
│  │  Frontend Container │   │  Backend Container   │        │
│  │    (Next.js)        │   │      (Go API)        │        │
│  │    Port: 3000       │   │    Port: 8080        │        │
│  └─────────────────────┘   └──────────┬───────────┘        │
│                                        │                     │
│                                        ▼                     │
│                          ┌──────────────────────┐           │
│                          │ PostgreSQL Container │           │
│                          │    Port: 5432        │           │
│                          │  (Internal Only)     │           │
│                          └──────────────────────┘           │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Persistent Volumes                        │ │
│  │  • postgres_data (Database)                            │ │
│  │  • backend_uploads (File Uploads)                      │ │
│  │  • nginx_ssl (SSL Certificates - optional)             │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Deployment Flow

```
┌──────────────┐
│ Local Dev    │
│ Machine      │
└──────┬───────┘
       │ git push
       ▼
┌──────────────┐
│ Git Repo     │
│ (GitHub/     │
│  GitLab)     │
└──────┬───────┘
       │ git pull
       ▼
┌──────────────────────────────────────┐
│ VPS Server                           │
│                                      │
│ 1. Pull latest code                  │
│ 2. Backup database                   │
│ 3. Build Docker images               │
│ 4. Stop old containers               │
│ 5. Start new containers              │
│ 6. Run migrations                    │
│ 7. Health check                      │
└──────────────────────────────────────┘
```

## Components and Interfaces

### 1. Docker Compose Configuration

**File**: `docker-compose.yml`

```yaml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: ${COMPOSE_PROJECT_NAME:-app}_postgres
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - app_network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Backend API (Go)
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: ${COMPOSE_PROJECT_NAME:-app}_backend
    environment:
      DATABASE_URL: postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable
      JWT_SECRET: ${JWT_SECRET}
      SERVER_PORT: 8080
      ENVIRONMENT: ${ENVIRONMENT:-production}
      ALLOWED_ORIGINS: ${ALLOWED_ORIGINS}
      SKIP_BALANCE_RESET: ${SKIP_BALANCE_RESET:-false}
    volumes:
      - backend_uploads:/app/uploads
      - ./backend/migrations:/app/migrations:ro
    networks:
      - app_network
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Frontend (Next.js)
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
      args:
        NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL}
    container_name: ${COMPOSE_PROJECT_NAME:-app}_frontend
    environment:
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL}
    networks:
      - app_network
    depends_on:
      - backend
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:3000"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Nginx Reverse Proxy
  nginx:
    image: nginx:alpine
    container_name: ${COMPOSE_PROJECT_NAME:-app}_nginx
    ports:
      - "${HTTP_PORT:-80}:80"
      - "${HTTPS_PORT:-443}:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - nginx_ssl:/etc/nginx/ssl:ro
    networks:
      - app_network
    depends_on:
      - frontend
      - backend
    restart: unless-stopped

volumes:
  postgres_data:
  backend_uploads:
  nginx_ssl:

networks:
  app_network:
    driver: bridge
```

### 2. Environment Configuration

**File**: `.env.production`

```bash
# ================================================================
# PRODUCTION ENVIRONMENT CONFIGURATION
# ================================================================

# Project Configuration
COMPOSE_PROJECT_NAME=accounting_app

# Server Configuration
# ====================
# Use your VPS IP or domain name
# Examples:
#   - IP only: http://192.168.1.100
#   - With domain: https://accounting.yourdomain.com
SERVER_HOST=http://YOUR_VPS_IP_HERE
HTTP_PORT=80
HTTPS_PORT=443

# Database Configuration
# ======================
POSTGRES_DB=sistem_akuntansi_prod
POSTGRES_USER=accounting_user
POSTGRES_PASSWORD=CHANGE_THIS_STRONG_PASSWORD_123!

# Backend Configuration
# =====================
JWT_SECRET=CHANGE_THIS_TO_VERY_LONG_RANDOM_SECRET_KEY_FOR_PRODUCTION
ENVIRONMENT=production
SKIP_BALANCE_RESET=false

# CORS Configuration
# ==================
# Add your domain or IP here
# Examples:
#   - IP only: http://192.168.1.100
#   - With domain: https://accounting.yourdomain.com,https://www.accounting.yourdomain.com
ALLOWED_ORIGINS=${SERVER_HOST}

# Frontend Configuration
# ======================
# This should point to your backend API
# Examples:
#   - IP only: http://192.168.1.100/api
#   - With domain: https://accounting.yourdomain.com/api
NEXT_PUBLIC_API_URL=${SERVER_HOST}/api

# SSL Configuration (Optional - for domain with HTTPS)
# =====================================================
ENABLE_SSL=false
SSL_CERT_PATH=/etc/nginx/ssl/cert.pem
SSL_KEY_PATH=/etc/nginx/ssl/key.pem

# Backup Configuration
# ====================
BACKUP_DIR=/opt/backups/accounting_app
BACKUP_RETENTION_DAYS=30
```

### 3. Nginx Configuration

**File**: `nginx/conf.d/default.conf`

```nginx
# Upstream definitions
upstream backend {
    server backend:8080;
}

upstream frontend {
    server frontend:3000;
}

# HTTP Server (works with IP or domain)
server {
    listen 80;
    server_name _;  # Accepts any hostname (IP or domain)
    
    client_max_body_size 100M;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    
    # API routes
    location /api/ {
        proxy_pass http://backend/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Health check endpoint
    location /health {
        proxy_pass http://backend/health;
        access_log off;
    }
    
    # Frontend routes
    location / {
        proxy_pass http://frontend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
    
    # Static files caching
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://frontend;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}

# HTTPS Server (optional - only if SSL is configured)
# Uncomment this block when you have a domain and SSL certificate
# server {
#     listen 443 ssl http2;
#     server_name your-domain.com;
#     
#     ssl_certificate /etc/nginx/ssl/cert.pem;
#     ssl_certificate_key /etc/nginx/ssl/key.pem;
#     ssl_protocols TLSv1.2 TLSv1.3;
#     ssl_ciphers HIGH:!aNULL:!MD5;
#     ssl_prefer_server_ciphers on;
#     
#     client_max_body_size 100M;
#     
#     # Same location blocks as HTTP server above
#     # ... (copy from above)
# }
```

### 4. Updated Dockerfiles

**Backend Dockerfile** (`backend/Dockerfile`):

```dockerfile
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o api ./main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

# Set timezone
ENV TZ=Asia/Jakarta

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/api .

# Copy migrations
COPY --from=builder /build/migrations ./migrations

# Create uploads directory
RUN mkdir -p /app/uploads

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./api"]
```

**Frontend Dockerfile** (`frontend/Dockerfile`):

```dockerfile
FROM node:18-alpine AS deps

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./
RUN npm ci --only=production

# Builder stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy dependencies
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Build arguments
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

# Build the application
RUN npm run build

# Production stage
FROM node:18-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production

# Create non-root user
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy necessary files
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

# Set ownership
RUN chown -R nextjs:nodejs /app

USER nextjs

EXPOSE 3000

ENV PORT 3000
ENV HOSTNAME "0.0.0.0"

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:3000 || exit 1

CMD ["node", "server.js"]
```

### 5. Frontend Next.js Configuration

**File**: `frontend/next.config.ts`

Update to support standalone output:

```typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone',
  
  // Environment variables
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  },
  
  // Image optimization
  images: {
    domains: [],
    unoptimized: true, // For Docker deployment
  },
  
  // Disable telemetry in production
  telemetry: false,
  
  // Webpack configuration
  webpack: (config) => {
    return config;
  },
};

export default nextConfig;
```

## Data Models

### Deployment Configuration Model

```typescript
interface DeploymentConfig {
  // Server
  serverHost: string;        // IP or domain
  httpPort: number;          // Default: 80
  httpsPort: number;         // Default: 443
  
  // Database
  postgresDb: string;
  postgresUser: string;
  postgresPassword: string;
  
  // Security
  jwtSecret: string;
  environment: 'development' | 'staging' | 'production';
  
  // CORS
  allowedOrigins: string[];
  
  // SSL (optional)
  enableSsl: boolean;
  sslCertPath?: string;
  sslKeyPath?: string;
  
  // Backup
  backupDir: string;
  backupRetentionDays: number;
}
```

### Deployment State Model

```typescript
interface DeploymentState {
  version: string;
  deployedAt: Date;
  gitCommit: string;
  gitBranch: string;
  containers: {
    name: string;
    status: 'running' | 'stopped' | 'error';
    health: 'healthy' | 'unhealthy' | 'starting';
  }[];
  volumes: {
    name: string;
    size: string;
  }[];
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Environment Variable Isolation
*For any* deployment configuration, changing environment variables should not require code changes or rebuilding application source code.
**Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5**

### Property 2: Container Health Persistence
*For any* container restart, all data in persistent volumes (database, uploads) should remain intact and accessible.
**Validates: Requirements 5.4, 6.3**

### Property 3: Service Dependency Resolution
*For any* service startup sequence, dependent services should wait for their dependencies to be healthy before starting.
**Validates: Requirements 1.3**

### Property 4: Configuration Flexibility
*For any* valid SERVER_HOST value (IP address or domain name), the system should route requests correctly to backend and frontend services.
**Validates: Requirements 2.5, 4.1, 4.3, 4.4**

### Property 5: Deployment Idempotency
*For any* deployment script execution, running the script multiple times should produce the same result without errors.
**Validates: Requirements 7.1, 7.2**

### Property 6: Backup Integrity
*For any* backup operation, the backed-up database should be restorable to a working state.
**Validates: Requirements 3.4, 5.5**

### Property 7: Zero-Downtime Updates
*For any* application update via Git pull, the old version should continue serving requests until the new version is healthy.
**Validates: Requirements 3.2, 3.5**

### Property 8: Port Isolation
*For any* deployment, database ports should not be accessible from outside the Docker network.
**Validates: Requirements 8.1**

### Property 9: Log Persistence
*For any* container restart, logs should be preserved and accessible for troubleshooting.
**Validates: Requirements 9.2, 9.3**

### Property 10: SSL Optional Configuration
*For any* deployment, the system should work correctly with or without SSL configuration.
**Validates: Requirements 4.2, 8.3**

## Error Handling

### Deployment Errors

1. **Git Pull Failures**
   - Retry with exponential backoff
   - Log error details
   - Notify administrator
   - Keep current version running

2. **Docker Build Failures**
   - Log build output
   - Preserve previous images
   - Rollback to last working version
   - Alert administrator

3. **Database Migration Failures**
   - Stop deployment process
   - Restore from backup
   - Log migration errors
   - Require manual intervention

4. **Health Check Failures**
   - Retry health checks (3 attempts)
   - If failed, rollback to previous version
   - Log failure reasons
   - Send alerts

5. **Volume Mount Errors**
   - Check permissions
   - Verify volume exists
   - Create if missing
   - Log detailed error

### Runtime Errors

1. **Container Crashes**
   - Auto-restart (unless-stopped policy)
   - Log crash details
   - Alert after 3 consecutive failures
   - Preserve logs for debugging

2. **Database Connection Failures**
   - Retry with exponential backoff
   - Check database health
   - Log connection attempts
   - Fallback to read-only mode if possible

3. **Disk Space Issues**
   - Monitor disk usage
   - Alert at 80% capacity
   - Auto-cleanup old logs
   - Rotate backups

4. **Network Issues**
   - Retry failed requests
   - Log network errors
   - Check container network connectivity
   - Restart networking if needed

## Testing Strategy

### Manual Testing

1. **Initial Deployment Test**
   - Deploy to fresh VPS
   - Verify all containers start
   - Test health endpoints
   - Verify database connectivity
   - Test file uploads
   - Check logs

2. **Update Deployment Test**
   - Make code changes
   - Push to Git
   - Run update script
   - Verify zero downtime
   - Test new features
   - Verify data persistence

3. **Backup and Restore Test**
   - Create backup
   - Verify backup file exists
   - Restore to new database
   - Verify data integrity
   - Test application functionality

4. **Rollback Test**
   - Deploy new version
   - Trigger rollback
   - Verify previous version restored
   - Test functionality
   - Check data consistency

5. **IP vs Domain Test**
   - Deploy with IP address
   - Test all functionality
   - Add domain name
   - Update configuration
   - Test with domain
   - Verify both work

6. **SSL Configuration Test** (when domain available)
   - Install SSL certificate
   - Update Nginx config
   - Test HTTPS access
   - Verify HTTP redirect
   - Check certificate validity

### Integration Testing

1. **Container Communication**
   - Frontend → Backend API calls
   - Backend → Database queries
   - Nginx → Frontend/Backend routing

2. **Volume Persistence**
   - Upload files
   - Restart containers
   - Verify files still accessible

3. **Environment Variables**
   - Change .env values
   - Restart containers
   - Verify new values applied

4. **Health Checks**
   - Stop backend
   - Verify health check fails
   - Restart backend
   - Verify health check passes

### Security Testing

1. **Database Access**
   - Attempt external connection
   - Verify connection refused
   - Test from within network
   - Verify connection succeeds

2. **CORS Configuration**
   - Test from allowed origin
   - Test from disallowed origin
   - Verify proper rejection

3. **SSL/TLS** (when configured)
   - Test certificate validity
   - Verify TLS version
   - Check cipher suites
   - Test HTTP to HTTPS redirect

### Performance Testing

1. **Load Testing**
   - Simulate concurrent users
   - Monitor response times
   - Check resource usage
   - Verify no memory leaks

2. **Database Performance**
   - Test query performance
   - Monitor connection pool
   - Check index usage
   - Verify backup speed

## Deployment Scenarios

### Scenario 1: Fresh Deployment (IP Only)

```bash
# On VPS
SERVER_HOST=http://192.168.1.100
ALLOWED_ORIGINS=http://192.168.1.100
NEXT_PUBLIC_API_URL=http://192.168.1.100/api
```

Access: `http://192.168.1.100`

### Scenario 2: Fresh Deployment (With Domain)

```bash
# On VPS
SERVER_HOST=https://accounting.company.com
ALLOWED_ORIGINS=https://accounting.company.com
NEXT_PUBLIC_API_URL=https://accounting.company.com/api
ENABLE_SSL=true
```

Access: `https://accounting.company.com`

### Scenario 3: Migration from IP to Domain

1. Deploy with IP initially
2. Configure domain DNS to point to VPS IP
3. Update `.env.production`:
   ```bash
   SERVER_HOST=https://accounting.company.com
   ALLOWED_ORIGINS=http://192.168.1.100,https://accounting.company.com
   ```
4. Install SSL certificate
5. Update Nginx config for HTTPS
6. Restart containers
7. Test both IP and domain access
8. Remove IP from ALLOWED_ORIGINS after migration complete

## Security Considerations

1. **Secrets Management**
   - Never commit `.env` files
   - Use strong passwords (min 20 characters)
   - Rotate JWT secrets regularly
   - Use different secrets per environment

2. **Network Security**
   - Database not exposed to public
   - Use Docker internal network
   - Configure firewall rules
   - Limit SSH access

3. **SSL/TLS**
   - Use Let's Encrypt for free certificates
   - Enable HTTPS in production
   - Redirect HTTP to HTTPS
   - Use strong cipher suites

4. **Container Security**
   - Run as non-root user
   - Use official base images
   - Keep images updated
   - Scan for vulnerabilities

5. **Backup Security**
   - Encrypt backup files
   - Store in secure location
   - Limit access permissions
   - Test restore regularly

## Monitoring and Maintenance

1. **Health Monitoring**
   - Container health checks
   - Application health endpoints
   - Database connectivity
   - Disk space usage

2. **Log Management**
   - Centralized logging
   - Log rotation
   - Error alerting
   - Audit trails

3. **Backup Schedule**
   - Daily database backups
   - Weekly full backups
   - Monthly archive backups
   - Automated cleanup

4. **Update Strategy**
   - Test in staging first
   - Backup before updates
   - Monitor after deployment
   - Rollback plan ready
