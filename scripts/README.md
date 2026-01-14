# Deployment Scripts

Automated scripts for deploying and managing the application on VPS.

## Scripts Overview

### 🚀 Deployment Scripts

#### `setup.sh` - Initial Setup
First-time deployment script.

```bash
./scripts/setup.sh
```

**What it does:**
- Checks system requirements
- Creates environment configuration
- Builds Docker images
- Starts containers
- Displays access information

**When to use:** First deployment to a new VPS

---

#### `deploy.sh` - Update/Redeploy
Updates the application with zero downtime.

```bash
./scripts/deploy.sh
```

**What it does:**
- Creates backup before update
- Pulls latest code from Git
- Rebuilds Docker images if needed
- Performs rolling update
- Runs database migrations
- Health check verification
- Rollback on failure

**When to use:** Updating to a new version

---

#### `backup.sh` - Create Backup
Backs up database and uploaded files.

```bash
./scripts/backup.sh
```

**What it does:**
- Dumps PostgreSQL database
- Archives uploaded files
- Saves configuration
- Compresses backup
- Cleans old backups

**When to use:** Before major changes, or scheduled via cron

---

#### `rollback.sh` - Restore from Backup
Restores application from a backup.

```bash
./scripts/rollback.sh
```

**What it does:**
- Lists available backups
- Restores database
- Restores uploaded files
- Restores configuration (optional)
- Restarts containers

**When to use:** After failed deployment or data loss

---

### 📊 Monitoring Scripts

#### `health-check.sh` - System Health Check
Checks health of all services.

```bash
./scripts/health-check.sh
```

**What it checks:**
- Container status
- Database connectivity
- Backend API health
- Frontend accessibility
- Disk space usage
- Memory usage
- Docker volumes

**When to use:** Regular monitoring, troubleshooting

---

#### `logs.sh` - View Logs
View and search container logs.

```bash
# View all logs
./scripts/logs.sh

# View specific service
./scripts/logs.sh backend

# Follow logs in real-time
./scripts/logs.sh -f

# Show last 50 lines
./scripts/logs.sh -n 50

# Search for errors
./scripts/logs.sh -s "error"

# Follow backend logs with search
./scripts/logs.sh -f -s "database" backend
```

**Options:**
- `-f, --follow` - Follow log output (live tail)
- `-n, --lines NUM` - Number of lines to show
- `-s, --search TEXT` - Search for specific text
- `-h, --help` - Show help message

**When to use:** Debugging, monitoring, troubleshooting

---

### ⚙️ Setup Scripts

#### `setup-env.sh` - Environment Configuration
Interactive environment setup.

```bash
./scripts/setup-env.sh
```

**What it does:**
- Prompts for server address
- Configures database credentials
- Generates strong secrets
- Creates `.env.production`

**When to use:** Initial setup, reconfiguration

---

#### `setup-ssl.sh` - SSL/HTTPS Setup
Sets up SSL certificates using Let's Encrypt.

```bash
sudo ./scripts/setup-ssl.sh
```

**What it does:**
- Installs certbot
- Obtains SSL certificate
- Updates nginx configuration
- Configures auto-renewal
- Restarts containers

**Requirements:**
- Domain name pointing to VPS
- Run as root (sudo)

**When to use:** Setting up HTTPS for production

---

## Quick Reference

### Common Workflows

#### First-Time Deployment
```bash
./scripts/setup.sh
```

#### Update Application
```bash
./scripts/deploy.sh
```

#### Check System Health
```bash
./scripts/health-check.sh
```

#### View Logs
```bash
./scripts/logs.sh -f
```

#### Create Backup
```bash
./scripts/backup.sh
```

#### Restore from Backup
```bash
./scripts/rollback.sh
```

#### Setup SSL
```bash
sudo ./scripts/setup-ssl.sh
```

---

## Automated Backups

Add to crontab for daily backups:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * cd /path/to/your/app && ./scripts/backup.sh >> /var/log/backup.log 2>&1
```

---

## Troubleshooting

### Scripts Not Executable

```bash
chmod +x scripts/*.sh
```

### Permission Denied

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and log back in
```

### Script Fails

```bash
# Check logs
./scripts/logs.sh

# Check health
./scripts/health-check.sh

# View script output
bash -x ./scripts/script-name.sh
```

---

## Environment Variables

All scripts use `.env.production` for configuration.

Key variables:
- `SERVER_HOST` - Your VPS IP or domain
- `POSTGRES_PASSWORD` - Database password
- `JWT_SECRET` - JWT secret key
- `BACKUP_DIR` - Backup directory path
- `BACKUP_RETENTION_DAYS` - How long to keep backups

See [ENVIRONMENT_VARIABLES.md](../ENVIRONMENT_VARIABLES.md) for complete list.

---

## Security Notes

1. **Never commit `.env.production`** - Contains secrets
2. **Use strong passwords** - Min 20 characters
3. **Rotate secrets regularly** - Every 90 days
4. **Secure backup directory** - Restrict permissions
5. **Monitor logs** - Check for suspicious activity

---

## Support

For issues:
1. Check [DEPLOYMENT.md](../DEPLOYMENT.md)
2. Run `./scripts/health-check.sh`
3. View logs with `./scripts/logs.sh`
4. Check [TROUBLESHOOTING.md](../TROUBLESHOOTING.md)

---

**Last Updated**: January 2026
