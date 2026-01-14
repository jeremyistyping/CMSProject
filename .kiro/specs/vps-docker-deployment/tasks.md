# Implementation Plan: VPS Docker Deployment

## Overview

Implementasi deployment aplikasi ke VPS menggunakan Docker Compose dengan konfigurasi fleksibel yang mendukung akses via IP address atau domain name. Semua konfigurasi menggunakan environment variables untuk menghindari hardcoded values.

## Tasks

- [x] 1. Prepare Docker configuration files
  - Create docker-compose.yml with all services (postgres, backend, frontend, nginx)
  - Configure health checks for all services
  - Set up persistent volumes for data
  - Configure internal Docker network
  - _Requirements: 1.3, 1.4, 1.5, 5.1, 5.2_

- [x] 2. Create environment configuration templates
  - Create .env.production template with all required variables
  - Create .env.example for reference
  - Document all environment variables
  - Add validation for required variables
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7_

- [x] 3. Update backend Dockerfile
  - Implement multi-stage build for optimization
  - Add health check endpoint
  - Configure timezone to Asia/Jakarta
  - Set up uploads directory
  - Copy migrations folder
  - _Requirements: 1.1, 5.1_

- [x] 4. Update frontend Dockerfile
  - Implement multi-stage build
  - Configure build args for API URL
  - Set up standalone output mode
  - Add health check
  - Run as non-root user
  - _Requirements: 1.2, 2.4_

- [x] 5. Update frontend Next.js configuration
  - Enable standalone output mode
  - Configure environment variables
  - Disable image optimization for Docker
  - Set up proper build configuration
  - _Requirements: 1.2, 2.4_

- [x] 6. Create Nginx reverse proxy configuration
  - Create nginx.conf base configuration
  - Create default.conf for routing
  - Configure API route proxying (/api → backend:8080)
  - Configure frontend route proxying (/ → frontend:3000)
  - Add security headers
  - Configure client_max_body_size for file uploads
  - Add health check endpoint
  - Support both IP and domain access (server_name _)
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 7. Create SSL configuration template (optional)
  - Add HTTPS server block template in Nginx config
  - Configure SSL certificate paths
  - Set up SSL protocols and ciphers
  - Add HTTP to HTTPS redirect
  - Document Let's Encrypt setup process
  - _Requirements: 4.2, 8.3_

- [x] 8. Create deployment scripts
  - [x] 8.1 Create initial setup script (setup.sh)
    - Check system requirements (Docker, Docker Compose, Git)
    - Clone repository
    - Create .env.production from template
    - Prompt for configuration values
    - Generate strong secrets
    - Build Docker images
    - Start containers
    - Run database migrations
    - Display access information
    - _Requirements: 3.1, 7.1, 7.5, 7.6_
  
  - [x] 8.2 Create update/redeploy script (deploy.sh)
    - Backup database before update
    - Pull latest code from Git
    - Rebuild Docker images
    - Perform rolling update (zero downtime)
    - Run new migrations
    - Health check verification
    - Rollback on failure
    - _Requirements: 3.2, 3.3, 3.4, 3.5, 7.2, 7.5, 7.6_
  
  - [x] 8.3 Create backup script (backup.sh)
    - Backup PostgreSQL database
    - Backup uploaded files
    - Compress backup files
    - Add timestamp to backup filename
    - Store in backup directory
    - Clean old backups based on retention policy
    - _Requirements: 3.4, 5.5, 6.4, 7.3, 7.5_
  
  - [x] 8.4 Create rollback script (rollback.sh)
    - List available backup versions
    - Stop current containers
    - Restore database from backup
    - Restore uploaded files
    - Start containers with previous version
    - Verify health checks
    - _Requirements: 7.4, 7.5, 7.6_
  
  - [x] 8.5 Create log viewing script (logs.sh)
    - View logs from all containers
    - Filter logs by service
    - Follow logs in real-time
    - Search logs by keyword
    - _Requirements: 9.2, 9.4, 7.5_

- [x] 9. Create helper scripts
  - [x] 9.1 Create health check script (health-check.sh)
    - Check all container statuses
    - Verify health endpoints
    - Check database connectivity
    - Display system resources
    - _Requirements: 9.1_
  
  - [x] 9.2 Create environment setup helper (setup-env.sh)
    - Interactive prompts for configuration
    - Generate strong random secrets
    - Validate input values
    - Create .env.production file
    - _Requirements: 2.7, 8.2_
  
  - [x] 9.3 Create SSL setup script (setup-ssl.sh)
    - Install certbot
    - Generate Let's Encrypt certificate
    - Update Nginx configuration
    - Set up auto-renewal
    - _Requirements: 4.2, 8.3_

- [x] 10. Update .gitignore
  - Add .env.production to gitignore
  - Add backup directories
  - Add SSL certificate directories
  - Add log files
  - Ensure .env.example is NOT ignored
  - _Requirements: 8.2_

- [x] 11. Create comprehensive documentation
  - [x] 11.1 Create DEPLOYMENT.md
    - Prerequisites (VPS requirements, Docker installation)
    - Initial deployment steps
    - Configuration guide
    - Update procedure
    - Backup and restore guide
    - Troubleshooting section
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_
  
  - [x] 11.2 Create ENVIRONMENT_VARIABLES.md
    - List all environment variables
    - Explain each variable purpose
    - Provide examples for different scenarios
    - Document default values
    - Security recommendations
    - _Requirements: 10.2_
  
  - [x] 11.3 Create ARCHITECTURE.md
    - System architecture diagram
    - Container communication flow
    - Network topology
    - Volume structure
    - Deployment flow diagram
    - _Requirements: 10.4_
  
  - [x] 11.4 Create TROUBLESHOOTING.md
    - Common issues and solutions
    - Container debugging commands
    - Log analysis guide
    - Performance tuning tips
    - Security checklist
    - _Requirements: 10.3, 8.6_
  
  - [x] 11.5 Create MIGRATION_GUIDE.md
    - Migrating from IP to domain
    - SSL certificate setup
    - Database migration
    - Zero-downtime deployment
    - _Requirements: 10.5_

- [x] 12. Create monitoring configuration
  - Set up Docker logging driver
  - Configure log rotation
  - Create log aggregation setup
  - Add disk space monitoring
  - _Requirements: 9.1, 9.2, 9.3_

- [x] 13. Create security hardening checklist
  - Document firewall rules
  - SSH security configuration
  - Docker security best practices
  - Secret rotation procedures
  - Backup encryption guide
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6_

- [x] 14. Create quick start guide
  - One-command deployment
  - Minimal configuration example
  - Testing checklist
  - Common commands reference
  - _Requirements: 10.1, 10.5_

- [x] 15. Final checkpoint - Test complete deployment flow
  - Test fresh deployment on clean VPS
  - Verify all containers start successfully
  - Test application functionality
  - Verify data persistence
  - Test backup and restore
  - Test update deployment
  - Test rollback procedure
  - Verify logs are accessible
  - Test with IP address access
  - Document any issues found
  - _Requirements: All_

## Notes

- All scripts should be executable (chmod +x)
- Scripts should have proper error handling and logging
- Environment variables should never be hardcoded
- Support both IP and domain-based access
- SSL configuration is optional but documented
- All sensitive data should be in .env files (not committed)
- Backup before any destructive operations
- Health checks should be comprehensive
- Documentation should be clear and beginner-friendly
