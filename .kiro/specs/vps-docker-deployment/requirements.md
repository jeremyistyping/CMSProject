# Requirements Document

## Introduction

Sistem deployment aplikasi full-stack (Go backend + Next.js frontend) ke VPS menggunakan Docker dengan konfigurasi yang fleksibel dan tidak hardcoded. Deployment harus mendukung environment variables untuk semua URL, IP, dan konfigurasi sensitif.

## Glossary

- **VPS**: Virtual Private Server - server virtual untuk hosting aplikasi
- **Docker**: Platform containerization untuk packaging dan deployment aplikasi
- **Docker_Compose**: Tool untuk mendefinisikan dan menjalankan multi-container Docker applications
- **Environment_Variables**: Variabel konfigurasi yang dapat diubah tanpa mengubah kode
- **Reverse_Proxy**: Server yang menerima request dan meneruskannya ke backend services (Nginx/Caddy)
- **Backend_Service**: Aplikasi Go yang berjalan di port 8080
- **Frontend_Service**: Aplikasi Next.js yang berjalan di port 3000
- **Database_Service**: PostgreSQL database
- **Git_Repository**: Repository kode sumber aplikasi
- **SSL_Certificate**: Sertifikat untuk HTTPS connection

## Requirements

### Requirement 1: Docker Configuration

**User Story:** As a DevOps engineer, I want to containerize the application using Docker, so that deployment is consistent across environments.

#### Acceptance Criteria

1. WHEN building the backend container, THE Docker_Build_Process SHALL use multi-stage build to optimize image size
2. WHEN building the frontend container, THE Docker_Build_Process SHALL use environment variables for API URL configuration
3. WHEN starting containers, THE Docker_Compose SHALL define all services (backend, frontend, database, reverse proxy)
4. THE Docker_Compose SHALL use environment variables for all configuration values
5. THE Docker_Compose SHALL define persistent volumes for database data and uploaded files

### Requirement 2: Environment Configuration

**User Story:** As a system administrator, I want all URLs and IPs configurable via environment variables, so that I can deploy to different environments without code changes.

#### Acceptance Criteria

1. THE Backend_Service SHALL read database connection string from environment variables
2. THE Backend_Service SHALL read JWT secrets from environment variables
3. THE Backend_Service SHALL read CORS allowed origins from environment variables
4. THE Frontend_Service SHALL read API URL from environment variables
5. THE Reverse_Proxy SHALL read domain names from environment variables
6. WHEN environment variables are missing, THE System SHALL use sensible defaults for development
7. THE System SHALL provide separate .env.example files for each service

### Requirement 3: Git-Based Deployment

**User Story:** As a developer, I want to deploy using Git, so that I can easily update the application on the VPS.

#### Acceptance Criteria

1. WHEN deploying for the first time, THE Deployment_Process SHALL clone the Git repository to the VPS
2. WHEN updating the application, THE Deployment_Process SHALL pull latest changes from Git
3. THE Deployment_Process SHALL provide deployment scripts for initial setup and updates
4. THE Deployment_Process SHALL backup database before updates
5. THE Deployment_Process SHALL rebuild Docker containers after Git pull

### Requirement 4: Reverse Proxy Configuration

**User Story:** As a system administrator, I want a reverse proxy to handle SSL and route traffic, so that the application is secure and accessible.

#### Acceptance Criteria

1. THE Reverse_Proxy SHALL route requests to appropriate services based on path
2. THE Reverse_Proxy SHALL handle SSL certificate management
3. THE Reverse_Proxy SHALL forward API requests to Backend_Service
4. THE Reverse_Proxy SHALL forward web requests to Frontend_Service
5. THE Reverse_Proxy SHALL be configurable via environment variables

### Requirement 5: Database Management

**User Story:** As a database administrator, I want PostgreSQL running in a container with persistent storage, so that data is preserved across container restarts.

#### Acceptance Criteria

1. THE Database_Service SHALL use persistent volumes for data storage
2. THE Database_Service SHALL be accessible only from backend container
3. THE Database_Service SHALL use credentials from environment variables
4. WHEN container restarts, THE Database_Service SHALL preserve all data
5. THE Deployment_Process SHALL provide database backup scripts

### Requirement 6: File Upload Persistence

**User Story:** As a user, I want uploaded files to persist across deployments, so that I don't lose data when the application updates.

#### Acceptance Criteria

1. THE Backend_Service SHALL store uploaded files in a persistent volume
2. THE Frontend_Service SHALL store any cached files in a persistent volume
3. WHEN containers are recreated, THE System SHALL preserve all uploaded files
4. THE Deployment_Process SHALL backup uploaded files before updates

### Requirement 7: Deployment Scripts

**User Story:** As a system administrator, I want automated deployment scripts, so that deployment is repeatable and error-free.

#### Acceptance Criteria

1. THE Deployment_Process SHALL provide an initial setup script
2. THE Deployment_Process SHALL provide an update/redeploy script
3. THE Deployment_Process SHALL provide a backup script
4. THE Deployment_Process SHALL provide a rollback script
5. WHEN scripts execute, THE System SHALL log all actions
6. WHEN errors occur, THE Scripts SHALL provide clear error messages

### Requirement 8: Security Configuration

**User Story:** As a security officer, I want secure defaults and configurable security settings, so that the application is protected in production.

#### Acceptance Criteria

1. THE System SHALL not expose database ports to the public internet
2. THE System SHALL use strong passwords from environment variables
3. THE System SHALL enable HTTPS in production
4. THE System SHALL configure CORS properly based on environment
5. THE System SHALL use secure cookie settings in production
6. THE Deployment_Process SHALL provide security checklist documentation

### Requirement 9: Monitoring and Logging

**User Story:** As a system administrator, I want centralized logging, so that I can troubleshoot issues easily.

#### Acceptance Criteria

1. THE Docker_Compose SHALL configure logging drivers for all services
2. THE System SHALL persist logs to the host filesystem
3. THE System SHALL rotate logs to prevent disk space issues
4. THE Deployment_Process SHALL provide log viewing scripts
5. WHEN errors occur, THE System SHALL log detailed error information

### Requirement 10: Documentation

**User Story:** As a new team member, I want comprehensive deployment documentation, so that I can deploy and maintain the application.

#### Acceptance Criteria

1. THE Documentation SHALL provide step-by-step VPS setup instructions
2. THE Documentation SHALL explain all environment variables
3. THE Documentation SHALL provide troubleshooting guide
4. THE Documentation SHALL include architecture diagrams
5. THE Documentation SHALL provide examples for different deployment scenarios
