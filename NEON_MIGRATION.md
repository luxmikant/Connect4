# Migrating from Render Postgres to Neon

## Why This Migration?

Render's free PostgreSQL tier expires after 90 days, causing the entire backend to crash because the server was tightly coupled to DB availability at startup. This guide migrates you to **Neon** (always-free serverless Postgres) and documents the resilience improvements made to prevent this from happening again.

---

## Quick Steps to Get Back Online

### 1. Create a Neon Database (2 minutes)

1. Go to [neon.tech](https://neon.tech) and sign up (GitHub SSO works)
2. Create a new project named `connect4`
3. Select region closest to your Render service (e.g., `US East (Ohio)` for `us-east`)
4. Copy the connection string — it looks like:
   ```
   postgresql://neondb_owner:password@ep-cool-name-12345.us-east-2.aws.neon.tech/neondb?sslmode=require
   ```

### 2. Update Render Environment Variable (1 minute)

1. Go to [Render Dashboard](https://dashboard.render.com) → your `connect4-server` service
2. Click **Environment** tab
3. Set `DATABASE_URL` to your Neon connection string
4. Do the same for `connect4-analytics` worker if deployed
5. Click **Save Changes** — Render will auto-redeploy

### 3. Verify

```bash
# Check server health
curl https://your-render-url.onrender.com/health

# Expected response:
# {"status":"healthy","service":"connect4-multiplayer","version":"1.0.0"}

# Check DB readiness
curl https://your-render-url.onrender.com/health/ready

# Expected response:
# {"status":"ready"}
```

---

## What Changed in the Codebase

### Problem: Tight Coupling

```
Before: DB down → config.Load() fails OR database.Initialize() fails → log.Fatalf → SERVER DIES
```

### Solution: Graceful Degradation

```
After:  DB down → server starts in degraded mode → retries in background → full service restored on reconnect
```

### Files Modified

| File | Change |
|------|--------|
| `internal/database/resilient.go` | **NEW** — Resilient DB wrapper with retry logic + exponential backoff |
| `cmd/server/main.go` | Server starts even without DB; services init lazily after reconnection |
| `internal/config/config.go` | DB URL and Kafka are no longer hard requirements; warnings instead of fatal errors |
| `config.prod.yaml` | Connection pool tuned for Neon's serverless limits (20 max connections) |
| `render.yaml` | Removed Render managed database; added Neon setup instructions |

### New Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Always returns `200` — keeps Render from killing the service |
| `GET /health/ready` | Returns `200` only when DB is connected; `503` otherwise |

### Resilience Features

- **Exponential backoff retry**: 5s → 10s → 20s → ... → max 2 min between retries
- **Auto-migration on reconnect**: Migrations run automatically when DB becomes available
- **Degraded mode**: Health check returns `200` with `"status": "degraded"` so Render doesn't restart
- **Connection health monitoring**: Active ping checks detect DB disconnections and trigger reconnection

---

## Neon vs Render Postgres Comparison

| Feature | Render Free DB | Neon Free Tier |
|---------|---------------|----------------|
| **Duration** | 90 days then deleted | Always free |
| **Storage** | 1 GB | 0.5 GB |
| **Compute** | Shared | Serverless (scales to zero) |
| **Cold start** | None | ~500ms after idle |
| **Branching** | No | Yes (great for dev/staging) |
| **Connection limit** | ~20 | 100 |
| **SSL** | Optional | Required |

---

## Neon-Specific Tips

### Connection Pooling
Neon uses a built-in connection pooler. Append `-pooler` to your endpoint hostname for pooled connections:
```
# Direct (for migrations)
postgresql://user:pass@ep-xxx.us-east-2.aws.neon.tech/connect4?sslmode=require

# Pooled (for application — recommended)
postgresql://user:pass@ep-xxx-pooler.us-east-2.aws.neon.tech/connect4?sslmode=require
```

### Cold Starts
Neon scales to zero after 5 minutes of inactivity on the free tier. The first query after idle takes ~500ms. The resilient DB layer handles this gracefully with retries.

### Branch Databases
Create branch databases for development without affecting production:
```bash
# In Neon dashboard: Create Branch → name it "dev" or "staging"
# Each branch gets its own connection string
```

---

## If You Need to Restore Data

If you had data in the old Render database and didn't export it before expiry, it's unfortunately gone. For future protection:

1. **Set up `pg_dump` cron**: Schedule regular backups to cloud storage
2. **Use Neon's point-in-time recovery**: Free tier includes 7 days of history
3. **Schema is auto-migrated**: The server auto-runs migrations on connect, so schema will be recreated

---

## Environment Variables Reference

```bash
# Required
DATABASE_URL=postgresql://user:pass@ep-xxx.us-east-2.aws.neon.tech/connect4?sslmode=require

# Optional (already configured)
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_SERVICE_KEY=eyJ...
KAFKA_BOOTSTRAP_SERVERS=...
KAFKA_API_KEY=...
KAFKA_API_SECRET=...
```
