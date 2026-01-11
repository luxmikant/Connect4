# Quick Credential Setup Guide

This guide helps you quickly set up Supabase and Confluent Cloud credentials for the Connect 4 multiplayer system.

## 🚀 Quick Start

### 1. Create Environment File

```bash
# Create .env file from template
make create-env

# Or manually copy
cp .env.example .env
```

### 2. Interactive Setup (Recommended)

```bash
# Run interactive setup script
make setup-credentials
```

This will guide you through:
- Creating .env file
- Validating configuration
- Testing database connection
- Testing Kafka connection
- Running database migrations

### 3. Manual Setup

Edit your `.env` file with your actual credentials:

```bash
# Supabase Database
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@db.YOUR_PROJECT_REF.supabase.co:5432/postgres?sslmode=require

# Confluent Cloud Kafka (optional for local development)
KAFKA_BOOTSTRAP_SERVERS=pkc-xxxxx.region.provider.confluent.cloud:9092
KAFKA_API_KEY=your-confluent-api-key
KAFKA_API_SECRET=your-confluent-api-secret
```

## 🔧 Getting Your Credentials

### Supabase (Database)

1. **Sign up**: Go to [supabase.com](https://supabase.com)
2. **Create project**: Click "New Project"
3. **Get connection string**: 
   - Go to Settings → Database
   - Copy the URI connection string
   - Replace `[YOUR-PASSWORD]` with your database password

### Confluent Cloud (Kafka) - Optional

1. **Sign up**: Go to [confluent.cloud](https://confluent.cloud)
2. **Create cluster**: Choose "Basic" (free tier)
3. **Create API key**: 
   - Go to Cluster Overview → API Keys
   - Click "Create key"
   - Save the key and secret
4. **Get bootstrap servers**: 
   - Go to Cluster Overview → Cluster settings
   - Copy the Bootstrap server URL

## 🧪 Testing Your Setup

```bash
# Test database connection
make test-db

# Test Kafka connection
make test-kafka

# Validate all environment variables
make validate-env

# Run database migrations
make migrate
```

## 🐳 Local Development (No Cloud Setup Required)

For local development, you can use Docker Compose without cloud services:

```bash
# Start local services (PostgreSQL, Redis, Kafka)
make docker-up

# Use local configuration in .env
DATABASE_URL=postgres://postgres:password@localhost:5432/connect4?sslmode=disable
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
REDIS_URL=localhost:6379

# Run migrations
make migrate

# Start the application
make run-server
```

## 📋 Environment Variables Reference

### Required
- `DATABASE_URL` - PostgreSQL connection string
- `KAFKA_BOOTSTRAP_SERVERS` - Kafka bootstrap servers

### Optional (with defaults)
- `SERVER_PORT` - Server port (default: 8080)
- `KAFKA_API_KEY` - Confluent Cloud API key
- `KAFKA_API_SECRET` - Confluent Cloud API secret
- `REDIS_URL` - Redis connection (default: localhost:6379)

## 🔍 Troubleshooting

### Database Connection Issues
```bash
# Check if database is accessible
psql "your-database-url" -c "SELECT version();"

# Common issues:
# - Wrong password
# - Incorrect project reference
# - Network connectivity
# - SSL mode mismatch
```

### Kafka Connection Issues
```bash
# For Confluent Cloud, check:
# - API key and secret are correct
# - Bootstrap servers URL is correct
# - Cluster is running

# For local Kafka:
make docker-up  # Ensure Kafka is running
```

### Environment Variable Issues
```bash
# Validate your .env file
make validate-env

# Check for common issues:
# - Missing required variables
# - Incorrect format
# - Special characters not escaped
```

## 🚀 Next Steps

After setting up credentials:

1. **Test the setup**: `make test-db && make test-kafka`
2. **Run migrations**: `make migrate`
3. **Start development**: `make dev`
4. **Build REST API**: Continue with Task 6 in the implementation plan

## 📚 Detailed Documentation

For comprehensive setup instructions, see:
- [Cloud Setup Guide](docs/cloud-setup-guide.md) - Detailed Supabase and Confluent Cloud setup
- [Development Guide](docs/development.md) - Development workflow and best practices

## 🆘 Need Help?

- **Supabase Issues**: [docs.supabase.com](https://docs.supabase.com)
- **Confluent Cloud Issues**: [docs.confluent.io](https://docs.confluent.io)
- **Project Issues**: Create a GitHub issue with your error details