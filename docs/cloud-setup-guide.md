# Cloud Services Setup Guide

This guide walks you through setting up Supabase (PostgreSQL) and Confluent Cloud (Kafka) for the Connect 4 multiplayer system.

## Prerequisites

- Supabase account (free tier available)
- Confluent Cloud account (free tier available)
- Basic understanding of environment variables

## 1. Supabase Setup (PostgreSQL Database)

### Step 1: Create Supabase Project

1. Go to [supabase.com](https://supabase.com) and sign up/login
2. Click "New Project"
3. Choose your organization
4. Fill in project details:
   - **Name**: `connect4-multiplayer`
   - **Database Password**: Generate a strong password (save this!)
   - **Region**: Choose closest to your users
   - **Pricing Plan**: Free tier is sufficient for development

### Step 2: Get Database Credentials

1. Go to **Settings** → **Database**
2. Find the **Connection string** section
3. Copy the **URI** format connection string
4. It should look like: `postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres`

### Step 3: Configure Database Access

1. Go to **Settings** → **Database** → **Connection pooling**
2. Enable connection pooling (recommended for production)
3. Note the pooled connection string for production use

### Step 4: Set Up Database Schema

The application will automatically run migrations, but you can also run them manually:

```sql
-- Run these in the Supabase SQL Editor if needed
-- The migrations in ./migrations/ will be applied automatically
```

## 2. Confluent Cloud Setup (Kafka)

### Step 1: Create Confluent Cloud Account

1. Go to [confluent.cloud](https://confluent.cloud) and sign up
2. Choose the **Free** plan to start

### Step 2: Create a Cluster

1. Click **Create cluster**
2. Choose **Basic** cluster (free tier)
3. Select your preferred cloud provider and region
4. Name your cluster: `connect4-cluster`
5. Wait for cluster creation (2-3 minutes)

### Step 3: Create API Keys

1. Go to **Cluster Overview** → **API Keys**
2. Click **Create key**
3. Choose **Global access** (or scope to specific topics if preferred)
4. **Save the API Key and Secret** - you won't see the secret again!

### Step 4: Create Topics

1. Go to **Topics** in your cluster
2. Create a new topic: `game-events`
3. Use default settings (1 partition, cleanup policy: delete)

### Step 5: Get Bootstrap Servers

1. Go to **Cluster Overview** → **Cluster settings**
2. Copy the **Bootstrap server** URL
3. It should look like: `pkc-xxxxx.region.provider.confluent.cloud:9092`

## 3. Environment Configuration

### For Local Development

Create a `.env` file in your project root:

```bash
# Copy from .env.example and update these values

# Application Environment
ENVIRONMENT=development

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_CORS_ORIGINS=http://localhost:3000,http://localhost:8080

# Supabase Database Configuration
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_CONN_MAX_LIFETIME=300

# Confluent Cloud Kafka Configuration
KAFKA_BOOTSTRAP_SERVERS=pkc-xxxxx.region.provider.confluent.cloud:9092
KAFKA_API_KEY=your-confluent-api-key
KAFKA_API_SECRET=your-confluent-api-secret
KAFKA_TOPIC=game-events
KAFKA_CONSUMER_GROUP=analytics-service

# Redis Configuration (local or cloud)
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### For Production Deployment

Set these environment variables in your deployment platform:

```bash
# Production Environment Variables
ENVIRONMENT=production
DATABASE_URL=postgresql://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require
KAFKA_BOOTSTRAP_SERVERS=pkc-xxxxx.region.provider.confluent.cloud:9092
KAFKA_API_KEY=your-confluent-api-key
KAFKA_API_SECRET=your-confluent-api-secret
SERVER_CORS_ORIGINS=https://your-frontend-domain.com
```

## 4. Testing the Connection

### Test Database Connection

```bash
# Run database migrations
go run cmd/migrate/main.go

# Test connection with a simple query
psql "postgresql://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require" -c "SELECT version();"
```

### Test Kafka Connection

```bash
# Install confluent CLI (optional)
curl -sL --http1.1 https://cnfl.io/cli | sh -s -- latest

# Test connection (replace with your values)
confluent kafka topic list --cluster your-cluster-id
```

### Run the Application

```bash
# Start the game server
go run cmd/server/main.go

# Start the analytics service
go run cmd/analytics/main.go
```

## 5. Deployment Platforms

### Railway

1. Connect your GitHub repository
2. Set environment variables in Railway dashboard
3. Deploy automatically on push

### Render

1. Create new Web Service
2. Connect GitHub repository
3. Set environment variables
4. Deploy

### Fly.io

1. Install flyctl CLI
2. Run `fly launch` in project directory
3. Set secrets: `fly secrets set DATABASE_URL=...`
4. Deploy with `fly deploy`

## 6. Monitoring and Troubleshooting

### Supabase Monitoring

- Check **Logs** in Supabase dashboard
- Monitor **Database** → **Usage** for connection limits
- Use **SQL Editor** for manual queries

### Confluent Cloud Monitoring

- Check **Cluster Overview** for health
- Monitor **Topics** for message throughput
- Check **Connectors** if using any

### Application Logs

```bash
# Check application logs
docker-compose logs game-server
docker-compose logs analytics-service

# Or if running directly
go run cmd/server/main.go 2>&1 | tee server.log
```

## 7. Security Best Practices

### Database Security

- Use connection pooling in production
- Enable SSL/TLS (sslmode=require)
- Rotate passwords regularly
- Use read-only users for analytics queries

### Kafka Security

- Rotate API keys regularly
- Use topic-level ACLs if needed
- Monitor usage to avoid quota limits
- Enable audit logging in production

### Environment Variables

- Never commit `.env` files to git
- Use secret management in production
- Rotate credentials regularly
- Use least-privilege access

## 8. Cost Optimization

### Supabase Free Tier Limits

- 500MB database storage
- 2GB bandwidth per month
- 50MB file storage
- Upgrade to Pro ($25/month) when needed

### Confluent Cloud Free Tier

- $400 free credits
- Basic cluster included
- Monitor usage in dashboard
- Set up billing alerts

## 9. Backup and Recovery

### Database Backups

- Supabase Pro includes automatic backups
- Export data regularly: `pg_dump`
- Test restore procedures

### Kafka Topic Backup

- Confluent Cloud handles replication
- Consider topic mirroring for critical data
- Export important events to long-term storage

## Next Steps

After setting up credentials:

1. Update your `.env` file with real credentials
2. Test local development setup
3. Run database migrations
4. Test the full application flow
5. Set up production deployment
6. Configure monitoring and alerts

## Support

- **Supabase**: [docs.supabase.com](https://docs.supabase.com)
- **Confluent Cloud**: [docs.confluent.io](https://docs.confluent.io)
- **Project Issues**: Create GitHub issues for project-specific problems