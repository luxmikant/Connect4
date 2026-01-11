# Technology Stack

## Architecture
Microservices architecture with three main components:
- **Go backend server** for game logic and WebSocket communication
- **React frontend** for user interface
- **Kafka-based analytics pipeline** for game metrics processing

## Backend Stack
- **Language**: Go
- **WebSocket Library**: [gorilla/websocket](https://github.com/gorilla/websocket)
- **Kafka Client**: [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go)
- **Database**: PostgreSQL for persistence
- **In-Memory Store**: Redis-like operations for active game sessions
- **Testing**: [Testify](https://github.com/stretchr/testify) with [gopter](https://github.com/leanovate/gopter) for property-based testing

## Frontend Stack
- **Framework**: React
- **WebSocket**: Native WebSocket API with automatic reconnection
- **State Management**: React hooks for game state
- **UI Components**: Custom Connect 4 game board and lobby components

## Infrastructure
- **Message Queue**: Kafka cluster for analytics event streaming
- **Database**: PostgreSQL for game data and player statistics
- **Development Environment**: Docker Compose for local setup

## Bot AI Implementation
- **Algorithm**: Minimax with alpha-beta pruning
- **Search Depth**: 7 levels (configurable)
- **Optimizations**: Transposition table, iterative deepening, move ordering
- **Response Time**: Sub-1-second move calculation

## AI Assistant Development Guidelines

### Required Libraries and Frameworks
When implementing this project, always use these specific libraries:

**Web Framework**: Gin (required)
```go
import "github.com/gin-gonic/gin"
```

**Database ORM**: GORM (required)
```go
import "gorm.io/gorm"
import "gorm.io/driver/postgres"
```

**Testing**: Testify + Gopter (required)
```go
import "github.com/stretchr/testify/assert"
import "github.com/leanovate/gopter"
```

**Validation**: Go Playground Validator (required)
```go
import "github.com/go-playground/validator/v10"
```

**Configuration**: Viper (required)
```go
import "github.com/spf13/viper"
```

**Documentation**: Swaggo (required)
```go
import "github.com/swaggo/gin-swagger"
```

### Code Generation Requirements
Always generate these when creating new endpoints:
1. **Swagger annotations** for all API endpoints
2. **GORM models** with proper tags and relationships
3. **Validation structs** for all request/response types
4. **Property-based tests** for core game logic
5. **Unit tests** with Testify for all services

### Performance Requirements
- **Database queries**: Use GORM preloading to avoid N+1 queries
- **WebSocket connections**: Implement connection pooling and cleanup
- **Bot AI**: Ensure minimax completes within 1 second
- **Memory management**: Clean up completed game sessions
- **Error handling**: Implement circuit breakers for external services

## Common Commands

### Development Setup
```bash
# Start all services locally
docker-compose up -d

# Run Go backend
go run cmd/server/main.go

# Run React frontend
npm start

# Run analytics service
go run cmd/analytics/main.go
```

### Testing
```bash
# Run all Go tests
go test ./...

# Run property-based tests specifically
go test -tags=property ./...

# Run frontend tests
npm test

# Run integration tests
go test -tags=integration ./...
```

### Build and Deploy
```bash
# Build Go binaries
go build -o bin/server cmd/server/main.go
go build -o bin/analytics cmd/analytics/main.go

# Build React production bundle
npm run build

# Run database migrations
go run cmd/migrate/main.go
```

### Code Quality
```bash
# Generate swagger documentation
swag init

# Run linter
golangci-lint run

# Format code
go fmt ./...
goimports -w .

# Generate mocks
go generate ./...
```

## AI Assistant Implementation Rules

### Always Include When Creating:
1. **Swagger annotations** on all HTTP handlers
2. **GORM tags** on all database models
3. **Validation tags** on all request structs
4. **Error handling** with proper HTTP status codes
5. **Logging** with structured fields
6. **Metrics** for monitoring (Prometheus format)
7. **Health checks** for all external dependencies

### Never Do:
- Skip input validation on API endpoints
- Use raw SQL instead of GORM
- Implement WebSocket without connection cleanup
- Create game logic without property-based tests
- Deploy without health check endpoints
- Use blocking operations in WebSocket handlers

### Code Style Requirements:
- Follow standard Go project layout (`cmd/`, `internal/`, `pkg/`)
- Use dependency injection for all services
- Implement repository pattern for data access
- Use interfaces for all external dependencies
- Add context.Context to all service methods