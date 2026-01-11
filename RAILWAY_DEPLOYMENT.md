# Railway Deployment Guide

Complete guide to deploy Connect4 multiplayer game on Railway.

## 🚀 Why Railway?

- **Better Free Tier**: $5 free credits monthly (vs Render's limited resources)
- **Faster Builds**: Railway has faster build times
- **Easy Database**: One-click PostgreSQL provisioning
- **Simpler Setup**: Automatic environment variable sharing
- **Better WebSockets**: Excellent WebSocket support

---

## 📋 Prerequisites

1. **Railway Account**: Sign up at [railway.app](https://railway.app)
2. **GitHub Account**: Your code repository
3. **Confluent Cloud Account**: For Kafka (free tier)
4. **Domain** (Optional): For custom domain

---

## 🔧 Part 1: Backend Deployment (Railway)

### Step 1: Create New Project

1. Go to [Railway Dashboard](https://railway.app/dashboard)
2. Click **New Project**
3. Select **Deploy from GitHub repo**
4. Connect your GitHub account if not already connected
5. Select your `Connect4` repository
6. Railway will detect it's a Go project

### Step 2: Add PostgreSQL Database

1. In your Railway project dashboard
2. Click **New** → **Database** → **PostgreSQL**
3. Railway automatically creates the database
4. Database URL is automatically available as `DATABASE_URL`

### Step 3: Configure Environment Variables

In Railway project → **Variables** tab, add:

```bash
# Server Configuration
PORT=8080
GIN_MODE=release

# Database (automatically provided by Railway)
# DATABASE_URL is auto-populated when you add PostgreSQL

# JWT Secret
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# CORS Origins (add your frontend URL after deployment)
CORS_ORIGINS=https://your-frontend.vercel.app,http://localhost:5173

# Kafka Configuration (from Confluent Cloud)
KAFKA_BOOTSTRAP_SERVERS=pkc-xxxxx.us-east-1.aws.confluent.cloud:9092
KAFKA_SASL_USERNAME=your-api-key
KAFKA_SASL_PASSWORD=your-api-secret
KAFKA_TOPIC=game-events
KAFKA_SECURITY_PROTOCOL=SASL_SSL
KAFKA_SASL_MECHANISM=PLAIN

# Analytics Service
ANALYTICS_ENABLED=true
```

### Step 4: Configure Build Settings

Railway auto-detects Go projects, but you can customize:

1. Go to **Settings** → **Build**
2. **Build Command** (optional):
   ```bash
   go build -o server ./cmd/server
   ```
3. **Start Command**:
   ```bash
   ./server
   ```

### Step 5: Add Railway Configuration File

Create `railway.json` in your project root:

```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "NIXPACKS"
  },
  "deploy": {
    "startCommand": "./server",
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

### Step 6: Database Migrations

Your app already auto-runs migrations on startup (see `cmd/server/main.go`).

Railway will:
1. Build your Go server
2. Start the server
3. Migrations run automatically on first connection
4. Server starts listening on port 8080

### Step 7: Deploy Backend

1. Commit and push to GitHub:
   ```bash
   git add railway.json
   git commit -m "Add Railway configuration"
   git push origin main
   ```

2. Railway automatically deploys on push
3. Wait for deployment to complete (~2-3 minutes)
4. Click **Settings** → **Generate Domain** to get your public URL

**Your backend URL**: `https://your-app.up.railway.app`

---

## 🎨 Part 2: Frontend Deployment

You have two options:

### Option A: Deploy Frontend on Vercel (Recommended)

Keep your current Vercel setup and update environment variables:

1. Go to Vercel Dashboard → Your Project → **Settings** → **Environment Variables**
2. Update `VITE_API_URL`:
   ```
   VITE_API_URL=https://your-app.up.railway.app
   ```
3. Update `VITE_WS_URL`:
   ```
   VITE_WS_URL=wss://your-app.up.railway.app
   ```
4. Redeploy: **Deployments** → **...** → **Redeploy**

### Option B: Deploy Frontend on Railway

1. In Railway Dashboard → **New** → **GitHub Repo**
2. Select your repository (same repo, different service)
3. Add environment variables:
   ```bash
   VITE_API_URL=https://your-backend.up.railway.app
   VITE_WS_URL=wss://your-backend.up.railway.app
   ```
4. Configure build settings:
   - **Root Directory**: `web`
   - **Build Command**: `npm install && npm run build`
   - **Start Command**: `npm run preview`
   - **Install Command**: `npm install`

5. Generate domain for frontend
6. Both services run in the same Railway project

---

## 🔐 Part 3: Database Setup

### Verify Database Connection

Railway automatically provides `DATABASE_URL` in the format:
```
postgresql://user:password@host:port/database
```

Your Go code already uses this! (See `internal/config/config.go`)

### Access Database (Optional)

1. In Railway → PostgreSQL service → **Data** tab
2. Use the built-in database browser
3. Or use external tools with connection string from **Connect** tab

### Backup Database

1. Railway PostgreSQL includes automatic backups (paid plan)
2. For free tier, use Railway CLI:
   ```bash
   railway run pg_dump $DATABASE_URL > backup.sql
   ```

---

## ⚙️ Part 4: Kafka Configuration

### Using Confluent Cloud (Recommended)

1. Your Kafka cluster is already set up on Confluent Cloud
2. Just add the environment variables to Railway (see Step 3 above)
3. Railway will use these for the analytics producer

### Verify Kafka Connection

Check Railway logs:
```
✅ Kafka Analytics Producer initialized successfully
```

---

## 🔄 Part 5: Update Backend CORS

After deploying frontend, update CORS in Railway:

1. Railway → Backend Service → **Variables**
2. Update `CORS_ORIGINS`:
   ```bash
   CORS_ORIGINS=https://your-frontend.vercel.app,https://your-frontend.up.railway.app,http://localhost:5173
   ```
3. Service will auto-redeploy

---

## 🧪 Part 6: Testing Deployment

### Test Backend Health

```bash
curl https://your-app.up.railway.app/health
```

Expected response:
```json
{
  "service": "connect4-multiplayer",
  "status": "healthy",
  "version": "1.0.0"
}
```

### Test API Endpoints

```bash
# Get leaderboard
curl https://your-app.up.railway.app/api/v1/leaderboard

# Create player (should return 201)
curl -X POST https://your-app.up.railway.app/api/v1/players \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser"}'
```

### Test WebSocket

1. Open browser DevTools → Console
2. Run:
   ```javascript
   const ws = new WebSocket('wss://your-app.up.railway.app/ws');
   ws.onopen = () => console.log('Connected!');
   ws.onmessage = (e) => console.log('Message:', e.data);
   ```

### Test Frontend

1. Visit your frontend URL
2. Click **Play with Bot**
3. Check browser console for:
   - `WebSocket: Connected to wss://your-app.up.railway.app/ws`
   - No CORS errors
4. Make a move → should see game updates
5. Check **Leaderboard** loads

---

## 📊 Part 7: Monitoring & Logs

### View Logs

1. Railway Dashboard → Your Service → **Logs** tab
2. Real-time log streaming
3. Filter by severity: Info, Warning, Error

### Monitor Resources

1. **Metrics** tab shows:
   - CPU usage
   - Memory usage
   - Network traffic
   - Request count

### Set Up Alerts (Paid Plans)

- Configure alerts for downtime
- Get notified of deployment failures
- Monitor resource usage

---

## 🚀 Part 8: Custom Domain (Optional)

### Add Custom Domain

1. Railway → Service → **Settings** → **Domains**
2. Click **Custom Domain**
3. Enter your domain: `connect4.yourdomain.com`
4. Add DNS records to your domain provider:
   ```
   Type: CNAME
   Name: connect4
   Value: your-app.up.railway.app
   ```
5. Wait for DNS propagation (~10 minutes)
6. Railway auto-provisions SSL certificate

### Update Environment Variables

After adding custom domain, update:
- Backend `CORS_ORIGINS`: add your custom domain
- Frontend `VITE_API_URL` and `VITE_WS_URL`: use custom domain

---

## 💰 Railway Pricing & Free Tier

### Free Tier Includes:
- **$5 credit per month**
- ~500 hours of usage
- Unlimited projects
- 512 MB RAM per service
- 1 GB disk space
- Shared CPU

### Usage Estimates (Connect4):
- **Backend**: ~$3/month (always running)
- **Database**: ~$1-2/month
- **Frontend** (if on Railway): ~$1/month

**Total**: Easily fits in $5 free credit for development/testing

### Cost Optimization Tips:
1. Use Vercel for frontend (free)
2. Keep only backend + database on Railway
3. Monitor usage in **Usage** tab
4. Pause unused services

---

## 🔧 Troubleshooting

### Build Fails

**Issue**: Go build errors
**Solution**:
1. Check Railway logs for specific error
2. Ensure `go.mod` and `go.sum` are committed
3. Verify Go version (1.21+)

### Database Connection Fails

**Issue**: Cannot connect to PostgreSQL
**Solution**:
1. Verify `DATABASE_URL` is set (Railway auto-sets this)
2. Check PostgreSQL service is running
3. Restart backend service

### WebSocket Connection Fails

**Issue**: WebSocket handshake error
**Solution**:
1. Ensure Railway service is running (not sleeping)
2. Check CORS settings include frontend domain
3. Verify `VITE_WS_URL` uses `wss://` (not `ws://`)

### Migrations Don't Run

**Issue**: Tables not created
**Solution**:
1. Check logs: `Migrations completed successfully`
2. Migrations auto-run on startup (see `cmd/server/main.go`)
3. Manually run: `railway run go run ./cmd/migrate -direction up`

### Frontend Shows 404/CORS Errors

**Issue**: API calls fail from frontend
**Solution**:
1. Verify `VITE_API_URL` in Vercel matches Railway backend URL
2. Check `CORS_ORIGINS` in Railway includes frontend domain
3. Clear Vercel deployment cache and redeploy

### Railway Service Sleeping

**Issue**: First request takes ~30 seconds
**Solution**:
- Railway free tier services don't sleep (unlike Render)
- If you need guaranteed uptime, upgrade to Hobby plan ($5/month)

---

## 🔄 CI/CD with Railway

### Automatic Deployments

Railway automatically deploys when you push to GitHub:

```bash
# Make changes
git add .
git commit -m "Update feature"
git push origin main

# Railway auto-deploys in ~2 minutes
```

### Manual Deployments

1. Railway Dashboard → Service → **Deployments**
2. Click **Redeploy** on any previous deployment
3. Or use Railway CLI:
   ```bash
   # Install Railway CLI
   npm i -g @railway/cli

   # Login
   railway login

   # Deploy
   railway up
   ```

### Deployment Rollback

1. **Deployments** tab → Previous deployment → **...** → **Redeploy**
2. Instant rollback to previous version

---

## 📝 Quick Reference Commands

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login to Railway
railway login

# Link local project to Railway
railway link

# View logs
railway logs

# Run migrations manually
railway run go run ./cmd/migrate -direction up

# Open Railway dashboard
railway open

# Get database connection string
railway variables

# Run command in Railway environment
railway run <command>

# Deploy current code
railway up
```

---

## 🎯 Deployment Checklist

### Before Deployment:
- [ ] GitHub repository is up to date
- [ ] `railway.json` configuration added
- [ ] Migrations work locally
- [ ] Environment variables prepared
- [ ] Confluent Cloud Kafka cluster ready

### After Backend Deployment:
- [ ] Health endpoint responds: `/health`
- [ ] Database migrations completed (check logs)
- [ ] Kafka connection successful (check logs)
- [ ] Generated Railway domain works
- [ ] API endpoints return data: `/api/v1/leaderboard`

### After Frontend Deployment:
- [ ] Frontend loads without errors
- [ ] WebSocket connects successfully
- [ ] Bot game works
- [ ] Matchmaking works (if testing with 2 tabs)
- [ ] Leaderboard displays
- [ ] No CORS errors in console

---

## 🆚 Railway vs Render Comparison

| Feature | Railway | Render |
|---------|---------|--------|
| Free Tier | $5 credit/month | 750 hours/month |
| Spin-up Time | Always on | ~30s cold start |
| Build Speed | Fast | Slower |
| Database | 1-click setup | Manual setup |
| SSL | Auto | Auto |
| Custom Domain | ✅ Free | ✅ Free |
| Logs Retention | 7 days | Limited |
| Best For | Full-stack apps | Static + API |

**Recommendation**: Railway is better for this Connect4 app because:
- No cold starts (important for WebSockets)
- Faster builds
- Better free tier for full-stack apps
- Easier database setup

---

## 🔗 Useful Links

- **Railway Dashboard**: https://railway.app/dashboard
- **Railway Docs**: https://docs.railway.app
- **Railway CLI**: https://docs.railway.app/develop/cli
- **Railway Discord**: https://discord.gg/railway
- **Status Page**: https://status.railway.app

---

## 🎓 Next Steps

After successful deployment:

1. **Custom Domain**: Add professional domain
2. **Monitoring**: Set up uptime monitoring (UptimeRobot)
3. **Analytics**: Enable Google Analytics in frontend
4. **Scaling**: Monitor usage and upgrade if needed
5. **Backups**: Schedule regular database backups

---

## 💡 Pro Tips

1. **Use Railway CLI** for faster debugging and log access
2. **Enable Railway's Postgres backups** (paid feature) for production
3. **Set up GitHub webhooks** for deployment notifications
4. **Use Railway environments** for staging/production separation
5. **Monitor your $5 credit usage** to avoid service interruption
6. **Keep Kafka on Confluent Cloud** - Railway doesn't offer managed Kafka
7. **Use Railway variables** for secrets - they're encrypted at rest

---

## 📧 Support

If you encounter issues:
1. Check Railway logs first
2. Review this guide's troubleshooting section
3. Railway Discord community is very responsive
4. Railway documentation is excellent

---

**Happy Deploying! 🚀**

Your Connect4 game will be live on Railway with better performance and easier management than Render!
