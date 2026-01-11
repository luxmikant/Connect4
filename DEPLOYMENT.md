# Connect4 Multiplayer - Deployment Guide

Complete guide to deploy your Connect4 game to production using **free hosting platforms**.

## 🏗️ Architecture Overview

```
┌─────────────────┐         ┌──────────────────┐
│   Vercel        │         │   Render.com     │
│   (Frontend)    │────────▶│   (Backend API)  │
│   React/Vite    │         │   Go Server      │
└─────────────────┘         └──────────────────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
            ┌───────▼─────┐  ┌──────▼──────┐  ┌─────▼──────┐
            │  Supabase   │  │ Confluent   │  │  Render    │
            │ PostgreSQL  │  │   Kafka     │  │ Analytics  │
            │  (Cloud)    │  │  (Cloud)    │  │  Service   │
            └─────────────┘  └─────────────┘  └────────────┘
```

## 📋 Prerequisites

Before deploying, ensure you have:

- [x] GitHub account
- [x] Render.com account (sign up at https://render.com)
- [x] Vercel account (sign up at https://vercel.com)
- [x] Supabase database (already configured)
- [x] Confluent Cloud Kafka (already configured)
- [x] Git repository pushed to GitHub

---

## Part 1: Deploy Backend to Render.com

### Step 1: Prepare Your Repository

1. **Create `render.yaml`** in project root:

```yaml
# render.yaml
services:
  # Main Game Server
  - type: web
    name: connect4-server
    env: go
    buildCommand: go build -o server ./cmd/server
    startCommand: ./server
    envVars:
      - key: PORT
        value: 8080
      - key: ENVIRONMENT
        value: production
      - key: DATABASE_URL
        fromDatabase:
          name: connect4-db
          property: connectionString
      - key: KAFKA_BOOTSTRAP_SERVERS
        sync: false
      - key: KAFKA_API_KEY
        sync: false
      - key: KAFKA_API_SECRET
        sync: false
    healthCheckPath: /health
    
  # Analytics Service
  - type: worker
    name: connect4-analytics
    env: go
    buildCommand: go build -o analytics ./cmd/analytics
    startCommand: ./analytics
    envVars:
      - key: DATABASE_URL
        fromDatabase:
          name: connect4-db
          property: connectionString
      - key: KAFKA_BOOTSTRAP_SERVERS
        sync: false
      - key: KAFKA_API_KEY
        sync: false
      - key: KAFKA_API_SECRET
        sync: false

databases:
  - name: connect4-db
    databaseName: connect4
    user: connect4_user
```

2. **Update `config/config.go`** to read from environment variables:

Already done! Your config reads from environment variables.

3. **Commit and push to GitHub**:

```bash
git add render.yaml
git commit -m "Add Render deployment configuration"
git push origin main
```

### Step 2: Create Render Services

#### 2.1 Create PostgreSQL Database

1. Go to https://dashboard.render.com
2. Click **"New +"** → **"PostgreSQL"**
3. Configure:
   - **Name**: `connect4-db`
   - **Database**: `connect4`
   - **User**: `connect4_user`
   - **Region**: Choose closest to you
   - **Plan**: **Free**
4. Click **"Create Database"**
5. **Copy the Internal Database URL** (starts with `postgresql://`)

#### 2.2 Deploy Backend Server

1. Click **"New +"** → **"Web Service"**
2. Connect your GitHub repository
3. Configure:
   - **Name**: `connect4-server`
   - **Region**: Same as database
   - **Branch**: `main`
   - **Root Directory**: Leave empty
   - **Environment**: `Go`
   - **Build Command**: `go build -o server ./cmd/server`
   - **Start Command**: `./server`
   - **Plan**: **Free**


4. **Add Environment Variables**:

Click **"Advanced"** → **"Add Environment Variable"**

```bash
# Server Config
PORT=8080
ENVIRONMENT=production
GIN_MODE=release

# Database (use Internal Database URL from step 2.1)
DATABASE_URL=postgresql://connect4_user:CuD3dhhVyDff5ijQQEMe1HJ9YVzDOs5X@dpg-d5dr4dhr0fns73aosvug-a/connect4_ojtw

# Kafka (from your Confluent Cloud)
KAFKA_BOOTSTRAP_SERVERS=pkc-9q8rv.ap-south-2.aws.confluent.cloud:9092
KAFKA_API_KEY=your_confluent_api_key
KAFKA_API_SECRET=your_confluent_api_secret
KAFKA_TOPIC=game-events

# CORS (add your Vercel domain later)
CORS_ORIGINS=https://your-app.vercel.app,http://localhost:5173

# Redis (optional - use Render Redis or skip)
REDIS_URL=redis://localhost:6379
```

5. Click **"Create Web Service"**

#### 2.3 Run Database Migrations

**⚠️ Note**: Shell access requires Render Premium plan. Use these **free tier alternatives**:

**Option 1: Automatic on Server Startup (Recommended)**

The server will auto-run migrations on startup. No action needed!

**Option 2: Use Build Hook in render.yaml**

Update your `render.yaml` build command to run migrations during build:
```yaml
buildCommand: go build -o server ./cmd/server && go run ./cmd/migrate -direction up
startCommand: ./server
```

This will:
1. Build the server executable
2. Run migrations automatically during build
3. Start the server after migrations complete

**Expected Log Output:**
```
$ go build -o server ./cmd/server && go run ./cmd/migrate -direction up
Running migrations...
Migrations completed successfully
$ ./server
2026/01/05 18:30:00 Running database migrations...
2026/01/05 18:30:02 Migrations completed successfully
2026/01/05 18:30:02 Server starting on port 8080
```

**Option 3: Connect from Local Machine**

If migrations don't run automatically, connect from your local machine:

```bash
# 1. Get External Database URL from Render Dashboard
#    Go to: Database → Info → External Database URL
#    Example: postgres://user:pass@oregon-postgres.render.com/dbname

# 2. Set environment variable locally
export DATABASE_URL="postgres://your-external-url-here"

# 3. Run migrations from your machine
go run ./cmd/migrate -direction up
```

**Verify Migrations**:
- Check server logs for "Migrations completed successfully"
- Or use database query tool to verify tables exist

#### 2.4 Deploy Analytics Service (Optional)

1. Click **"New +"** → **"Background Worker"**
2. Connect your GitHub repository
3. Configure:
   - **Name**: `connect4-analytics`
   - **Environment**: `Go`
   - **Build Command**: `go build -o analytics ./cmd/analytics`
   - **Start Command**: `./analytics`
   - **Plan**: **Free**

4. Add same environment variables as server

---

## Part 2: Deploy Frontend to Vercel

### Step 1: Prepare Frontend

1. **Update `web/.env.production`**:

```env
VITE_API_URL=https://connect4-server.onrender.com
VITE_WS_URL=wss://connect4-server.onrender.com/ws
```

2. **Update `web/src/services/websocket.ts`**:

Make sure WebSocket connects to production URL:

```typescript
const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';
```

3. **Create `vercel.json`** in `web/` directory:

⚠️ **IMPORTANT**: Replace `https://connect4-server.onrender.com` with your **actual Render URL** (e.g., `https://connect4-server-qgi9.onrender.com`)

```json
{
  "rewrites": [
    {
      "source": "/api/:path*",
      "destination": "https://connect4-server-qgi9.onrender.com/api/:path*"
    }
  ],
  "headers": [
    {
      "source": "/assets/(.*)",
      "headers": [
        {
          "key": "Cache-Control",
          "value": "public, max-age=31536000, immutable"
        }
      ]
    }
  ]
}
```

**How to get your Render URL:**
- Go to https://dashboard.render.com
- Click `connect4-server` service
- Copy the URL shown at the top of the page
- Use that URL in the `destination` field above

4. **Commit changes**:

```bash
cd web
git add .env.production vercel.json
git commit -m "Add Vercel deployment config"
git push origin main
```

### Step 2: Deploy to Vercel

1. Go to https://vercel.com/dashboard
2. Click **"Add New Project"**
3. Import your GitHub repository
4. Configure:
   - **Framework Preset**: Vite
   - **Root Directory**: `web`
   - **Build Command**: `npm run build`
   - **Output Directory**: `dist`

5. **Environment Variables**:

```bash
VITE_API_URL=https://connect4-server.onrender.com
VITE_WS_URL=wss://connect4-server.onrender.com/ws
```

6. Click **"Deploy"**

7. **Copy your Vercel URL** (e.g., `https://connect4-game.vercel.app`)

### Step 3: Update CORS

Go back to Render → Your backend service → Environment:

Update `CORS_ORIGINS`:
```
https://connect4-game.vercel.app,http://localhost:5173
```

Click **"Save Changes"** (this will redeploy)

---

## Part 3: Post-Deployment Configuration

### Update Backend Config

Edit `config.prod.yaml`:

```yaml
environment: production

server:
  port: ${PORT}
  host: "0.0.0.0"
  cors_origins:
    - ${CORS_ORIGINS}
  read_timeout: 30
  write_timeout: 30

database:
  url: ${DATABASE_URL}
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300
  ssl_mode: "require"

kafka:
  bootstrap_servers: ${KAFKA_BOOTSTRAP_SERVERS}
  api_key: ${KAFKA_API_KEY}
  api_secret: ${KAFKA_API_SECRET}
  topic: ${KAFKA_TOPIC}
  consumer_group: "analytics-service"

redis:
  url: ${REDIS_URL}
  password: ""
  db: 0
```

### Test Your Deployment

1. **Backend Health Check**:
```bash
curl https://connect4-server.onrender.com/health
```

Expected response: `{"status":"ok"}`

2. **Frontend**:
Visit your Vercel URL: `https://connect4-game.vercel.app`

3. **Play a Game**:
- Enter username
- Join game
- Test bot or multiplayer

4. **Check Leaderboard**:
Visit `/leaderboard`

---

## 🔧 Troubleshooting

### Backend Issues

**Problem**: Server won't start
```bash
# Check logs in Render Dashboard
# Verify environment variables are set
# Check database connection
```

**Problem**: WebSocket connection fails
- Ensure Render service is using `wss://` protocol
- Check CORS settings include Vercel domain
- Verify PORT environment variable is set

### Frontend Issues

**Problem**: API calls fail
- Check `VITE_API_URL` is correct
- Verify CORS is configured on backend
- Check browser console for errors

**Problem**: Build fails
```bash
# Locally test production build
cd web
npm run build
npm run preview
```

### Database Issues

**Problem**: Migrations not applied
```bash
# SSH into Render service
# Run migrations manually
./server migrate
```

---

## 📊 Monitoring

### Render Dashboard
- View logs: Service → Logs
- Check metrics: Service → Metrics
- Database stats: Database → Metrics

### Vercel Dashboard
- Deployment logs: Project → Deployments
- Analytics: Project → Analytics
- Runtime logs: Project → Logs

---

## 🚀 Continuous Deployment

Both Render and Vercel auto-deploy when you push to GitHub:

```bash
# Make changes
git add .
git commit -m "Update feature"
git push origin main

# Automatic deployments:
# - Vercel deploys frontend
# - Render deploys backend
```

---

## 💰 Cost Optimization

### Free Tier Limits

**Render (Free)**:
- 750 hours/month per service
- Services sleep after 15 min inactivity
- 100 GB bandwidth/month
- PostgreSQL: 1 GB storage

**Vercel (Free)**:
- 100 GB bandwidth/month
- 100 deployments/day
- Unlimited preview deployments

**Tips**:
1. Keep services active with health checks
2. Use Vercel for static assets
3. Optimize images and bundle size
4. Cache API responses

---

## 🔒 Security Checklist

- [x] Environment variables (not in code)
- [x] HTTPS/WSS only in production
- [x] CORS properly configured
- [x] Database SSL enabled
- [x] API rate limiting (add if needed)
- [x] Secure Kafka credentials
- [x] Input validation enabled

---

## 📝 Custom Domain (Optional)

### Add Custom Domain to Vercel
1. Go to Project Settings → Domains
2. Add your domain
3. Update DNS records as instructed

### Add Custom Domain to Render
1. Go to Service Settings
2. Click "Custom Domain"
3. Add your domain
4. Update DNS records

---

## Alternative: Railway.app (All-in-One)

If you prefer single platform:

1. **Railway.app** can host:
   - Go backend
   - PostgreSQL database
   - Analytics service
   - Redis (optional)

2. **Advantages**:
   - One platform for everything
   - $5 free credit/month
   - Easy service linking
   - Good Go support

3. **Deploy**:
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Initialize
railway init

# Deploy
railway up
```

---

## 🎯 Quick Start Commands

```bash
# Backend (Render)
git push origin main  # Auto-deploys

# Frontend (Vercel)  
git push origin main  # Auto-deploys

# Manual redeploy
# Render: Dashboard → Manual Deploy
# Vercel: Dashboard → Redeploy

# View logs
# Render: Service → Logs
# Vercel: Project → Functions → View Logs
```

---

## 📞 Support Resources

- **Render Docs**: https://render.com/docs
- **Vercel Docs**: https://vercel.com/docs
- **Render Discord**: https://discord.gg/render
- **Vercel Discord**: https://discord.gg/vercel

---

## ✅ Deployment Checklist

Before going live:

- [ ] GitHub repository created and pushed
- [ ] Render account created
- [ ] Vercel account created
- [ ] Database credentials obtained (Supabase or Render)
- [ ] Kafka credentials ready (Confluent Cloud)
- [ ] Environment variables configured on Render
- [ ] Backend deployed successfully
- [ ] **Migrations run automatically** (check server logs)
- [ ] Frontend deployed to Vercel
- [ ] CORS settings updated with Vercel URL
- [ ] Frontend API URLs point to Render backend
- [ ] SSL/TLS enabled (automatic on both platforms)
- [ ] Health check passing: `curl https://your-app.onrender.com/health`
- [ ] Test WebSocket connections from frontend
- [ ] Test game functionality (bot + multiplayer)
- [ ] Test matchmaking queue
- [ ] Test leaderboard display
- [ ] Analytics service running (optional)
- [ ] Monitor logs for errors

---

## 🎉 Success!

Your Connect4 game is now live! Share your Vercel URL with friends and start playing!

**Common Issues?** See Troubleshooting section above or check:
- Server logs on Render
- Browser console for frontend errors
- Database connection string is correct
- CORS origins include your Vercel domain

**Need Help?** 
- Check [DEPLOYMENT.md](DEPLOYMENT.md) for detailed steps
- Review server logs for migration status
- Test locally first: `./server.exe` and `cd web && npm run dev`
