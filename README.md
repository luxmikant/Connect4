# Connect 4 Multiplayer System

A real-time multiplayer Connect 4 game system built with Go backend, React frontend, and Kafka-based analytics pipeline.

## ✨ Key Highlights

🎮 **Real-Time Multiplayer** - Play live games via WebSocket with instant move synchronization  
🤖 **Smart AI Bot** - Challenge an intelligent bot using minimax algorithm with alpha-beta pruning  
⚡ **Auto Matchmaking** - 10-second queue with automatic bot fallback  
📊 **Kafka Analytics** - Real-time game metrics and player behavior tracking via Apache Kafka  
🏆 **Live Leaderboard** - Track rankings, win rates, and player statistics  
🔄 **Session Persistence** - Automatic reconnection with 30-second grace period  

## Features

- **Real-time multiplayer gameplay** via WebSocket connections
- **Intelligent bot opponents** using minimax algorithm with alpha-beta pruning
- **Automatic matchmaking** with 10-second timeout fallback to bot games
- **Player reconnection support** with 30-second session persistence
- **Live leaderboard** with player statistics and rankings
- **Kafka-powered analytics** - Tracks gameplay metrics, player behavior, game duration, and peak hours

## Architecture

- **Go Backend Server**: Game logic, WebSocket communication, REST API
- **React Frontend**: User interface for gameplay and leaderboard
- **PostgreSQL**: Game data and player statistics storage
- **Kafka**: Analytics event streaming
- **Redis**: Session management and caching

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Make (optional, for convenience commands)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd connect4-multiplayer
   ```

2. **Start the backend services**

   **Option A: Using Docker (Recommended)**
   ```bash
   # Start all services (PostgreSQL, Redis, Kafka, Backend)
   docker-compose up
   ```

   **Option B: Local Development**
   ```bash
   # 1. Start dependencies (PostgreSQL, Redis, Kafka)
   docker-compose up postgres redis kafka

   # 2. Run database migrations
   go run cmd/migrate/main.go

   # 3. Start the backend server
   go run cmd/server/main.go
   # OR with Make:
   make run-server

   # 4. (Optional) Start analytics service
   go run cmd/analytics/main.go
   ```

   The server will start on **http://localhost:8080** with WebSocket endpoint at **ws://localhost:8080/ws**

3. **Start the frontend**
   ```bash
   cd web
   npm install
   npm run dev
   ```
   Frontend will be available at **http://localhost:5173**

4. **Set up credentials** (Choose one option)
   
   **Option A: Interactive Setup (Recommended)**
   ```bash
   make setup-credentials
   ```
   
   **Option B: Manual Setup**
   ```bash
   # Create environment file
   make create-env
   
   # Edit .env with your Supabase and Confluent Cloud credentials
   # See CREDENTIALS_SETUP.md for detailed instructions
   ```
   
   **Option C: Local Development Only**
   ```bash
   # Use local services (no cloud setup required)
   make docker-up
   # Uses local PostgreSQL, Redis, and Kafka
   ```

3. **Validate your setup**
   ```bash
   # Test database and Kafka connections
   make test-db
   make test-kafka
   ```

4. **Install development tools**
   ```bash
   make setup
   ```

5. **Run database migrations**
   ```bash
   make migrate
   ```

6. **Start the development server**
   ```bash
   make dev
   ```

The server will start at `http://localhost:8080` with hot reload enabled.

> 📋 **Need help with credentials?** See [CREDENTIALS_SETUP.md](CREDENTIALS_SETUP.md) for a quick setup guide or [docs/cloud-setup-guide.md](docs/cloud-setup-guide.md) for detailed instructions.



# Analytics
make run-analytics   # Start Kafka analytics consumer
```

## 📊 Kafka Analytics

The system includes **real-time analytics** powered by Apache Kafka. Game events are published to Kafka and consumed by an analytics service that tracks:

- **Gameplay Metrics**: Average game duration, games per hour/day, min/max duration
- **Player Metrics**: Unique players per hour, active games, player win rates
- **Event Tracking**: All game events (started, moves, completed) stored for analysis

### Running Analytics Service

```bash
# Build analytics consumer
go build -o analytics.exe ./cmd/analytics

