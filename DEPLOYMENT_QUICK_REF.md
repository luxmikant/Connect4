# 🚀 Quick Deployment Reference

## For Render.com (Free Tier - No Shell Access)

### ✅ Migrations Solution
**Migrations run automatically!** The server now runs migrations on startup.

Check server logs for:
```
Running database migrations...
Migrations completed successfully
```

### Alternative Methods (if auto-migration fails):

#### Method 1: Connect from Local Machine
```bash
# Get External Database URL from Render Dashboard
# Database → Info → External Database URL

# Set environment variable
export DATABASE_URL="postgres://user:pass@host.render.com/dbname"
# Windows PowerShell:
$env:DATABASE_URL="postgres://user:pass@host.render.com/dbname"

# Run migrations
go run ./cmd/migrate -direction up
```

#### Method 2: Use Build Hook
Update `render.yaml`:
```yaml
buildCommand: go build -o server ./cmd/server && go run ./cmd/migrate -direction up
```

---

## Environment Variables Cheat Sheet

### Backend (Render)
```bash
# Required
PORT=8080
ENVIRONMENT=production
GIN_MODE=release
DATABASE_URL=postgresql://user:pass@host/dbname

# Kafka (from Confluent Cloud)
KAFKA_BOOTSTRAP_SERVERS=pkc-xxxxx.aws.confluent.cloud:9092
KAFKA_API_KEY=your_api_key
KAFKA_API_SECRET=your_api_secret
KAFKA_TOPIC=game-events

# CORS (update after frontend deploy)
CORS_ORIGINS=https://your-app.vercel.app,http://localhost:5173

# Optional
REDIS_URL=redis://localhost:6379
```

### Frontend (Vercel)
```bash
# Update these after backend is deployed
VITE_API_URL=https://connect4-server.onrender.com
VITE_WS_URL=wss://connect4-server.onrender.com/ws
```

---

## Quick Deploy Steps

### 1. Push to GitHub
```bash
git add .
git commit -m "Ready for deployment"
git push origin main
```

### 2. Deploy Backend (Render)
1. Go to https://dashboard.render.com
2. Click "New +" → "Web Service"
3. Connect GitHub repo
4. Settings:
   - Name: `connect4-server`
   - Build: `go build -o server ./cmd/server`
   - Start: `./server`
5. Add environment variables (see above)
6. Click "Create Web Service"
7. **Wait for "Migrations completed successfully" in logs**

### 3. Deploy Frontend (Vercel)
1. Go to https://vercel.com/dashboard
2. Click "New Project"
3. Import GitHub repo
4. Settings:
   - Root Directory: `web`
   - Build Command: `npm run build`
5. Add environment variables (see above)
6. Click "Deploy"

### 4. Update CORS
1. Copy Vercel URL (e.g., `https://connect4-xyz.vercel.app`)
2. Go back to Render → Your service → Environment
3. Update `CORS_ORIGINS`:
   ```
   https://connect4-xyz.vercel.app,http://localhost:5173
   ```
4. Save (auto-redeploys)

---

## Testing Deployment

```bash
# 1. Test backend health
curl https://connect4-server.onrender.com/health
# Expected: {"status":"ok"}

# 2. Test API
curl https://connect4-server.onrender.com/api/v1/leaderboard
# Expected: JSON array

# 3. Visit frontend
https://your-app.vercel.app
```

---

## Common Issues

### ❌ "Database connection failed"
- Check `DATABASE_URL` is set correctly
- Use **Internal Database URL** for Render services
- Use **External Database URL** for local connections

### ❌ "CORS error"
- Add your Vercel URL to `CORS_ORIGINS`
- Format: `https://app.vercel.app,http://localhost:5173`
- Redeploy after changing

### ❌ "WebSocket connection failed"
- Use `wss://` not `ws://` in production
- Check `VITE_WS_URL` environment variable
- Verify backend is running

### ❌ "Migrations failed"
- Check database URL is correct
- View full error in Render logs
- Try manual migration from local machine

---

## Monitoring

### Render
- **Logs**: Service → Logs tab
- **Metrics**: Service → Metrics tab
- **Events**: Service → Events tab

### Vercel
- **Deployments**: Project → Deployments
- **Logs**: Deployment → View Function Logs
- **Analytics**: Project → Analytics

---

## Rollback

### Render
1. Go to Service → Events
2. Click "Rollback" on previous deployment

### Vercel
1. Go to Project → Deployments
2. Click "..." on previous deployment
3. Click "Promote to Production"

---

## Cost Warnings

### Free Tier Limits
- **Render**: Services sleep after 15 min inactivity (cold start ~30s)
- **Vercel**: 100 GB bandwidth/month
- Keep analytics service disabled to save resources

### Upgrade When:
- Need 24/7 uptime (no cold starts)
- High traffic (>100k requests/month)
- Want shell access for debugging

---

## Support

- **Full Guide**: See [DEPLOYMENT.md](DEPLOYMENT.md)
- **Render Docs**: https://render.com/docs/web-services
- **Vercel Docs**: https://vercel.com/docs
- **Issues**: Create issue on GitHub

---

**Last Updated**: January 2026
