# Project Structure

## Repository Organization

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
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── hooks/        # Custom React hooks
│   │   └── services/     # API and WebSocket services
│   └── public/           # Static assets
├── migrations/            # Database schema migrations
├── docker-compose.yml     # Local development environment
├── .kiro/                # Kiro configuration
│   ├── specs/            # Feature specifications
│   └── steering/         # AI assistant guidance
└── docs/                 # Project documentation
```

## Code Organization Principles

### Go Backend Structure
- **cmd/**: Contains main application entry points, one per service
- **internal/**: Private packages not intended for external use
- **pkg/**: Public packages that could be imported by other projects
- **Domain-driven design**: Each internal package represents a business domain

### Frontend Structure
- **components/**: Reusable UI components (GameBoard, Leaderboard, etc.)
- **hooks/**: Custom React hooks for game state and WebSocket management
- **services/**: API clients and WebSocket connection logic

### Key Architectural Patterns
- **Microservices**: Separate services for game logic and analytics
- **Event-driven**: Kafka for decoupled analytics processing
- **Real-time communication**: WebSocket for game state synchronization
- **Clean architecture**: Clear separation between business logic and infrastructure

### Testing Organization
- **Unit tests**: Alongside source code in same package
- **Integration tests**: In `tests/` directory with `integration` build tag
- **Property tests**: Tagged with `property` build tag
- **Test data**: Shared fixtures in `testdata/` directories

### Configuration Management
- **Environment variables**: For service configuration
- **Docker Compose**: For local development dependencies
- **Migration scripts**: Versioned database schema changes
- **Kiro specs**: Feature requirements and design documentation

## AI Assistant File Creation Guidelines

### When Creating New Files, Always Follow These Patterns:

#### 1. Go Service Files (`internal/*/service.go`)
```go
package servicename

import (
    "context"
    "github.com/your-project/pkg/models"
)

type Service interface {
    // Define interface methods
}

type service struct {
    repo Repository
    // other dependencies
}

func NewService(repo Repository) Service {
    return &service{repo: repo}
}

// Implement methods with context.Context as first parameter
func (s *service) MethodName(ctx context.Context, params) error {
    // Implementation
}
```

#### 2. Repository Files (`internal/database/repositories/*.go`)
```go
package repositories

import (
    "context"
    "gorm.io/gorm"
    "github.com/your-project/pkg/models"
)

type PlayerRepository interface {
    Create(ctx context.Context, player *models.Player) error
    GetByID(ctx context.Context, id string) (*models.Player, error)
    // other methods
}

type playerRepository struct {
    db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) PlayerRepository {
    return &playerRepository{db: db}
}
```

#### 3. HTTP Handlers (`internal/api/handlers/*.go`)
```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

type GameHandler struct {
    gameService GameService
    validator   *validator.Validate
}

// @Summary Create new game
// @Description Create a new Connect 4 game session
// @Tags games
// @Accept json
// @Produce json
// @Param request body CreateGameRequest true "Game creation request"
// @Success 201 {object} models.GameSession
// @Failure 400 {object} ErrorResponse
// @Router /games [post]
func (h *GameHandler) CreateGame(c *gin.Context) {
    // Always include Swagger annotations
    // Always validate input
    // Always handle errors properly
    // Always return structured responses
}
```

#### 4. Model Files (`pkg/models/*.go`)
```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Player struct {
    ID        string    `json:"id" gorm:"primaryKey" validate:"required"`
    Username  string    `json:"username" gorm:"uniqueIndex" validate:"required,min=3,max=20"`
    CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
    
    // Always include JSON, GORM, and validation tags
    // Always include timestamps for audit trail
}

// Always include table name method for GORM
func (Player) TableName() string {
    return "players"
}
```

#### 5. Test Files (`*_test.go`)
```go
package servicename_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
    suite.Suite
    service Service
    mockRepo *MockRepository
}

func (suite *ServiceTestSuite) SetupTest() {
    // Setup before each test
}

func (suite *ServiceTestSuite) TestMethodName() {
    // Arrange
    // Act  
    // Assert
    assert.NoError(suite.T(), err)
}

func TestServiceTestSuite(t *testing.T) {
    suite.Run(t, new(ServiceTestSuite))
}
```

#### 6. Property-Based Test Files (`*_property_test.go`)
```go
//go:build property
// +build property

package servicename_test

import (
    "testing"
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
)

// Feature: connect-4-multiplayer, Property X: Description
func TestPropertyName(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("property description", prop.ForAll(
        func(input InputType) bool {
            // Property test logic
            return true
        },
        gen.SomeGenerator(),
    ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### File Naming Conventions
- **Services**: `service.go` (interface + implementation in same file)
- **Repositories**: `repository.go` or `{entity}_repository.go`
- **Handlers**: `{entity}_handler.go`
- **Models**: `{entity}.go`
- **Tests**: `{file}_test.go`
- **Property Tests**: `{file}_property_test.go`
- **Mocks**: `mock_{interface}.go` in `mocks/` subdirectory

### Directory Creation Rules
1. **Always create `internal/` packages** for application-specific code
2. **Use `pkg/` only for reusable libraries** that could be imported by other projects
3. **Group related functionality** in the same package (e.g., all game logic in `internal/game/`)
4. **Separate concerns** (handlers, services, repositories in different packages)
5. **Create `mocks/` subdirectories** for generated mock files

### Import Organization
Always organize imports in this order:
1. Standard library imports
2. Third-party imports  
3. Local project imports

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // Third-party
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    // Local
    "github.com/your-project/internal/database"
    "github.com/your-project/pkg/models"
)
```