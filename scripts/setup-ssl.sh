#!/bin/bash

# ================================================================
# SSL Setup Script
# ================================================================
# This script helps set up SSL/HTTPS using Let's Encrypt
# ================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}   SSL/HTTPS Setup with Let's Encrypt${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_error "This script must be run as root (use sudo)"
    exit 1
fi

# Step 1: Check prerequisites
print_info "Step 1: Checking prerequisites..."

# Check if certbot is installed
if ! command -v certbot &> /dev/null; then
    print_warning "Certbot is not installed"
    read -p "Install certbot now? (Y/n): " -n 1 -r
    echo ""
    
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        print_info "Installing certbot..."
        
        if command -v apt-get &> /dev/null; then
            apt-get update
            apt-get install -y certbot
        elif command -v yum &> /dev/null; then
            yum install -y certbot
        else
            print_error "Unable to install certbot automatically"
            echo "Please install certbot manually: https://certbot.eff.org/"
            exit 1
        fi
        
        print_success "Certbot installed"
    else
        exit 1
    fi
fi

print_success "Certbot is installed"
echo ""

# Step 2: Get domain information
print_info "Step 2: Domain Configuration"
echo ""

read -p "Enter your domain name (e.g., accounting.company.com): " DOMAIN

if [ -z "$DOMAIN" ]; then
    print_error "Domain name is required"
    exit 1
fi

read -p "Add www subdomain? (Y/n): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    DOMAINS="-d $DOMAIN -d www.$DOMAIN"
else
    DOMAINS="-d $DOMAIN"
fi

read -p "Enter your email address: " EMAIL

if [ -z "$EMAIL" ]; then
    print_error "Email address is required"
    exit 1
fi

print_success "Domain: $DOMAIN"
echo ""

# Step 3: Stop nginx temporarily
print_info "Step 3: Stopping nginx temporarily..."

cd "$PROJECT_ROOT"
docker-compose stop nginx

print_success "Nginx stopped"
echo ""

# Step 4: Obtain SSL certificate
print_info "Step 4: Obtaining SSL certificate..."

print_warning "This will use Let's Encrypt to obtain a free SSL certificate"
echo ""

if certbot certonly --standalone \
    $DOMAINS \
    --email "$EMAIL" \
    --agree-tos \
    --no-eff-email \
    --non-interactive; then
    print_success "SSL certificate obtained successfully"
else
    print_error "Failed to obtain SSL certificate"
    docker-compose start nginx
    exit 1
fi

echo ""

# Step 5: Copy certificates to project
print_info "Step 5: Copying certificates to project..."

mkdir -p nginx/ssl

cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" nginx/ssl/
cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" nginx/ssl/

chmod 644 nginx/ssl/fullchain.pem
chmod 600 nginx/ssl/privkey.pem
chown -R $SUDO_USER:$SUDO_USER nginx/ssl/

print_success "Certificates copied"
echo ""

# Step 6: Update nginx configuration
print_info "Step 6: Updating nginx configuration..."

# Backup current config
cp nginx/conf.d/default.conf nginx/conf.d/default.conf.backup

# Update server_name in HTTPS block
sed -i "s/server_name your-domain.com www.your-domain.com;/server_name $DOMAIN www.$DOMAIN;/g" nginx/conf.d/default.conf

# Uncomment HTTPS server block
sed -i '/# server {/,/# }/s/^# //' nginx/conf.d/default.conf

print_success "Nginx configuration updated"
echo ""

# Step 7: Update environment variables
print_info "Step 7: Updating environment variables..."

if [ -f ".env.production" ]; then
    # Backup current .env
    cp .env.production .env.production.backup
    
    # Update SERVER_HOST to use HTTPS
    sed -i "s|SERVER_HOST=http://.*|SERVER_HOST=https://$DOMAIN|g" .env.production
    sed -i "s|ALLOWED_ORIGINS=.*|ALLOWED_ORIGINS=https://$DOMAIN,https://www.$DOMAIN|g" .env.production
    sed -i "s|NEXT_PUBLIC_API_URL=.*|NEXT_PUBLIC_API_URL=https://$DOMAIN/api|g" .env.production
    sed -i "s|ENABLE_SSL=false|ENABLE_SSL=true|g" .env.production
    
    print_success "Environment variables updated"
