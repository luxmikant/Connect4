#!/bin/bash
# Quick deployment setup script for Connect4 Multiplayer

echo "🚀 Connect4 Multiplayer - Deployment Setup"
echo "=========================================="
echo ""

# Check if git is initialized
if [ ! -d .git ]; then
    echo "📦 Initializing Git repository..."
    git init
    git add .
    git commit -m "Initial commit: Connect4 Multiplayer Game"
fi

# Check if remote exists
if ! git remote | grep -q origin; then
    echo ""
    echo "⚠️  No GitHub remote found!"
    echo "Please create a GitHub repository and run:"
    echo "   git remote add origin https://github.com/YOUR_USERNAME/Connect4.git"
    echo "   git push -u origin main"
    echo ""
else
    echo "✅ Git repository configured"
fi

echo ""
echo "📋 Deployment Checklist:"
echo ""
echo "1️⃣  Backend (Render.com):"
echo "   - Sign up at https://render.com"
echo "   - Create PostgreSQL database (or use Supabase)"
echo "   - Create Web Service from GitHub repo"
echo "   - Set environment variables (see DEPLOYMENT.md)"
echo "   - Migrations will run automatically on first deploy"
echo ""
echo "2️⃣  Frontend (Vercel):"
echo "   - Sign up at https://vercel.com"
echo "   - Import GitHub repository"
echo "   - Set Root Directory: web"
echo "   - Set Build Command: npm run build"
echo "   - Add environment variables (VITE_API_URL, VITE_WS_URL)"
echo ""
echo "3️⃣  Environment Variables Needed:"
echo "   Backend:"
echo "   - DATABASE_URL (from Render DB or Supabase)"
echo "   - KAFKA_BOOTSTRAP_SERVERS"
echo "   - KAFKA_API_KEY"
echo "   - KAFKA_API_SECRET"
echo "   - CORS_ORIGINS (add your Vercel URL)"
echo ""
echo "   Frontend:"
echo "   - VITE_API_URL=https://your-app.onrender.com"
echo "   - VITE_WS_URL=wss://your-app.onrender.com/ws"
echo ""
echo "📖 Full guide: See DEPLOYMENT.md"
echo ""

# Check if connected to database
echo "🔍 Testing local build..."
if go build -o server ./cmd/server 2>/dev/null; then
    echo "✅ Server builds successfully"
    rm -f server
else
    echo "⚠️  Server build failed - check dependencies"
fi

echo ""
echo "🎯 Next Steps:"
echo "1. Push to GitHub: git push origin main"
echo "2. Deploy backend on Render.com"
echo "3. Deploy frontend on Vercel"
echo "4. Update CORS_ORIGINS with Vercel URL"
echo ""
echo "Good luck! 🍀"