# Run analytics service
./analytics.exe
```

### Viewing Analytics Data

```sql
-- View latest analytics snapshot
SELECT 
    timestamp,
    games_completed_hour,
    games_completed_day,
    avg_game_duration_sec,
    unique_players_hour
FROM analytics_snapshots
ORDER BY timestamp DESC
LIMIT 10;

-- Most frequent winners
SELECT username, games_won, win_rate 
FROM player_stats 
ORDER BY games_won DESC 
LIMIT 10;
```

📖 **For detailed analytics guide, see [docs/kafka-analytics-guide.md](docs/kafka-analytics-guide.md)**


### Project Structure

```
/
├── cmd/                    # Application entry points
│   ├── server/            # Game server main
│   ├── analytics/         # Analytics service main
│   └── migrate/           # Database migration tool
├── internal/              # Private application code
│   ├── game/             # Game logic and engine
│   ├── websocket/        # WebSocket connection management
│   ├── bot/              # AI bot implementation
│   ├── matchmaking/      # Player matching service
│   ├── analytics/        # Analytics event processing
│   └── database/         # Database models and operations
├── pkg/                   # Public library code
│   ├── models/           # Shared data structures
│   └── utils/            # Common utilities
├── web/                   # Frontend React application
├── migrations/            # Database schema migrations
└── docs/                 # Project documentation
```

## API Documentation

Once the server is running, visit:
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Health Check: `http://localhost:8080/health`

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ENVIRONMENT` | Application environment | `development` |
| `SERVER_PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | Local PostgreSQL |
| `KAFKA_BOOTSTRAP_SERVERS` | Kafka brokers | `localhost:9092` |
| `KAFKA_API_KEY` | Confluent Cloud API key | - |
| `KAFKA_API_SECRET` | Confluent Cloud API secret | - |
| `REDIS_URL` | Redis connection string | `localhost:6379` |

### Cloud Services Setup

#### Supabase (Database)
1. Create a new Supabase project
2. Get your connection string from Settings > Database
3. Set `DATABASE_URL` environment variable

#### Confluent Cloud (Kafka)
1. Create a Confluent Cloud account
2. Create a new cluster
3. Generate API keys
4. Set `KAFKA_BOOTSTRAP_SERVERS`, `KAFKA_API_KEY`, and `KAFKA_API_SECRET`

## Deployment

🚀 **Deploy your game for FREE!** See our comprehensive deployment guides:

- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Complete step-by-step deployment guide
  - Deploy backend to **Render.com** (Free tier with WebSocket support)
  - Deploy frontend to **Vercel** (Free tier with auto-deploy)
  - **No shell access needed** - Migrations run automatically!
  - Environment variable templates included
  - Troubleshooting guide for common issues

- **[DEPLOYMENT_QUICK_REF.md](DEPLOYMENT_QUICK_REF.md)** - Quick reference card
  - Environment variables cheat sheet
  - One-page deployment steps
  - Common issues and solutions

### Recommended Free Stack
- **Backend**: Render.com (Go server + PostgreSQL)
- **Frontend**: Vercel (React/Vite)  
- **Database**: Supabase or Render PostgreSQL
- **Kafka**: Confluent Cloud (already configured)

### Alternative Options

#### Railway
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login and deploy
railway login
railway init
railway up
```

#### Fly.io
```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Deploy
fly launch
fly deploy
```

## Testing

### Unit Tests
```bash
make test
```

### Property-Based Tests
```bash
make test-property
```

### Integration Tests
```bash
make test-integration
```

### Coverage Report
```bash
make test-coverage
open coverage.html
```

## Project Narrative

For a detailed walkthrough of the most interesting technical decisions made while building this system—WebSocket architecture, minimax AI, Kafka analytics, and reconnection handling—see **[TECHNICAL_NARRATIVE.md](TECHNICAL_NARRATIVE.md)**.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

For questions and support, please open an issue on GitHub.