# SSL/HTTPS Setup Guide

This guide explains how to set up SSL/HTTPS for your application using Let's Encrypt (free SSL certificates).

## Prerequisites

- Domain name pointing to your VPS IP address
- Application already running on HTTP (port 80)
- Root or sudo access to VPS

## Option 1: Automatic Setup with Certbot (Recommended)

### Step 1: Install Certbot

```bash
# On Ubuntu/Debian
sudo apt update
sudo apt install certbot

# On CentOS/RHEL
sudo yum install certbot
```

### Step 2: Stop Nginx Container Temporarily

```bash
cd /path/to/your/app
docker-compose stop nginx
```

### Step 3: Obtain SSL Certificate

```bash
# Replace your-domain.com with your actual domain
sudo certbot certonly --standalone \
  -d your-domain.com \
  -d www.your-domain.com \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email
```

### Step 4: Copy Certificates to Project

```bash
# Create SSL directory
mkdir -p nginx/ssl

# Copy certificates
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem nginx/ssl/
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem nginx/ssl/

# Set permissions
sudo chmod 644 nginx/ssl/fullchain.pem
sudo chmod 600 nginx/ssl/privkey.pem
sudo chown $USER:$USER nginx/ssl/*
```

### Step 5: Update Nginx Configuration

Edit `nginx/conf.d/default.conf`:

1. Uncomment the HTTPS server block (lines starting with # server {)
2. Replace `your-domain.com` with your actual domain
3. Uncomment the HTTP to HTTPS redirect block at the bottom

### Step 6: Update Environment Variables

Edit `.env.production`:

```bash
# Change from HTTP to HTTPS
SERVER_HOST=https://your-domain.com
ALLOWED_ORIGINS=https://your-domain.com,https://www.your-domain.com
NEXT_PUBLIC_API_URL=https://your-domain.com/api
ENABLE_SSL=true
```

### Step 7: Restart Containers

```bash
docker-compose down
docker-compose up -d
```

### Step 8: Test HTTPS

Visit `https://your-domain.com` in your browser. You should see a secure connection (padlock icon).

### Step 9: Set Up Auto-Renewal

Let's Encrypt certificates expire after 90 days. Set up automatic renewal:

```bash
# Test renewal
sudo certbot renew --dry-run

# Add cron job for auto-renewal
sudo crontab -e

# Add this line (runs twice daily):
0 0,12 * * * certbot renew --quiet --deploy-hook "cd /path/to/your/app && ./scripts/update-ssl-certs.sh"
```

Create renewal hook script:

```bash
cat > scripts/update-ssl-certs.sh << 'EOF'
#!/bin/bash
# Update SSL certificates in Docker volume

DOMAIN="your-domain.com"
SSL_DIR="nginx/ssl"

# Copy new certificates
sudo cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem $SSL_DIR/
sudo cp /etc/letsencrypt/live/$DOMAIN/privkey.pem $SSL_DIR/

# Set permissions
sudo chmod 644 $SSL_DIR/fullchain.pem
sudo chmod 600 $SSL_DIR/privkey.pem
sudo chown $USER:$USER $SSL_DIR/*

# Reload nginx
docker-compose exec nginx nginx -s reload

echo "SSL certificates updated successfully"
EOF

chmod +x scripts/update-ssl-certs.sh
```

## Option 2: Manual Setup with Existing Certificates

If you already have SSL certificates from another provider:

### Step 1: Prepare Certificates

You need two files:
- `fullchain.pem` (certificate + intermediate certificates)
- `privkey.pem` (private key)

### Step 2: Copy to Project

```bash
mkdir -p nginx/ssl
cp /path/to/your/fullchain.pem nginx/ssl/
cp /path/to/your/privkey.pem nginx/ssl/
chmod 644 nginx/ssl/fullchain.pem
chmod 600 nginx/ssl/privkey.pem
```

### Step 3: Follow Steps 5-7 from Option 1

## Testing SSL Configuration

### Test SSL Certificate

```bash
# Check certificate validity
openssl s_client -connect your-domain.com:443 -servername your-domain.com

# Check certificate expiration
echo | openssl s_client -connect your-domain.com:443 -servername your-domain.com 2>/dev/null | openssl x509 -noout -dates
```

### Test SSL Grade

Visit: https://www.ssllabs.com/ssltest/analyze.html?d=your-domain.com

Aim for A+ rating.

## Troubleshooting

### Certificate Not Found

```bash
# Check if certificates exist
ls -la nginx/ssl/

# Check nginx logs
docker-compose logs nginx
```

### Permission Denied

```bash
# Fix permissions
sudo chown -R $USER:$USER nginx/ssl/
chmod 644 nginx/ssl/fullchain.pem
chmod 600 nginx/ssl/privkey.pem
```

### Mixed Content Warnings

Update all API calls in frontend to use HTTPS:
- Check `NEXT_PUBLIC_API_URL` in `.env.production`
- Ensure it starts with `https://`

### Certificate Expired

```bash
# Renew certificate
sudo certbot renew

# Update certificates in Docker
./scripts/update-ssl-certs.sh
```

## Security Best Practices

1. **Use Strong Ciphers**: The default configuration uses modern, secure ciphers
2. **Enable HSTS**: Already configured in the HTTPS server block
3. **Redirect HTTP to HTTPS**: Uncomment the redirect block
4. **Keep Certificates Updated**: Set up auto-renewal
5. **Monitor Expiration**: Set up alerts 30 days before expiration
6. **Use HTTP/2**: Already enabled with `http2` directive
7. **Disable Weak Protocols**: Only TLS 1.2 and 1.3 are enabled

## Migration from HTTP to HTTPS

### Zero-Downtime Migration

1. **Keep HTTP running** while setting up HTTPS
2. **Test HTTPS** thoroughly before redirecting
3. **Update environment variables** to support both HTTP and HTTPS:
   ```bash
   ALLOWED_ORIGINS=http://your-domain.com,https://your-domain.com
   ```
4. **Enable HTTP to HTTPS redirect** only after testing
5. **Update all external links** to use HTTPS

### Rollback Plan

If issues occur:

1. Comment out HTTPS server block in `nginx/conf.d/default.conf`
2. Comment out HTTP to HTTPS redirect
3. Restart nginx: `docker-compose restart nginx`
4. Revert environment variables to HTTP

## Additional Resources

- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [Certbot Documentation](https://certbot.eff.org/docs/)
- [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/)
- [SSL Labs Testing Tool](https://www.ssllabs.com/ssltest/)

## Support

For issues with SSL setup, check:
1. Domain DNS is pointing to correct IP
2. Firewall allows ports 80 and 443
3. Nginx logs: `docker-compose logs nginx`
4. Certificate files exist and have correct permissions
