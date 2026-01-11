# Connect 4 Multiplayer System

A real-time multiplayer Connect 4 game system built with Go backend, React frontend, and Kafka-based analytics pipeline.

## Features

- **Real-time multiplayer gameplay** via WebSocket connections
- **Intelligent bot opponents** using minimax algorithm with alpha-beta pruning
- **Automatic matchmaking** with 10-second timeout fallback to bot games
- **Player reconnection support** with 30-second session persistence
- **Live leaderboard** with player statistics and rankings
- **Analytics pipeline** for game metrics and player behavior tracking

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

2. **Set up credentials** (Choose one option)
   
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

### Using Docker (Full Stack)

```bash
# Build and start all services
make docker-build
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## Development

### Available Commands

```bash
# Development
make dev              # Start with hot reload
make run-server       # Run server directly
make run-analytics    # Run analytics service

# Credential Setup
make setup-credentials # Interactive credential setup
make create-env       # Create .env from template
make validate-env     # Validate environment config
make test-db          # Test database connection
make test-kafka       # Test Kafka connection

# Building
make build           # Build all binaries
make build-prod      # Build for production

# Testing
make test            # Run all tests
make test-coverage   # Run tests with coverage
make test-property   # Run property-based tests

# Code Quality
make lint            # Run linter
make fmt             # Format code
make vet             # Run go vet

# Database
make migrate         # Run migrations
make migrate-up      # Run migrations (with binary)
make migrate-down    # Rollback migrations

# Documentation
make docs            # Generate API docs
make docs-serve      # Serve documentation

# Docker
make docker-build    # Build Docker images
make docker-up       # Start with Docker
make docker-down     # Stop Docker services
```

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

### Railway
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login and deploy
railway login
railway init
railway up
```

### Render
1. Connect your GitHub repository
2. Create a new Web Service
3. Set build command: `make build-prod`
4. Set start command: `./bin/server`

### Fly.io
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