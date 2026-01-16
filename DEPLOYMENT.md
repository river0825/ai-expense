# Deployment Guide

This guide covers deploying AIExpense to production.

## Prerequisites

- Go 1.21+ (for building from source)
- Docker (for containerized deployment)
- LINE Messaging API credentials
- Google Gemini API key
- A server with public IP (for LINE webhook)

## Local Development

### Setup

```bash
# Clone repository
git clone <repo>
cd aiexpense

# Create .env file from template
cp .env.example .env

# Edit .env with your credentials
vim .env

# Install dependencies
go mod download
```

### Running Locally

```bash
# Load environment variables
source .env

# Run the server
go run ./cmd/server
```

Server runs on `http://localhost:8080`

### Testing Locally

```bash
# Health check
curl http://localhost:8080/health

# Auto-signup
curl -X POST http://localhost:8080/api/users/auto-signup \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test_user_123",
    "messenger_type": "line"
  }'

# Parse expenses
curl -X POST http://localhost:8080/api/expenses/parse \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test_user_123",
    "text": "早餐$20午餐$30"
  }'
```

## Docker Deployment

### Build Docker Image

```bash
docker build -t aiexpense:latest .
```

### Run with Docker Compose

```bash
# Copy and edit environment
cp .env.example .env
vim .env

# Start services
docker-compose up -d

# View logs
docker-compose logs -f aiexpense

# Stop services
docker-compose down
```

### Run Single Container

```bash
docker run -d \
  --name aiexpense \
  -p 8080:8080 \
  -v aiexpense_data:/data \
  -e LINE_CHANNEL_TOKEN=<token> \
  -e LINE_CHANNEL_ID=<id> \
  -e GEMINI_API_KEY=<key> \
  -e ADMIN_API_KEY=<optional_key> \
  aiexpense:latest
```

## Production Deployment

### Infrastructure Setup

1. **Server Requirements**
   - Minimum: 1 CPU, 512MB RAM, 1GB storage
   - Recommended: 2 CPU, 2GB RAM, 10GB storage
   - Ubuntu 20.04+ or similar Linux distro

2. **Network Configuration**
   - Public IP address
   - Firewall: Allow ports 80, 443, 8080
   - SSL/TLS certificate (for production)

3. **Database**
   - SQLite is embedded in the binary (zero setup)
   - Persistent storage: Mount `/data` volume

### Deployment Options

#### Option 1: Docker on VPS

```bash
# SSH into server
ssh user@server_ip

# Clone repository
git clone <repo>
cd aiexpense

# Setup environment
cp .env.example .env
vim .env  # Edit credentials

# Start with docker-compose
docker-compose up -d

# Verify health
curl http://localhost:8080/health
```

#### Option 2: Systemd Service

```bash
# Build binary
CGO_ENABLED=1 go build -o /usr/local/bin/aiexpense ./cmd/server

# Create systemd service file
sudo tee /etc/systemd/system/aiexpense.service > /dev/null <<EOF
[Unit]
Description=AIExpense Service
After=network.target

[Service]
Type=simple
User=aiexpense
EnvironmentFile=/etc/aiexpense/.env
ExecStart=/usr/local/bin/aiexpense
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Create user and directories
sudo useradd -r aiexpense
sudo mkdir -p /etc/aiexpense /var/lib/aiexpense
sudo chown aiexpense:aiexpense /var/lib/aiexpense

# Edit environment
sudo vim /etc/aiexpense/.env

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable aiexpense
sudo systemctl start aiexpense

# Check status
sudo systemctl status aiexpense
```

#### Option 3: Cloud Platforms

##### Google Cloud Run

```bash
# Build and push to container registry
gcloud builds submit --tag gcr.io/<project>/aiexpense

# Deploy
gcloud run deploy aiexpense \
  --image gcr.io/<project>/aiexpense \
  --platform managed \
  --region us-central1 \
  --memory 512Mi \
  --set-env-vars LINE_CHANNEL_TOKEN=<token>,LINE_CHANNEL_ID=<id>,GEMINI_API_KEY=<key>
```