else
    print_warning ".env.production not found, skipping"
fi

echo ""

# Step 8: Restart containers
print_info "Step 8: Restarting containers..."

docker-compose down
docker-compose up -d

print_success "Containers restarted"
echo ""

# Step 9: Wait for services
print_info "Step 9: Waiting for services to be healthy..."

sleep 15

# Step 10: Test HTTPS
print_info "Step 10: Testing HTTPS..."

if curl -sf "https://$DOMAIN/health" > /dev/null 2>&1; then
    print_success "HTTPS is working!"
else
    print_warning "HTTPS test failed, but this might be temporary"
    print_info "Check logs with: docker-compose logs nginx"
fi

echo ""

# Step 11: Set up auto-renewal
print_info "Step 11: Setting up auto-renewal..."

# Create renewal hook script
cat > /usr/local/bin/renew-ssl-certs.sh << EOF
#!/bin/bash
# SSL Certificate Renewal Hook

DOMAIN="$DOMAIN"
PROJECT_DIR="$PROJECT_ROOT"

# Copy new certificates
cp "/etc/letsencrypt/live/\$DOMAIN/fullchain.pem" "\$PROJECT_DIR/nginx/ssl/"
cp "/etc/letsencrypt/live/\$DOMAIN/privkey.pem" "\$PROJECT_DIR/nginx/ssl/"

# Set permissions
chmod 644 "\$PROJECT_DIR/nginx/ssl/fullchain.pem"
chmod 600 "\$PROJECT_DIR/nginx/ssl/privkey.pem"
chown -R $SUDO_USER:$SUDO_USER "\$PROJECT_DIR/nginx/ssl/"

# Reload nginx
cd "\$PROJECT_DIR"
docker-compose exec nginx nginx -s reload

echo "SSL certificates renewed and nginx reloaded"
EOF

chmod +x /usr/local/bin/renew-ssl-certs.sh

# Add cron job for auto-renewal
CRON_JOB="0 0,12 * * * certbot renew --quiet --deploy-hook '/usr/local/bin/renew-ssl-certs.sh'"

(crontab -l 2>/dev/null | grep -v "certbot renew"; echo "$CRON_JOB") | crontab -

print_success "Auto-renewal configured"
echo ""

# Test renewal
print_info "Testing renewal process..."

if certbot renew --dry-run; then
    print_success "Renewal test passed"
else
    print_warning "Renewal test failed, but certificates are installed"
fi

echo ""

# Display summary
echo -e "${GREEN}================================================================${NC}"
echo -e "${GREEN}   SSL Setup Complete!${NC}"
echo -e "${GREEN}================================================================${NC}"
echo ""
echo -e "${BLUE}SSL Information:${NC}"
echo -e "  Domain:           ${GREEN}$DOMAIN${NC}"
echo -e "  Certificate:      ${GREEN}/etc/letsencrypt/live/$DOMAIN/${NC}"
echo -e "  Auto-renewal:     ${GREEN}Enabled (twice daily)${NC}"
echo ""
echo -e "${BLUE}Your application is now accessible at:${NC}"
echo -e "  ${GREEN}https://$DOMAIN${NC}"
echo ""
echo -e "${BLUE}Certificate expires in:${NC} ${YELLOW}90 days${NC}"
echo -e "${BLUE}Auto-renewal will run:${NC} ${YELLOW}Twice daily${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Test your site: https://$DOMAIN"
echo "  2. Check SSL grade: https://www.ssllabs.com/ssltest/analyze.html?d=$DOMAIN"
echo "  3. Update any external links to use HTTPS"
echo ""

print_success "SSL setup complete!"
