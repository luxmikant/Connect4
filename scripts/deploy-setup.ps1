# Quick deployment setup script for Connect4 Multiplayer
# PowerShell version for Windows

Write-Host "🚀 Connect4 Multiplayer - Deployment Setup" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Check if git is initialized
if (-not (Test-Path .git)) {
    Write-Host "📦 Initializing Git repository..." -ForegroundColor Yellow
    git init
    git add .
    git commit -m "Initial commit: Connect4 Multiplayer Game"
}

# Check if remote exists
$hasRemote = git remote | Select-String -Pattern "origin" -Quiet
if (-not $hasRemote) {
    Write-Host ""
    Write-Host "⚠️  No GitHub remote found!" -ForegroundColor Yellow
    Write-Host "Please create a GitHub repository and run:"
    Write-Host "   git remote add origin https://github.com/YOUR_USERNAME/Connect4.git" -ForegroundColor Green
    Write-Host "   git push -u origin main" -ForegroundColor Green
    Write-Host ""
} else {
    Write-Host "✅ Git repository configured" -ForegroundColor Green
}

Write-Host ""
Write-Host "📋 Deployment Checklist:" -ForegroundColor Cyan
Write-Host ""
Write-Host "1️⃣  Backend (Render.com):" -ForegroundColor Yellow
Write-Host "   - Sign up at https://render.com"
Write-Host "   - Create PostgreSQL database (or use Supabase)"
Write-Host "   - Create Web Service from GitHub repo"
Write-Host "   - Set environment variables (see DEPLOYMENT.md)"
Write-Host "   - Migrations will run automatically on first deploy" -ForegroundColor Green
Write-Host ""
Write-Host "2️⃣  Frontend (Vercel):" -ForegroundColor Yellow
Write-Host "   - Sign up at https://vercel.com"
Write-Host "   - Import GitHub repository"
Write-Host "   - Set Root Directory: web"
Write-Host "   - Set Build Command: npm run build"
Write-Host "   - Add environment variables (VITE_API_URL, VITE_WS_URL)"
Write-Host ""
Write-Host "3️⃣  Environment Variables Needed:" -ForegroundColor Yellow
Write-Host "   Backend:"
Write-Host "   - DATABASE_URL (from Render DB or Supabase)"
Write-Host "   - KAFKA_BOOTSTRAP_SERVERS"
Write-Host "   - KAFKA_API_KEY"
Write-Host "   - KAFKA_API_SECRET"
Write-Host "   - CORS_ORIGINS (add your Vercel URL)"
Write-Host ""
Write-Host "   Frontend:"
Write-Host "   - VITE_API_URL=https://your-app.onrender.com"
Write-Host "   - VITE_WS_URL=wss://your-app.onrender.com/ws"
Write-Host ""
Write-Host "📖 Full guide: See DEPLOYMENT.md" -ForegroundColor Cyan
Write-Host ""

# Check if connected to database
Write-Host "🔍 Testing local build..." -ForegroundColor Yellow
$buildResult = go build -o server.exe ./cmd/server 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Server builds successfully" -ForegroundColor Green
    Remove-Item server.exe -ErrorAction SilentlyContinue
} else {
    Write-Host "⚠️  Server build failed - check dependencies" -ForegroundColor Red
}

Write-Host ""
Write-Host "🎯 Next Steps:" -ForegroundColor Cyan
Write-Host "1. Push to GitHub: " -NoNewline
Write-Host "git push origin main" -ForegroundColor Green
Write-Host "2. Deploy backend on Render.com"
Write-Host "3. Deploy frontend on Vercel"
Write-Host "4. Update CORS_ORIGINS with Vercel URL"
Write-Host ""
Write-Host "Good luck! 🍀" -ForegroundColor Green