##### AWS ECS

```bash
# Create ECR repository
aws ecr create-repository --repository-name aiexpense

# Build and push
docker build -t aiexpense:latest .
docker tag aiexpense:latest <account>.dkr.ecr.<region>.amazonaws.com/aiexpense:latest
docker push <account>.dkr.ecr.<region>.amazonaws.com/aiexpense:latest

# Deploy via ECS Fargate console or CLI
```

### Reverse Proxy Setup (nginx)

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## LINE Webhook Configuration

1. Go to LINE Developers Console
2. Create a new Messaging API channel
3. Get your Channel Token and Channel ID
4. Set webhook URL to: `https://your-domain.com/webhook/line`
5. Enable webhook functionality
6. Test webhook connection

## Monitoring

### Health Checks

```bash
# Manual health check
curl https://your-domain.com/health

# With monitoring tools (Healthchecks.io, UptimeRobot, etc.)
# Configure to ping /health endpoint periodically
```

### Logs

Docker:
```bash
docker-compose logs -f aiexpense
```

Systemd:
```bash
sudo journalctl -u aiexpense -f
```

### Database Backups

SQLite database is stored at `/data/aiexpense.db`

```bash
# Manual backup
cp /data/aiexpense.db /backup/aiexpense-$(date +%Y%m%d-%H%M%S).db

# Automated backup (cron)
0 2 * * * cp /data/aiexpense.db /backup/aiexpense-$(date +\%Y\%m\%d-\%H\%M\%S).db
```

## Scaling

### Horizontal Scaling

For high volume, consider:

1. **Multiple instances** behind load balancer (nginx, HAProxy)
2. **Shared PostgreSQL** instead of SQLite
   - Update repository implementations to use PostgreSQL
   - Allows multiple server instances
   - Better concurrency handling

3. **Message queue** for async processing (optional)
   - Redis for caching parsed results
   - Job queue for heavy AI API calls

### Performance Tuning

1. **Database indexes** - Already configured for common queries
2. **Response caching** - Cache category lists, parsed results
3. **AI caching** - Avoid duplicate parsing of same text
4. **Connection pooling** - Configured in database/sql

## Troubleshooting

### Server won't start

```bash
# Check port is available
netstat -an | grep 8080

# Check database file is writable
ls -la /data/aiexpense.db

# Check environment variables
env | grep -E 'LINE_|GEMINI_'

# View error logs
docker-compose logs aiexpense
```

### LINE webhook not receiving events

1. Verify webhook URL is publicly accessible
2. Check LINE developer console webhook settings
3. Verify signature verification in webhook handler
4. Check server logs for webhook events

### Database errors

```bash
# Reset database (WARNING: loses all data)
rm /data/aiexpense.db

# Restart server to recreate schema
docker-compose restart aiexpense
```

## Security Checklist

- [ ] Set strong `ADMIN_API_KEY`
- [ ] Use HTTPS/SSL in production
- [ ] Keep LINE Channel Token secret (rotate regularly)
- [ ] Keep Gemini API Key secret
- [ ] Regular database backups
- [ ] Monitor for unusual API usage
- [ ] Keep Go dependencies updated
- [ ] Use firewall to restrict access
- [ ] Enable VPC if available
- [ ] Monitor server logs for errors

## Cost Optimization

### Gemini API

- Monitor API calls in Google Cloud Console
- Cache parsed results to avoid duplicate calls
- Use regex fallback for simple patterns
- Consider rate limiting for public APIs

### Infrastructure

- Start with small instance, scale as needed
- Use free tier of cloud services where available
- Monitor monthly bills
- Set up cost alerts

## Support

For deployment issues:
1. Check logs: `docker-compose logs`
2. Verify environment variables
3. Check LINE Messaging API credentials
4. Test endpoints with curl
5. Create issue with error details
