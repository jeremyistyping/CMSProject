# Deployment Guide - VPS Docker Deployment

Complete guide for deploying the Accounting System to a VPS using Docker.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Configuration](#configuration)
4. [Deployment](#deployment)
5. [Post-Deployment](#post-deployment)
6. [Updating](#updating)
7. [Backup & Restore](#backup--restore)
8. [Troubleshooting](#troubleshooting)

## Prerequisites

### VPS Requirements

- **Operating System**: Ubuntu 20.04+ or Debian 11+ (recommended)
- **RAM**: Minimum 2GB, recommended 4GB+
- **Disk Space**: Minimum 20GB free space
- **CPU**: 2+ cores recommended
- **Network**: Public IP address
- **Ports**: 80 (HTTP) and 443 (HTTPS) open

### Software Requirements

- Docker 20.10+
- Docker Compose 2.0+
- Git
- OpenSSL (for generating secrets)

### Domain (Optional)

- Domain name pointing to your VPS IP (for HTTPS/SSL)
- DNS A record configured

## Initial Setup

### Step 1: Install Docker

```bash
# Update system
sudo apt update
sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add your user to docker group
sudo usermod -aG docker $USER

# Log out and log back in for group changes to take effect
```

### Step 2: Install Docker Compose

```bash
# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker-compose --version
```

### Step 3: Clone Repository

```bash
# Clone your repository
git clone https://github.com/your-username/your-repo.git
cd your-repo

# Or if already cloned, pull latest
git pull origin main
```

## Configuration

### Option 1: Automated Setup (Recommended)

Run the setup script which will guide you through configuration:

```bash
./scripts/setup.sh
```

The script will:
- Check system requirements
- Create `.env.production` interactively
- Build Docker images
- Start containers
- Display access information

### Option 2: Manual Configuration

1. **Copy environment template**:
   ```bash
   cp .env.production.example .env.production
   ```

2. **Edit `.env.production`**:
   ```bash
   nano .env.production
   ```

3. **Update these values**:
   ```bash
   # Replace YOUR_VPS_IP_HERE with your actual IP
   SERVER_HOST=http://YOUR_VPS_IP
   
   # Strong database password (min 20 characters)
   POSTGRES_PASSWORD=your_strong_password_here
   
   # Strong JWT secret (min 32 characters)
   JWT_SECRET=your_jwt_secret_here
   
   # Update CORS and API URL
   ALLOWED_ORIGINS=http://YOUR_VPS_IP
   NEXT_PUBLIC_API_URL=http://YOUR_VPS_IP/api
   ```

4. **Generate strong secrets**:
   ```bash
   # Generate database password
   openssl rand -base64 32
   
   # Generate JWT secret
   openssl rand -base64 32
   ```

## Deployment

### First-Time Deployment

```bash
# Make scripts executable (if not already)
chmod +x scripts/*.sh

# Run setup script
./scripts/setup.sh
```

Or manually:

```bash
# Build images
docker-compose build

# Start containers
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### Verify Deployment

1. **Check container health**:
   ```bash
   ./scripts/health-check.sh
   ```

2. **Access application**:
   - Open browser: `http://YOUR_VPS_IP`
   - Should see the login page

3. **Check logs**:
   ```bash
   ./scripts/logs.sh
   ```

## Post-Deployment

### 1. Test Application

- Login with default credentials (if applicable)
- Test core functionality
- Upload a test file
- Create test data

### 2. Set Up SSL (If you have a domain)

```bash
# Run SSL setup script
sudo ./scripts/setup-ssl.sh
```

See [nginx/SSL_SETUP.md](nginx/SSL_SETUP.md) for detailed SSL setup instructions.

### 3. Configure Firewall

```bash
# Allow SSH
sudo ufw allow 22/tcp

# Allow HTTP
sudo ufw allow 80/tcp

# Allow HTTPS
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable
```

### 4. Set Up Automated Backups

Add to crontab:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * cd /path/to/your/app && ./scripts/backup.sh >> /var/log/backup.log 2>&1
```

### 5. Set Up Monitoring

Consider setting up:
- Uptime monitoring (e.g., UptimeRobot)
- Log aggregation
- Resource monitoring
- Alert notifications

## Updating

### Update Application

```bash
# Run deployment script
./scripts/deploy.sh
```

The script will:
1. Create backup
2. Pull latest code
3. Rebuild if needed
4. Perform rolling update
5. Run migrations
6. Health check

### Manual Update

```bash
# Backup first
./scripts/backup.sh

# Pull latest code
git pull origin main

# Rebuild and restart
docker-compose build
docker-compose up -d

# Check health
./scripts/health-check.sh
```

## Backup & Restore

### Create Backup

```bash
# Manual backup
./scripts/backup.sh
```

Backup includes:
- Database dump
- Uploaded files
- Configuration
- Git commit info

### Restore from Backup

```bash
# List available backups and restore
./scripts/rollback.sh

# Or restore specific backup
./scripts/rollback.sh backup_20240115_120000
```

### Backup Location

Default: `./backups/`

Backups are compressed as `.tar.gz` files with timestamp.

## Troubleshooting

### Common Issues

#### 1. Containers Won't Start

```bash
# Check logs
docker-compose logs

# Check specific service
docker-compose logs backend

# Restart services
docker-compose restart
```

#### 2. Database Connection Failed

```bash
# Check postgres is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Verify credentials in .env.production
```

#### 3. Port Already in Use

```bash
# Check what's using port 80
sudo lsof -i :80

# Stop conflicting service
sudo systemctl stop apache2  # or nginx, etc.
```

#### 4. Permission Denied

```bash
# Fix script permissions
chmod +x scripts/*.sh

# Fix Docker permissions
sudo usermod -aG docker $USER
# Log out and log back in
```

#### 5. Out of Disk Space

```bash
# Check disk usage
df -h

# Clean Docker
docker system prune -a

# Clean old backups
find ./backups -name "backup_*.tar.gz" -mtime +30 -delete
```

### Useful Commands

```bash
# View all logs
./scripts/logs.sh

# View specific service logs
./scripts/logs.sh backend

# Follow logs in real-time
./scripts/logs.sh -f

# Search logs
./scripts/logs.sh -s "error"

# Check system health
./scripts/health-check.sh

# Restart all services
docker-compose restart

# Stop all services
docker-compose stop

# Start all services
docker-compose start

# View resource usage
docker stats

# Access container shell
docker-compose exec backend sh
docker-compose exec frontend sh
```

### Getting Help

1. Check logs: `./scripts/logs.sh`
2. Check health: `./scripts/health-check.sh`
3. Review [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
4. Check container status: `docker-compose ps`
5. Verify configuration: `cat .env.production`

## Security Best Practices

1. **Use strong passwords** (min 20 characters)
2. **Enable firewall** (ufw or iptables)
3. **Set up SSL/HTTPS** for production
4. **Regular updates** (system and application)
5. **Regular backups** (automated daily)
6. **Limit SSH access** (key-based auth, disable root)
7. **Monitor logs** for suspicious activity
8. **Keep secrets secure** (never commit .env.production)
9. **Use fail2ban** to prevent brute force attacks
10. **Regular security audits**

## Performance Optimization

1. **Enable caching** in nginx
2. **Optimize database** queries and indexes
3. **Monitor resource usage** (CPU, RAM, disk)
4. **Scale horizontally** if needed (multiple instances)
5. **Use CDN** for static assets
6. **Enable compression** (gzip)
7. **Optimize images** before upload
8. **Regular database maintenance** (VACUUM, ANALYZE)

## Maintenance Schedule

### Daily
- Automated backups
- Log rotation
- Health checks

### Weekly
- Review logs for errors
- Check disk space
- Update system packages

### Monthly
- Security updates
- Database optimization
- Backup verification (test restore)
- SSL certificate check

### Quarterly
- Full system audit
- Performance review
- Capacity planning
- Disaster recovery test

## Additional Resources

- [Environment Variables Guide](ENVIRONMENT_VARIABLES.md)
- [Architecture Documentation](ARCHITECTURE.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [SSL Setup Guide](nginx/SSL_SETUP.md)
- [Migration Guide](MIGRATION_GUIDE.md)

## Support

For issues or questions:
1. Check documentation
2. Review logs
3. Search existing issues
4. Create new issue with details

---

**Last Updated**: January 2026
**Version**: 1.0.0
